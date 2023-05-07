package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/internal/utils"
	"github.com/thushjandan/pifina/pkg/controller"
)

func main() {
	logLevel := flag.String("level", "info", "set the log level. The default is info. Possible options: trace, debug, info, warn, error, off")
	bfrt_endpoint := flag.String("bfrt", "127.0.0.1:50052", "BF runtime GRPC server address (Dataplane endpoint)")
	p4_name := flag.String("p4name", "", "Name of the P4 application. e.g. myapp")
	collector_server := flag.String("server", "127.0.0.1:8654", "PIFINA collector address")
	version_flag := flag.Bool("version", false, "show version")
	connect_timeout := flag.Int("connect-timeout", 5, "Connect timeout for the GRPC connection to the switch.")
	sample_interval := flag.Int("sample-interval-ms", 1000, "Sample interval in ms. Default 100ms")

	flag.Parse()

	if *version_flag {
		fmt.Printf("version=%s", utils.Commit)
		os.Exit(0)
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "PIFINA-control-plane",
		Level: hclog.LevelFromString(*logLevel),
		Color: hclog.AutoColor,
	})
	logger.Debug("configured endpoints", "bfrt_endpoint", *bfrt_endpoint, "pifina_collector", *collector_server)

	_, _, err := net.SplitHostPort(*bfrt_endpoint)
	if err != nil {
		logger.Error("Invalid BFRT address. example format 127.0.0.1:50052")
		os.Exit(1)
	}
	if *p4_name == "" {
		logger.Error("Invalid P4 app name. Specify the name of the running P4 application on the switch. e.g. myapp")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	var wg sync.WaitGroup
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	controller := controller.NewTofinoController(logger, *bfrt_endpoint, *p4_name, *collector_server, *sample_interval)
	err = controller.StartController(ctx, &wg, *connect_timeout)
	if err != nil {
		logger.Error("cannot start the controller", "err", err)
	}

	// Block until a termination signal has received and all threads gracefully shutdown
	wg.Wait()
	logger.Info("Graceful shutdown completed. Bye!")
}

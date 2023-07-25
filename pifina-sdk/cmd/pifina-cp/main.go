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
	"github.com/thushjandan/pifina/pkg/controller"
	"github.com/thushjandan/pifina/pkg/debugserver"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	logLevel := flag.String("level", "info", "set the log level. The default is info. Possible options: trace, debug, info, warn, error, off")
	bfrt_endpoint := flag.String("bfrt", "127.0.0.1:50052", "BF runtime GRPC server address (Dataplane endpoint)")
	p4_name := flag.String("p4name", "", "Name of the P4 application. e.g. myapp")
	collector_server := flag.String("server", "127.0.0.1:8654", "PIFINA collector address")
	debug_server := flag.String("debug-server", "127.0.0.1:6060", "Golang debug http pprof listen address")
	api_port := flag.String("port", ":8656", "Controller API port to listen")
	version_flag := flag.Bool("version", false, "show version")
	connect_timeout := flag.Int("connect-timeout", 5, "Connect timeout for the GRPC connection to the switch.")
	sample_interval := flag.Int("sample-interval-ms", 50, "Sample interval in ms. Default 100ms")
	lpf_time_constant_int := flag.Int("lpf-time-ns", 80, "LPF time constant for computing moving average of the ingress jitter value.")
	pipeline_count := flag.Int("pipe-count", 4, "Amount of pipeline existing on the tofino. Used to retrieve TrafficManager metrics per pipeline.")

	flag.Parse()

	if *version_flag {
		fmt.Printf("pifina-cp (PIFINA control plane) version %s, commit %s, built at %s\n", version, commit, date)
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

	// Start Debug server if log level is lower equals debug
	var ds *debugserver.DebugServer
	if logger.GetLevel() <= hclog.Debug {
		ds := debugserver.NewDebugServer(*debug_server)
		ds.StartDebugServer()
	}

	options := &controller.TofinoControllerOptions{
		Logger:                  logger,
		Endpoint:                *bfrt_endpoint,
		ConnectTimeout:          *connect_timeout,
		P4name:                  *p4_name,
		CollectorServerEndpoint: *collector_server,
		SampleInterval:          *sample_interval,
		APIPort:                 *api_port,
		LpfTimeConst:            float32(*lpf_time_constant_int),
		PipelineCount:           *pipeline_count,
	}
	controller := controller.NewTofinoController(options)
	err = controller.StartController(ctx, &wg)
	if err != nil {
		logger.Error("cannot start the controller", "err", err)
	}

	// Shutdown Debug Server if existed
	if ds != nil {
		ds.ShutdownDebugServer()
	}

	// Block until a termination signal has received and all threads gracefully shutdown
	wg.Wait()
	logger.Info("Graceful shutdown completed. Bye!")
}

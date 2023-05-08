package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/internal/utils"
	"github.com/thushjandan/pifina/pkg/model"
	"github.com/thushjandan/pifina/pkg/web/http"
	"github.com/thushjandan/pifina/pkg/web/receiver"
)

func main() {

	logLevel := flag.String("level", "info", "set the log level. The default is info. Possible options: trace, debug, info, warn, error, off")
	listen_metric_port := flag.String("collector", "8654", "PIFINA metric port to listen")
	listen_web_port := flag.String("web", "8655", "PIFINA web port to listen")
	keyFile := flag.String("key", "assets/key.pem", "TLS private key file path")
	certFile := flag.String("cert", "assets/cert.pem", "TLS certificate file path")
	version_flag := flag.Bool("version", false, "show version")

	flag.Parse()

	if *version_flag {
		fmt.Printf("version=%s", utils.Commit)
		os.Exit(0)
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "PIFINA-web",
		Level: hclog.LevelFromString(*logLevel),
		Color: hclog.AutoColor,
	})
	logger.Info("configured listening ports", "web", *listen_web_port, "metric", *listen_metric_port)
	// Termination handling
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	metricChannel := make(chan []*model.MetricItem)

	receiver := receiver.NewPifinaMetricReceiver(logger)
	err := receiver.StartServer(ctx, *listen_metric_port, metricChannel)
	if err != nil {
		logger.Error("cannot start metric receiver", "err", err)
	}
	webServer := http.NewPifinaHttpServer(logger)
	go webServer.StartWebServer(ctx, *listen_web_port, *keyFile, *certFile, metricChannel)

	<-ctx.Done()
	receiver.Shutdown()
	webServer.Shutdown()

}

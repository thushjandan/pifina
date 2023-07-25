package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/model"
	"github.com/thushjandan/pifina/pkg/web/endpoints"
	"github.com/thushjandan/pifina/pkg/web/http"
	"github.com/thushjandan/pifina/pkg/web/receiver"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {

	logLevel := flag.String("level", "info", "set the log level. The default is info. Possible options: trace, debug, info, warn, error, off")
	listen_metric_port := flag.String("collector", "8654", "PIFINA metric port to listen")
	listen_web_port := flag.String("web", "8655", "PIFINA web port to listen")
	controller_api_port := flag.Int("controller", 8656, "Default PIFINA controller API port to connect")
	keyFile := flag.String("key", "assets/key.pem", "TLS private key file path")
	certFile := flag.String("cert", "assets/cert.pem", "TLS certificate file path")
	version_flag := flag.Bool("version", false, "show version")

	flag.Parse()

	if *version_flag {
		fmt.Printf("pifina-web (PIFINA telemetry server) version %s, commit %s, built at %s\n", version, commit, date)
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

	endpointDirectory := endpoints.NewPifinaEndpointDirectory(*controller_api_port)

	telemetryChannel := make(chan *model.TelemetryMessage)

	receiver := receiver.NewPifinaMetricReceiver(logger, endpointDirectory)
	err := receiver.StartServer(ctx, *listen_metric_port, telemetryChannel)
	if err != nil {
		logger.Error("cannot start metric receiver", "err", err)
	}
	webServer := http.NewPifinaHttpServer(logger, endpointDirectory)
	go webServer.StartWebServer(ctx, *listen_web_port, *keyFile, *certFile, telemetryChannel)

	<-ctx.Done()
	receiver.Shutdown()
	webServer.Shutdown()

}

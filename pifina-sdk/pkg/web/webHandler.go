package web

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/model"
	"github.com/thushjandan/pifina/pkg/web/endpoints"
	"github.com/thushjandan/pifina/pkg/web/http"
	"github.com/thushjandan/pifina/pkg/web/receiver"
	"github.com/urfave/cli/v2"
)

func ServeWebserverHandler(cCtx *cli.Context) error {

	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "PIFINA-web",
		Level: hclog.LevelFromString(cCtx.String("level")),
		Color: hclog.AutoColor,
	})
	logger.Info("configured listening ports", "web", cCtx.Uint("listen-web"), "metric", cCtx.Uint("listen-collector"))
	// Termination handling
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	endpointDirectory := endpoints.NewPifinaEndpointDirectory(int(cCtx.Uint("probe-port")))

	telemetryChannel := make(chan *model.TelemetryMessage)

	receiver := receiver.NewPifinaMetricReceiver(logger, endpointDirectory)
	err := receiver.StartServer(ctx, cCtx.Uint("listen-collector"), telemetryChannel)
	if err != nil {
		logger.Error("cannot start metric receiver", "err", err)
		return err
	}
	webServer := http.NewPifinaHttpServer(logger, endpointDirectory)
	go webServer.StartWebServer(ctx, cCtx.Uint("listen-web"), cCtx.String("key"), cCtx.String("cert"), telemetryChannel)

	<-ctx.Done()
	receiver.Shutdown()
	webServer.Shutdown()

	return nil
}

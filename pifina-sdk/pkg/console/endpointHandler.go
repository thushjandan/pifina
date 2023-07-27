package console

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/endpoint"
	"github.com/urfave/cli/v2"
)

func ListMlxDevicesCliAction(cCtx *cli.Context) error {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "PIFINA-cli",
		Level: hclog.LevelFromString(cCtx.String("level")),
		Color: hclog.AutoColor,
	})
	if os.Getuid() != 0 {
		logger.Error("Need to be root. Please use sudo or run as root.")
		os.Exit(1)
		return nil
	}

	neoMode := cCtx.String("neo-mode")
	var neoPort int
	switch neoMode {
	case "shell":
		neoMode = "--mode=shell"
	case "socket":
		neoMode = "--mode=socket"
		neoPort = cCtx.Int("neo-port")
		if neoPort == 0 {
			logger.Error("Missing neo-port parameter.")
			os.Exit(1)
			return nil
		}
	default:
		logger.Error("Invalid neo-mode parameter given. Needs to be either shell or socket", "mode", neoMode)
		os.Exit(1)
		return nil
	}

	logger.Debug("Retrieving system devices")
	collector := endpoint.NewEndpointCollector(&endpoint.EndpointCollectorOptions{
		Logger:  logger,
		SDKPath: cCtx.String("sdk"),
		NEOMode: neoMode,
		NEOPort: neoPort,
	})
	err := collector.ListMlxNetworkCards()
	if err != nil {
		logger.Error("Error occured retrieving all Connect-X NICs", "err", err)
		return err
	}
	return nil
}

func CollectNICPerfCounterCliAction(cCtx *cli.Context) error {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "PIFINA-cli",
		Level: hclog.LevelFromString(cCtx.String("level")),
		Color: hclog.AutoColor,
	})
	if os.Getuid() != 0 {
		logger.Error("Need to be root. Please use sudo or run as root.")
		os.Exit(1)
		return nil
	}

	// Validate neo-mode parameter
	neoMode := cCtx.String("neo-mode")
	var neoPort int

	switch neoMode {
	case "shell":
		neoMode = "--mode=shell"
	case "socket":
		neoMode = "--mode=socket"
		neoPort = cCtx.Int("neo-port")
		if neoPort == 0 {
			logger.Error("Missing neo-port parameter.")
			os.Exit(1)
			return nil
		}
	default:
		logger.Error("Invalid neo-mode parameter given. Needs to be either shell or socket", "mode", neoMode)
		os.Exit(1)
		return nil
	}

	// init signal handler
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	collector := endpoint.NewEndpointCollector(&endpoint.EndpointCollectorOptions{
		Logger:            logger,
		SDKPath:           cCtx.String("sdk"),
		NEOMode:           neoMode,
		NEOPort:           neoPort,
		TelemetryEndpoint: cCtx.String("server"),
	})
	targetDevices := cCtx.StringSlice("dev")
	logger.Debug("Retrieving performance counters", "dev", targetDevices)
	err := collector.GetMlxPerformanceCounters(ctx, targetDevices, cCtx.Int("sample-interval"))
	if err != nil {
		logger.Error("Error occured retrieving all Connect-X NICs", "err", err)
		return err
	}
	return nil
}

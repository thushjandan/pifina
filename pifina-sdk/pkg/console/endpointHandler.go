package console

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/controller/sink"
	"github.com/thushjandan/pifina/pkg/endpoint"
	"github.com/thushjandan/pifina/pkg/model"
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

	var wg sync.WaitGroup

	metricSinkChan := make(chan *model.SinkEmitCommand)

	// Init sink
	sink := sink.NewSink(logger, cCtx.String("server"))
	wg.Add(1)
	go sink.StartSink(ctx, &wg, metricSinkChan)

	collector := endpoint.NewEndpointCollector(&endpoint.EndpointCollectorOptions{
		Logger:         logger,
		MetricSinkChan: metricSinkChan,
		SampleInterval: cCtx.Int("sample-interval"),
		SDKPath:        cCtx.String("sdk"),
		NEOMode:        neoMode,
		NEOPort:        neoPort,
	})
	targetDevices := cCtx.StringSlice("dev")
	logger.Debug("Retrieving performance counters", "dev", targetDevices)
	err := collector.CollectMlxPerfCounters(ctx, &wg, targetDevices)
	if err != nil {
		logger.Error("Error occured retrieving all Connect-X NICs", "err", err)
		return err
	}

	// Wait until all threads have terminated gracefully
	wg.Wait()

	return nil
}

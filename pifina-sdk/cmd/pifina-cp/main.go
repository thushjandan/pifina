package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/internal/utils"
)

func main() {
	logLevel := flag.String("level", "info", "set the log level. The default is info. Possible options: trace, debug, info, warn, error, off")
	bfrt_endpoint := flag.String("bfrt", "127.0.0.1:50052", "BF runtime GRPC server address (Dataplane endpoint)")
	collector_server := flag.String("server", "127.0.0.1:8654", "PIFINA collector address")
	version_flag := flag.Bool("version", false, "show version")

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

}

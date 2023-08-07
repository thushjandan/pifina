package main

import (
	"os"
	"time"

	"github.com/thushjandan/pifina/pkg/console"
	"github.com/urfave/cli/v2"
)

const (
	P4_MATCH_LPM     = "lpm"
	SESSION_ID_WIDTH = 7
)

var (
	version        = "dev"
	commit         = "none"
	date           = time.Now().Format(time.RFC3339)
	P4_MATCH_TYPES = []string{"exact", "ternary", "lpm"}
)

func main() {
	compiled_date, _ := time.Parse(time.RFC3339, date)
	app := &cli.App{
		Name:     "pifina-cli",
		Version:  version,
		Compiled: compiled_date,
		Authors: []*cli.Author{
			{
				Name:  "Thushjandan Ponnudurai",
				Email: "thushjandan@gmail.com",
			},
		},
		Usage: "Customize P4 code with PIFINA cli to match on user-defined header fields.",
		Commands: []*cli.Command{
			{
				Name:    "generate",
				Aliases: []string{"g"},
				Usage:   `Example: pifina-cli create -k hdr.ipv4.protocol:exact -k hdr.ipv4.dstAddr:ternary -k hdr.ipv4.srcAddr:ternary -o src/myP4app/include `,
				Description: `Creates customized Pifina P4 source code with user defined match fields. 
Use for every match key the flag -key and define the name of the field together with its match type delimited by a colon (:) like field1:matchType
In addition the output directory for the generated P4 source code files needs to be defined with flag -o
Following match types can be used: exact, ternary, lpm`,
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:     "key",
						Aliases:  []string{"k"},
						Required: true,
						Usage:    "which P4 header fields to match => Table keys for PIFINA",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Value:   ".",
						Usage:   "output directory for generated P4 source code files.",
					},
					&cli.StringFlag{
						Name:  "ig-hdr",
						Value: "ingress_headers_t",
						Usage: "Name of your ingress header struct.",
					},
					&cli.StringFlag{
						Name:  "eg-hdr",
						Value: "egress_headers_t",
						Usage: "Name of your egress header struct.",
					},
					&cli.IntFlag{
						Name:  "ig-probe",
						Value: 0,
						Usage: "Count of additional ingress probes",
					},
					&cli.IntFlag{
						Name:  "eg-probe",
						Value: 0,
						Usage: "Count of additional egress probes",
					},
					&cli.BoolFlag{
						Name:  "gen-skeleton",
						Value: false,
						Usage: "If true, a basic skeleton of a P4 program with PIFINA will be generated.",
					},
				},
				Action: console.CreateTemplateCliAction,
			},
			{
				Name:    "nic",
				Aliases: []string{"n"},
				Action:  console.ListMlxDevicesCliAction,
				Subcommands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"l"},
						Action:  console.ListMlxDevicesCliAction,
					},
					{
						Name:    "collect",
						Aliases: []string{"c"},
						Action:  console.CollectNICPerfCounterCliAction,
						Flags: []cli.Flag{
							&cli.StringSliceFlag{
								Name:     "dev",
								Aliases:  []string{"d"},
								Required: true,
								Usage:    "Dev-UID, ibdevice name or iface name to collect the metrics. This flag can be used multiple times to collect counter from multiple NICs.",
							},
							&cli.StringFlag{
								Name:     "server",
								Aliases:  []string{"s"},
								Value:    "127.0.0.1:8654",
								Required: false,
								Usage:    "PIFINA collector server address",
							},
							&cli.UintFlag{
								Name:     "group-id",
								Value:    1,
								Required: false,
								Usage:    "Group Identifier for PIFINA collector server. Used to group multiple probes together.",
							},
							&cli.IntFlag{
								Name:     "sample-interval",
								Aliases:  []string{"i"},
								Value:    15,
								Required: false,
								Usage:    "Sample interval in seconds.",
							},
							&cli.BoolFlag{
								Name:     "disable-neohost",
								Value:    false,
								Required: false,
								Usage:    "Do not collect metrics from NEO Host SDK",
							},
						},
					},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "level",
						Value:    "info",
						Required: false,
						Usage:    "log level",
					},
					&cli.StringFlag{
						Name:     "sdk",
						Value:    "/opt/neohost/sdk",
						Required: false,
						Usage:    "Path to the Mellanox NEO-Host SDK folder.",
					},
					&cli.StringFlag{
						Name:     "neo-mode",
						Value:    "shell",
						Required: false,
						Usage:    "Running mode for neohost shell/socket",
					},
					&cli.IntFlag{
						Name:     "neo-port",
						Required: false,
						Usage:    "port where NEO-host is running. Only required if neo-mode=socket",
					},
				},
			},
		},
	}
	// Parse CLI arguments
	app.Run(os.Args)
}

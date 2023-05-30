package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/generator"
	"github.com/thushjandan/pifina/pkg/model"
	"github.com/urfave/cli/v2"
)

const (
	P4_MATCH_LPM     = "lpm"
	SESSION_ID_WIDTH = 7
)

var P4_MATCH_TYPES = []string{"exact", "ternary", "lpm"}

func main() {
	app := &cli.App{
		Name:     "pifina-cli",
		Version:  "0.0.1",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  "Thushjandan Ponnudurai",
				Email: "thushjandan@gmail.com",
			},
		},
		Usage: "Customize P4 code with PIFINA cli to match on user-defined header fields.",
		Commands: []*cli.Command{
			{
				Name:    "create",
				Aliases: []string{"c"},
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
				},
				Action: createAction,
			},
		},
	}
	// Parse CLI arguments
	app.Run(os.Args)
}

func createAction(cCtx *cli.Context) error {
	keys := cCtx.StringSlice("key")
	if len(keys) < 1 {
		return cli.Exit("Define at least one header field as key", 1)
	}
	countLpm := 0
	templateKeys := make([]*model.P4CodeTemplateKey, 0, len(keys))
	for _, k := range keys {
		keyObj := strings.Split(k, ":")
		// key:matchtype => Exact two values after split
		if len(keyObj) != 2 {
			return cli.Exit(
				fmt.Sprintf(
					"%s is an invalid input. An header field needs be defined together with the corresponding match type (%s). Must be following format <key>:<type>",
					k,
					strings.Join(P4_MATCH_TYPES, ", "),
				), 1,
			)
		}
		matchKey := keyObj[0]
		matchType := keyObj[1]
		// Match type to lower case
		matchType = strings.ToLower(matchType)
		// Check if valid match type is given
		validType := false
		for i := range P4_MATCH_TYPES {
			if matchType == P4_MATCH_TYPES[i] {
				validType = true
			}
		}
		if !validType {
			return cli.Exit(
				fmt.Sprintf(
					"%s is an invalid match type. Following match types are supported %s",
					matchType,
					strings.Join(P4_MATCH_TYPES, ", "),
				), 1,
			)
		}

		// Only at most 1 LPM is allowed by P4 compiler
		if matchType == P4_MATCH_LPM {
			countLpm++
		}
		if countLpm > 1 {
			return cli.Exit("Cannot have more than one header field with match type LPM", 1)
		}
		templateKeys = append(templateKeys, &model.P4CodeTemplateKey{
			Name:      matchKey,
			MatchType: matchType,
		})
	}
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "PIFINA-cli",
		Level: hclog.LevelFromString("info"),
		Color: hclog.AutoColor,
	})

	outputDir := filepath.Dir(cCtx.String("output"))

	templateOptions := &model.P4CodeTemplate{
		SessionIdWidth: SESSION_ID_WIDTH,
		MatchKeys:      templateKeys,
	}

	logger.Info("Generating files...")
	// Generator template
	err := generator.GenerateP4App(logger, templateOptions, outputDir)
	if err != nil {
		logger.Error("Error occured!", "err", err)
		os.Exit(1)
		return nil
	}
	logger.Info("All necessary PIFINA files have been generated. Include these files according to the manual in your P4 application source code.")

	return nil
}

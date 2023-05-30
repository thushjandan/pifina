package main

import (
	"fmt"
	"os"
	"strings"
	"time"

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
				Usage:   "Creates customized P4 source code",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:     "key",
						Aliases:  []string{"k"},
						Required: true,
						Usage:    "which P4 header fields to match => Table keys for PIFINA",
					},
				},
				Action: createAction,
			},
		},
	}
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
	templateOptions := &model.P4CodeTemplate{
		SessionIdWidth: SESSION_ID_WIDTH,
		MatchKeys:      templateKeys,
	}

	// Generator template
	err := generator.GenerateP4App(templateOptions)
	if err != nil {
		return cli.Exit(err, 1)
	}

	return nil
}

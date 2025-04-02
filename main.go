package main

import (
	"context"
	"d2tool/heroesGrid"
	"d2tool/startup"
	"fmt"
	cli "github.com/urfave/cli/v3"
	"log"
	"os"
	"slices"
)

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:  "startup",
				Usage: "options for the application run on OS startup",
				Commands: []*cli.Command{
					{
						Name: "register",
						Usage: "register the application to run on OS startup\n" +
							"the program will be run with the heroes-grid command, so you can use all its arguments",
						SkipFlagParsing: true,
						Action: func(ctx context.Context, cmd *cli.Command) error {
							return startup.StartupRegister(cmd.Args().Slice())
						},
					},
					{
						Name:  "remove",
						Usage: "remove the application from OS startup",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							return startup.StartupRemove()
						},
					},
				},
			},
			{
				Name:  "heroes-grid",
				Usage: "update the hero grid config files",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:  "config",
						Usage: "Provide a `PATH` to a hero grid config file, can be used multiple times",
						Config: cli.StringConfig{
							TrimSpace: true,
						},
					},
					&cli.StringSliceFlag{
						Name:  "positions",
						Usage: "Provide positions for the config in needed order for a generated config, comma-separated",
						Config: cli.StringConfig{
							TrimSpace: true,
						},
						Value: []string{"1", "2", "3", "4", "5"},
						Validator: func(positions []string) error {
							validPositions := []string{"1", "2", "3", "4", "5"}
							for _, pos := range positions {
								if !slices.Contains(validPositions, pos) {
									return fmt.Errorf("invalid position %s", pos)
								}
							}
							return nil
						},
					},
					&cli.BoolFlag{
						Name:  "periodic",
						Usage: "Set this flag to update the config files periodically in background",
						Value: false,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					config := heroesGrid.UpdateHeroGridConfig{
						ConfigFilePaths: cmd.StringSlice("config"),
						Positions:       cmd.StringSlice("positions"),
						Periodic:        cmd.Bool("periodic"),
					}
					return heroesGrid.UpdateHeroesGrid(config)
				},
			},
		},
		Name:           "d2tool",
		Usage:          "dota 2 tool",
		DefaultCommand: "help",
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

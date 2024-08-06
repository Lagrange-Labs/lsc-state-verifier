package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/Lagrange-Labs/lagrange-node/logger"
	"github.com/Lagrange-Labs/lsc-state-verifier/config"
	"github.com/Lagrange-Labs/lsc-state-verifier/db"
	"github.com/Lagrange-Labs/lsc-state-verifier/utils"
	"github.com/urfave/cli/v2"
)

var (
	configFileFlag = &cli.StringFlag{
		Name:    config.FlagCfg,
		Value:   "./config.toml",
		Usage:   "Configuration `FILE`",
		Aliases: []string{"c"},
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "State Committee Verifier"

	app.Commands = []*cli.Command{
		{
			Name:  "version",
			Usage: "Prints the version of the State Committee Verifier",
			Action: func(c *cli.Context) error {
				w := os.Stdout
				fmt.Fprintf(w, "Version:      %s\n", "v0.1.0")
				fmt.Fprintf(w, "Go version:   %s\n", runtime.Version())
				fmt.Fprintf(w, "OS/Arch:      %s/%s\n", runtime.GOOS, runtime.GOARCH)
				return nil
			},
		},
		{
			Name:  "api",
			Usage: "Uses API to get state proof data",
			Flags: []cli.Flag{
				configFileFlag,
			},
			Action: fetchStateProofFromAPI,
		},
		{
			Name:  "db",
			Usage: "Uses db call to get state proof data",
			Flags: []cli.Flag{
				configFileFlag,
			},
			Action: fetchStateProofFromDB,
		},
	}
	err := app.Run(os.Args)

	if err != nil {
		logger.Fatalf("Error running app: ", err)
		os.Exit(1)
	}
}

func fetchStateProofFromAPI(c *cli.Context) error {
	cfg, err := config.LoadCLIConfig(c)
	if err != nil {
		logger.Fatalf("Error loading config: %s", err)
	}

	var wg sync.WaitGroup
	for _, chain := range cfg.Chains {
		wg.Add(1)
		go func(chain config.ChainConfig) {
			defer wg.Done()
			err := utils.ProcessChainUsingAPI(cfg.ApiUrl, chain)
			if err != nil {
				logger.Errorf("Error processing chain ID %d: %s", chain.ChainID, err)
			}
		}(chain)
	}
	wg.Wait()
	return nil
}

func fetchStateProofFromDB(c *cli.Context) error {
	cfg, err := config.LoadCLIConfig(c)
	if err != nil {
		logger.Fatalf("Error loading config: %s", err)
	}

	mongoDB, err := db.NewMongoDatabase(cfg.DatabaseURI)
	if err != nil {
		logger.Fatalf("Error creating MongoDB client: %s", err)
	}

	var wg sync.WaitGroup
	for _, chain := range cfg.Chains {
		wg.Add(1)
		go func(chain config.ChainConfig) {
			defer wg.Done()
			err := utils.ProcessChainUsingDB(mongoDB, chain)
			if err != nil {
				logger.Errorf("Error processing chain ID %d: %s", chain.ChainID, err)
			}
		}(chain)
	}
	wg.Wait()
	return nil
}

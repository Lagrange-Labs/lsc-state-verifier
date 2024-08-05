package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Lagrange-Labs/lagrange-node/logger"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

const FlagCfg = "config"

type ChainConfig struct {
	ChainID         int64 `mapstructure:"chain_id"`
	FromBatchNumber int64 `mapstructure:"from_batch_number"`
}

type CLIConfig struct {
	ApiUrl      string        `mapstructure:"api_url"`
	Chains      []ChainConfig `mapstructure:"chains"`
	DatabaseURI string        `mapstructure:"database_uri"`
}

// LoadCLIConfig loads the State Committee Verifier configuration.
func LoadCLIConfig(ctx *cli.Context) (*CLIConfig, error) {
	var cfg CLIConfig
	viper.SetConfigType("toml")

	configFilePath := ctx.String(FlagCfg)
	if configFilePath != "" {
		dirName, fileName := filepath.Split(configFilePath)

		fileExtension := strings.TrimPrefix(filepath.Ext(fileName), ".")
		fileNameWithoutExtension := strings.TrimSuffix(fileName, "."+fileExtension)

		viper.AddConfigPath(dirName)
		viper.SetConfigName(fileNameWithoutExtension)
		viper.SetConfigType(fileExtension)
	}
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("STATE_VERIFIER")
	if err := viper.ReadInConfig(); err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if !ok {
			return nil, err
		} else if len(configFilePath) > 0 {
			logger.Warnf("config file `%s` not found, the path should be absolute or relative to the current working directory like `./config.toml`", configFilePath)
			return nil, fmt.Errorf("config file not found: %s", err)
		}
	}

	decodeHooks := []viper.DecoderConfigOption{
		// this allows arrays to be decoded from env var separated by ",", example: MY_VAR="value1,value2,value3"
		viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(mapstructure.TextUnmarshallerHookFunc(), mapstructure.StringToSliceHookFunc(","))),
	}

	if err := viper.Unmarshal(&cfg, decodeHooks...); err != nil {
		return nil, err
	}

	return &cfg, nil
}

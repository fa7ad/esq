package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fa7ad/esq/internal/options"
	"github.com/fa7ad/esq/internal/validation"
)

const (
	DefaultSize = 100
	AppName     = "esq"
)

func InitConfig(cfgFile string, appName string, args *options.CliArgs) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigName("." + appName)
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix(strings.ToUpper(appName))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// No config file found; optional.
	} else {
		return fmt.Errorf("error reading config file: %w", err)
	}

	if err := viper.Unmarshal(args); err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return validation.ValidateCliArgs(*args)
}

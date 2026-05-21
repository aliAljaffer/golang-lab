// 05-env-and-config — env vars + viper config precedence.
//
// Goal: resolve a "log_level" setting using this precedence (highest first):
//   1. --log-level CLI flag
//   2. APP_LOG_LEVEL env var
//   3. ./config.yaml ("log_level: ...")
//   4. built-in default "info"
//
// Demonstrates: os.LookupEnv vs os.Getenv, viper.BindPFlag, viper.BindEnv,
// viper.SetConfigName / SetConfigType / AddConfigPath, viper.ReadInConfig.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "envconfig",
		Short: "demo: env + config + flag precedence",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: print the resolved log level via viper.GetString("log_level").
			return nil
		},
	}

	// TODO: rootCmd.Flags().String("log-level", "info", "log level")

	// TODO: bind the flag to viper under the key "log_level".
	//   viper.BindPFlag("log_level", rootCmd.Flags().Lookup("log-level"))

	// TODO: tell viper to also read APP_LOG_LEVEL.
	//   viper.SetEnvPrefix("APP")
	//   viper.AutomaticEnv()
	//   viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// TODO: configure config file lookup.
	//   viper.SetConfigName("config")
	//   viper.SetConfigType("yaml")
	//   viper.AddConfigPath(".")
	//   _ = viper.ReadInConfig() // OK if missing

	// TODO (bonus): demonstrate os.LookupEnv("APP_LOG_LEVEL")
	// — returns (value, ok). ok is false when the var is unset (different from empty).

	_ = viper.GetString // remove this line once you actually call viper above

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

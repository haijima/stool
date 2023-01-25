package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func useConfig(cmd *cobra.Command) {
	var cfgFile string

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.stool.yaml)")

	cobra.OnInitialize(func() {
		if cfgFile != "" {
			// Use config file from the flag.
			viper.SetConfigFile(cfgFile)
		} else {
			wd, err := os.Getwd()
			cobra.CheckErr(err)
			viper.AddConfigPath(wd)
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)
			viper.AddConfigPath(home)
			viper.SetConfigName(".stool")
		}

		viper.AutomaticEnv() // read in environment variables that match

		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err == nil {
			zap.L().Info(fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed()))
		}
	})

}

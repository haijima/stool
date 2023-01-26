package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func useConfig(cmd *cobra.Command, v *viper.Viper) {
	var cfgFile string

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.stool.yaml)")

	cobra.OnInitialize(func() {
		if cfgFile != "" {
			// Use config file from the flag.
			v.SetConfigFile(cfgFile)
		} else {
			wd, err := os.Getwd()
			cobra.CheckErr(err)
			v.AddConfigPath(wd)
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)
			v.AddConfigPath(home)
			v.SetConfigName(".stool")
		}

		v.AutomaticEnv() // read in environment variables that match

		// If a config file is found, read it in.
		if err := v.ReadInConfig(); err == nil {
			zap.L().Info(fmt.Sprintf("Using config file: %s", v.ConfigFileUsed()))
		}
	})
}

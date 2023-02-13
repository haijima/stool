package main

import (
	"os"

	"github.com/haijima/stool/cmd"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func main() {
	rootCmd := cmd.NewRootCmd(viper.New(), afero.NewOsFs())
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

package main

import (
	"os"

	"github.com/haijima/stool/cmd"
	"github.com/mattn/go-colorable"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func main() {
	rootCmd := cmd.NewRootCmd(viper.New(), afero.NewOsFs())
	rootCmd.SetErr(colorable.NewColorableStderr())
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

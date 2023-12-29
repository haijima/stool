package main

import (
	"log/slog"
	"os"

	"github.com/haijima/stool/cmd"
	"github.com/mattn/go-colorable"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func main() {
	slog.SetDefault(slog.New(cmd.NewCliSlogHandler(nil)))
	v := viper.NewWithOptions(viper.WithLogger(slog.Default()))
	fs := afero.NewOsFs()
	v.SetFs(fs)
	rootCmd := cmd.NewRootCmd(v, fs)
	rootCmd.SetOut(colorable.NewColorableStdout())
	rootCmd.SetErr(colorable.NewColorableStderr())
	if err := rootCmd.Execute(); err != nil {
		slog.Error(err.Error(), slog.Any(cmd.TraceErrorKey, err))
		os.Exit(1)
	}
}

package main

import (
	"log/slog"
	"os"

	"github.com/haijima/cobrax"
	"github.com/haijima/stool/cmd"
	"github.com/haijima/stool/internal"
	"github.com/mattn/go-colorable"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

var (
	// https://goreleaser.com/cookbooks/using-main.version/
	version string
	commit  string
	date    string
)

func main() {
	slog.SetDefault(slog.New(internal.NewCliSlogHandler(nil)))
	v := viper.NewWithOptions(viper.WithLogger(slog.Default()))
	fs := afero.NewOsFs()
	v.SetFs(fs)
	rootCmd := cmd.NewRootCmd(v, fs)
	rootCmd.Version = cobrax.VersionFunc(version, commit, date)
	rootCmd.SetOut(colorable.NewColorableStdout())
	rootCmd.SetErr(colorable.NewColorableStderr())
	if err := rootCmd.Execute(); err != nil {
		slog.Error(err.Error(), slog.Any(internal.TraceErrorKey, err))
		os.Exit(1)
	}
}

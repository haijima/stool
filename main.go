package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/haijima/cobrax"
	"github.com/haijima/stool/cmd"
	"github.com/haijima/stool/internal"
	"github.com/mattn/go-colorable"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// https://goreleaser.com/cookbooks/using-main.version/
var version, commit, date string

func main() {
	stdout, stderr := colorable.NewColorableStdout(), colorable.NewColorableStderr()
	l := slog.New(internal.NewCliSlogHandler(stderr, &slog.HandlerOptions{Level: &cmd.Lv}))
	slog.SetDefault(l)
	cobrax.SetLogger(l)
	v := viper.NewWithOptions(viper.WithLogger(l))
	fs := afero.NewOsFs()
	v.SetFs(fs)
	rootCmd := cmd.NewRootCmd(v, fs)
	rootCmd.Version = cobrax.VersionFunc(version, commit, date)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	if err := rootCmd.Execute(); err != nil {
		if slog.Default().Enabled(rootCmd.Context(), slog.LevelDebug) {
			slog.Error(fmt.Sprintf("%+v", err))
		} else {
			slog.Error(err.Error())
		}
		os.Exit(1)
	}
}

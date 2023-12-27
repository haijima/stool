package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func init() {

}

// NewRootCmd returns the base command used when called without any subcommands
func NewRootCmd(v *viper.Viper, fs afero.Fs) *cobra.Command {
	rootCmd := &cobra.Command{}
	rootCmd.Use = "stool"
	rootCmd.Short = "stool is access log profiler"
	rootCmd.Args = cobra.NoArgs
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.SilenceUsage = true // don't show help content when error occurred
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		slog.SetDefault(Logger(*v))
		color.NoColor = color.NoColor || v.GetBool("no_color")
		if err := ReadConfigFile(v); err != nil {
			return err
		}
		if err := v.MergeConfigMap(v.GetStringMap(strings.ToLower(cmd.Name()))); err != nil {
			return err
		}
		return v.BindPFlags(cmd.Flags())
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/.stool.yaml)")
	//rootCmd.PersistentFlags().BoolP("version", "v", false, "Show the version of this command")
	rootCmd.PersistentFlags().Int("verbosity", 0, "verbosity level")
	_ = v.BindPFlag("verbosity", rootCmd.PersistentFlags().Lookup("verbosity"))
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "quiet output")
	_ = v.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	rootCmd.PersistentFlags().Bool("no_color", false, "disable colorized output")
	_ = v.BindPFlag("no_color", rootCmd.PersistentFlags().Lookup("no_color"))

	rootCmd.AddCommand(NewTrendCmd(internal.NewTrendProfiler(), v, fs))
	rootCmd.AddCommand(NewTransitionCmd(internal.NewTransitionProfiler(), v, fs))
	rootCmd.AddCommand(NewScenarioCmd(internal.NewScenarioProfiler(), v, fs))
	rootCmd.AddCommand(NewParamCmd(internal.NewParamProfiler(), v, fs))
	rootCmd.AddCommand(NewAaCmd(v, fs))
	rootCmd.AddCommand(NewGenConfCmd(v, fs))

	return rootCmd
}

func OpenOrStdIn(filename string, fs afero.Fs, stdin io.Reader) (io.ReadCloser, error) {
	if filename != "" {
		f, err := fs.Open(filename)
		if err != nil {
			return nil, err
		}
		return f, nil
	} else {
		return io.NopCloser(stdin), nil
	}
}

func ReadConfigFile(v *viper.Viper) error {
	if cfgFile != "" {
		v.SetConfigFile(cfgFile) // Use config file from the flag.
		// If a config file is found, read it in.
		if err := v.ReadInConfig(); err == nil {
			slog.Info(fmt.Sprintf("Using config file: %s", v.ConfigFileUsed()))
			slog.Debug(fmt.Sprintf("%+v", v.AllSettings()))
		}
	} else {
		xdgViper := viper.New()
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			xdgConfig = filepath.Join(home, ".config")
		}
		xdgViper.AddConfigPath(filepath.Join(xdgConfig, "stool")) // use XDG config directory as global config path
		xdgViper.SetConfigName("config")

		homeViper := viper.New()
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		homeViper.AddConfigPath(home) // use home directory as global config path
		homeViper.SetConfigName(".stool")

		projectViper := viper.New()
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		projectViper.AddConfigPath(wd) // use current working directory as project config path
		projectViper.SetConfigName(".stool")

		// Read in global config file (XDG > HOME)
		if err := xdgViper.ReadInConfig(); err == nil {
			slog.Info(fmt.Sprintf("Using file as global configuration: %s", xdgViper.ConfigFileUsed()))
			slog.Debug(fmt.Sprintf("%+v", xdgViper.AllSettings()))
		} else if err := homeViper.ReadInConfig(); err == nil {
			slog.Info(fmt.Sprintf("Using file as global configuration: %s", homeViper.ConfigFileUsed()))
			slog.Debug(fmt.Sprintf("%+v", homeViper.AllSettings()))
		}

		// Read in project config file
		if err := projectViper.ReadInConfig(); err == nil {
			slog.Info(fmt.Sprintf("Using file as project configuration: %s", projectViper.ConfigFileUsed()))
			slog.Debug(fmt.Sprintf("%+v", projectViper.AllSettings()))
		}

		// Merge all config files
		cobra.CheckErr(v.MergeConfigMap(xdgViper.AllSettings()))
		cobra.CheckErr(v.MergeConfigMap(homeViper.AllSettings()))
		cobra.CheckErr(v.MergeConfigMap(projectViper.AllSettings()))
		slog.Debug(fmt.Sprintf("%+v", v.AllSettings()))
	}
	return nil
}

func Logger(v viper.Viper) *slog.Logger {
	if v.GetBool("quiet") {
		return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	} else if v.GetInt("verbosity") >= 3 {
		return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))
	} else if v.GetInt("verbosity") == 2 {
		return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else if v.GetInt("verbosity") == 1 {
		return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	} else {
		return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	}
}

package cmd

import (
	"fmt"
	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"runtime/debug"
	"time"
)

// Version is set in build step
var Version = ""

// NewRootCmd returns the base command used when called without any subcommands
func NewRootCmd(v *viper.Viper, fs afero.Fs) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "stool",
		Short:        "stool is access log profiler",
		SilenceUsage: true, // don't show help content when error occurred
		Version:      version(),
	}

	rootCmd.Flags().Bool("version", false, "Show the version of this command")
	addLoggingOption(rootCmd, v)
	useConfig(rootCmd, v)

	v.SetFs(fs)

	rootCmd.AddCommand(NewTrendCommand(internal.NewTrendProfiler(), v, fs))
	rootCmd.AddCommand(NewTransitionCmd(internal.NewTransitionProfiler(), v, fs))
	rootCmd.AddCommand(NewScenarioCmd(internal.NewScenarioProfiler(), v, fs))
	rootCmd.AddCommand(NewAaCommand(v, fs))

	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	return rootCmd
}

func addLoggingOption(cmd *cobra.Command, v *viper.Viper) *zap.Logger {
	cmd.PersistentFlags().Bool("debug", false, "debug level output")
	cmd.PersistentFlags().BoolP("verbose", "v", false, "verbose level output")
	cmd.MarkFlagsMutuallyExclusive("debug", "verbose")
	_ = v.BindPFlags(cmd.PersistentFlags())

	logger := zap.NewNop()
	var restoreGlobal func()
	cobra.OnInitialize(func() {
		if v.GetBool("debug") {
			cfg := zap.NewProductionConfig() // human readable
			cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
			cfg.DisableStacktrace = true
			debugLogger, err := cfg.Build()
			cobra.CheckErr(err)
			logger = debugLogger
		} else if v.GetBool("verbose") {
			cfg := zap.NewDevelopmentConfig() // json
			cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
			cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Local().Format("2006-01-02 15:04:05 MST"))
			}
			verboseLogger, err := cfg.Build()
			cobra.CheckErr(err)
			logger = verboseLogger
		}

		restoreGlobal = zap.ReplaceGlobals(logger)
	})

	cobra.OnFinalize(func() {
		_ = logger.Sync()
		restoreGlobal()
	})

	return logger
}

func useConfig(cmd *cobra.Command, v *viper.Viper) {
	var cfgFile string
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.stool.yaml)")

	cobra.OnInitialize(func() {
		if cfgFile != "" {
			v.SetConfigFile(cfgFile) // Use config file from the flag.
		} else {
			wd, err := os.Getwd()
			cobra.CheckErr(err)
			v.AddConfigPath(wd) // ./.stool.yaml
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)
			v.AddConfigPath(home) // ~/.stool.yaml
			v.SetConfigName(".stool")
		}

		v.AutomaticEnv() // read in environment variables that match

		// If a config file is found, read it in.
		if err := v.ReadInConfig(); err == nil {
			zap.L().Info(fmt.Sprintf("Using config file: %s", v.ConfigFileUsed()))
			zap.L().Debug(fmt.Sprintf("%+v", v.AllSettings()))
		}
	})
}

func version() string {
	if Version != "" {
		return Version
	}
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		return buildInfo.Main.Version
	}
	return "(devel)"
}

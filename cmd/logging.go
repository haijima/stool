package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func addLoggingOption(cmd *cobra.Command, v *viper.Viper) *zap.Logger {
	cmd.PersistentFlags().Bool("debug", false, "debug level output")
	cmd.PersistentFlags().BoolP("verbose", "v", false, "verbose level output")
	cmd.MarkFlagsMutuallyExclusive("debug", "verbose")
	_ = v.BindPFlags(cmd.PersistentFlags())

	var logger *zap.Logger
	var restoreGlobal func()
	cobra.OnInitialize(func() {
		if v.GetBool("debug") {
			cfg := zap.NewProductionConfig()
			cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
			cfg.DisableStacktrace = true
			debugLogger, err := cfg.Build()
			cobra.CheckErr(err)
			logger = debugLogger
		} else if v.GetBool("verbose") {
			cfg := zap.NewDevelopmentConfig()
			cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
			cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Local().Format("2006-01-02 15:04:05 MST"))
			}
			verboseLogger, err := cfg.Build()
			cobra.CheckErr(err)
			logger = verboseLogger
		} else {
			logger = zap.NewNop()
		}

		restoreGlobal = zap.ReplaceGlobals(logger)
	})

	cobra.OnFinalize(func() {
		_ = logger.Sync()
		restoreGlobal()
	})

	return logger
}

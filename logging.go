package stool

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggingMode represents a logging configuration specification.
type LoggingMode int

// LoggingMode values
const (
	LoggingNop LoggingMode = iota
	LoggingVerbose
	LoggingDebug
)

var (
	logging = LoggingNop

	// DebugLogConfig is used to generate a *zap.Logger for debug mode.
	DebugLogConfig = func() zap.Config {
		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		cfg.DisableStacktrace = true
		return cfg
	}()
	// VerboseLogConfig is used to generate a *zap.Logger for verbose mode.
	VerboseLogConfig = func() zap.Config {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Local().Format("2006-01-02 15:04:05 MST"))
		}
		return cfg
	}()
)

// Debug sets a debug logger in global.
func Debug() func() {
	logging = LoggingDebug
	return replaceLogger(DebugLogConfig)
}

// Verbose sets a verbose logger in global.
func Verbose() func() {
	logging = LoggingVerbose
	return replaceLogger(VerboseLogConfig)
}

// IsDebug returns true if a debug logger is used.
func IsDebug() bool { return logging == LoggingDebug }

// IsVerbose returns true if a verbose logger is used.
func IsVerbose() bool { return logging == LoggingVerbose }

// Logging returns a current logging mode.
func Logging() LoggingMode { return logging }

func replaceLogger(cfg zap.Config) func() {
	l, err := cfg.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize a debug logger: %v\n", err)
	}

	restoreOriginal := zap.ReplaceGlobals(l)
	return func() {
		l.Sync()
		restoreOriginal()
	}
}

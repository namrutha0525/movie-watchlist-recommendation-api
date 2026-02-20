package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a structured Zap logger.
// In production mode it writes JSON; in development it writes human-readable output.
func New(mode string) *zap.Logger {
	var cfg zap.Config

	if mode == "release" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	logger, err := cfg.Build()
	if err != nil {
		// Fallback — should never happen with default configs
		fallback, _ := zap.NewProduction()
		fallback.Error("failed to build logger, using fallback", zap.Error(err))
		return fallback
	}

	// Replace the global logger so zap.L() / zap.S() work everywhere
	zap.ReplaceGlobals(logger)

	logger.Info("logger initialized",
		zap.String("mode", mode),
		zap.Int("pid", os.Getpid()),
	)

	return logger
}

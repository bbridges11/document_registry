package logger

import (
	"github.com/bbridges_11/document-registry/internal/platform/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(cfg config.LoggingConfig) (*zap.Logger, error) {
	var zapConfig zap.Config

	if cfg.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	return zapConfig.Build()
}

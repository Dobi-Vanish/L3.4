package logger

import (
	"L3.4/internal/config"

	"github.com/wb-go/wbf/logger"
)

func Init(cfg *config.Config) (logger.Logger, error) {
	lvl := parseLevel(cfg.LogLevel)
	opts := []logger.Option{
		logger.WithLevel(lvl),
	}
	return logger.InitLogger(logger.ZerologEngine, "image-service", "dev", opts...)
}

func parseLevel(level string) logger.Level {
	switch level {
	case "debug":
		return logger.DebugLevel
	case "info":
		return logger.InfoLevel
	case "warn":
		return logger.WarnLevel
	case "error":
		return logger.ErrorLevel
	default:
		return logger.InfoLevel
	}
}

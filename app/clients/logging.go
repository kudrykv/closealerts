package clients

import (
	"fmt"

	"go.uber.org/zap"
)

func NewLogger() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment(
		zap.AddStacktrace(zap.DPanicLevel),
		zap.IncreaseLevel(zap.InfoLevel),
		zap.WithCaller(false),
	)
	if err != nil {
		return nil, fmt.Errorf("zap new development: %w", err)
	}

	return logger, nil
}

func NewSugaredLogger(log *zap.Logger) *zap.SugaredLogger {
	return log.Sugar()
}

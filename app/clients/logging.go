package clients

import (
	"fmt"

	"go.uber.org/zap"
)

func NewLogger() (*zap.Logger, error) {
	logger, err := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
		DisableCaller:     true,
		DisableStacktrace: true,
		Encoding:          "json",
		EncoderConfig:     zap.NewDevelopmentEncoderConfig(),
		OutputPaths:       []string{"stderr", "./log.log"},
		ErrorOutputPaths:  []string{"stderr", "./internal.log"},
		InitialFields:     nil,
	}.Build()

	if err != nil {
		return nil, fmt.Errorf("zap new development: %w", err)
	}

	return logger, nil
}

func NewSugaredLogger(log *zap.Logger) *zap.SugaredLogger {
	return log.Sugar()
}

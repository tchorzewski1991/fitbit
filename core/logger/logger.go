package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(service string, build string) (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableCaller = true
	config.InitialFields = map[string]any{
		"service": service,
		"version": build,
	}

	l, err := config.Build()
	if err != nil {
		return nil, err
	}

	return l.Sugar(), nil
}

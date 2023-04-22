package logger

import (
	"context"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func provideZapLogger(lifecycle fx.Lifecycle) (*zap.Logger, error) {
	logConfig := zap.NewProductionConfig()
	logEncoderConfig := zap.NewProductionEncoderConfig()

	logEncoderConfig.ConsoleSeparator = "\t"
	logConfig.Encoding = "console"

	logConfig.Sampling = nil
	logConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logConfig.EncoderConfig = logEncoderConfig

	logger, err := logConfig.Build()
	if err != nil {
		return nil, err
	}
	lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			_ = logger.Sync()
			return nil
		},
	})
	return logger, nil
}

// var LoggerExport = fx.Options(fx.Provide(provideZapLogger), fx.Provide(provideZapSugar), fx.NopLogger)

var Export = fx.Options(fx.Provide(provideZapLogger), fx.WithLogger(func(zapLogger *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{Logger: zapLogger}
}))

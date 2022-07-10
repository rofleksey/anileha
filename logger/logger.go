package logger

import (
	"context"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func provideZapLogger(lifecycle fx.Lifecycle) (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return logger.Sync()
		},
	})
	return logger, nil
}

func provideZapSugar(lifecycle fx.Lifecycle, zapLogger *zap.Logger) (*zap.SugaredLogger, error) {
	sugar := zapLogger.Sugar()
	lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return sugar.Sync()
		},
	})
	return sugar, nil
}

// var LoggerExport = fx.Options(fx.Provide(provideZapLogger), fx.Provide(provideZapSugar), fx.NopLogger)

var Export = fx.Options(fx.Provide(provideZapLogger), fx.Provide(provideZapSugar), fx.WithLogger(func(zapLogger *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{Logger: zapLogger}
}))

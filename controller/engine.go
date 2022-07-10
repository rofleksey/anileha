package controller

import (
	"anileha/config"
	"context"
	"fmt"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"time"
)

func newEngine(logger *zap.Logger) (*gin.Engine, error) {
	// gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.MaxMultipartMemory = 1024 * 1024 * 5
	err := engine.SetTrustedProxies(nil)
	if err != nil {
		return nil, err
	}
	engine.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	engine.Use(ginzap.RecoveryWithZap(logger, true))
	return engine, nil
}

func startEngine(lifecycle fx.Lifecycle, log *zap.SugaredLogger, config *config.Config, gin *gin.Engine, shutdowner fx.Shutdowner) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				log.Infof("Starting application on port %d", config.Rest.Port)
				go func() {
					err := gin.Run(fmt.Sprintf(":%d", config.Rest.Port))
					log.Error("Gin fatal error", err)
					err = shutdowner.Shutdown()
					if err != nil {
						log.Fatal("Failed to shutdown gracefully")
					}
				}()
				return nil
			},
		},
	)
}

var RestExport = fx.Options(fx.Provide(newEngine), fx.Invoke(startEngine))

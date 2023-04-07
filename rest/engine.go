package rest

import (
	"anileha/config"
	"anileha/db"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"path"
	"time"
)

func newEngine(config *config.Config, logger *zap.Logger) (*gin.Engine, error) {
	gob.Register(&db.AuthUser{})
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.MaxMultipartMemory = 1024 * 1024 * 5
	err := engine.SetTrustedProxies(nil)
	if err != nil {
		return nil, err
	}

	// logging
	engine.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	engine.Use(ginzap.RecoveryWithZap(logger, true))

	// frontend
	engine.Use(static.Serve("/", static.LocalFile(path.Join("frontend", "dist"), false)))

	// error handling
	engine.Use(ErrorMiddleware(logger))

	// user login
	hashKey := []byte(config.User.CookieHashKey)
	if len(hashKey) != 32 && len(hashKey) != 64 {
		return nil, errors.New(fmt.Sprintf("hash key length should be 32 or 64 bytes, current = %d", len(hashKey)))
	}
	encryptKey := []byte(config.User.CookieEncryptKey)
	if len(encryptKey) != 16 && len(encryptKey) != 24 && len(encryptKey) != 32 {
		return nil, errors.New(fmt.Sprintf("encrypt key length should be 16, 24 or 32 bytes, current = %d", len(encryptKey)))
	}
	store := cookie.NewStore(hashKey, encryptKey)
	engine.Use(sessions.Sessions("login_session", store))

	return engine, nil
}

func startEngine(lifecycle fx.Lifecycle, log *zap.Logger, config *config.Config, gin *gin.Engine, shutdowner fx.Shutdowner) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				log.Info("server started", zap.Uint("port", config.Rest.Port))
				go func() {
					err := gin.Run(fmt.Sprintf("0.0.0.0:%d", config.Rest.Port))
					log.Error("rest fatal error", zap.Error(err))
					err = shutdowner.Shutdown()
					if err != nil {
						log.Fatal("failed to shutdown gracefully", zap.String("where", "rest"), zap.Error(err))
					}
				}()
				return nil
			},
		},
	)
}

var Export = fx.Options(fx.Provide(newEngine), fx.Invoke(startEngine))

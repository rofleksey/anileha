package controller

import (
	"anileha/config"
	"anileha/db"
	"anileha/util"
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
	"net/http"
	"path"
	"time"
)

func newEngine(config *config.Config, logger *zap.Logger) (*gin.Engine, error) {
	gob.Register(&db.AuthUser{})
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.RedirectTrailingSlash = false
	engine.RedirectFixedPath = false
	engine.RemoveExtraSlash = false
	engine.MaxMultipartMemory = 1024 * 1024 * 5
	err := engine.SetTrustedProxies(nil)
	if err != nil {
		return nil, err
	}

	// logging
	engine.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	engine.Use(ginzap.RecoveryWithZap(logger, true))
	engine.Use(static.Serve("/", static.LocalFile(path.Join("frontend", "dist"), false)))

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

var UserKey = "user"

func forceLoginAsAdmin(config *config.Config, session sessions.Session) error {
	user := db.NewUser(config.Admin.Username, "", "")
	user.Admin = true
	session.Set(UserKey, db.NewAuthUser(user))
	err := session.Save()
	return err
}

func CheckLocalhostAdmin(config *config.Config, session sessions.Session, entry interface{}, c *gin.Context) bool {
	if c.ClientIP() == "::1" {
		if entry == nil || !entry.(*db.AuthUser).Admin {
			if err := forceLoginAsAdmin(config, session); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return true
			}
		}
		c.Next()
		return true
	}
	return false
}

func AdminMiddleware(config *config.Config) func(c *gin.Context) {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		entry := session.Get(UserKey)
		if CheckLocalhostAdmin(config, session, entry, c) {
			return
		}
		if entry == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": util.ErrUnauthorized.Error()})
			return
		}
		user := entry.(*db.AuthUser)
		if !user.Admin {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": util.ErrUnauthorized.Error()})
			return
		}
		c.Next()
	}
}

func startEngine(lifecycle fx.Lifecycle, log *zap.Logger, config *config.Config, gin *gin.Engine, shutdowner fx.Shutdowner) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				log.Info("server started", zap.Uint("port", config.Rest.Port))
				go func() {
					err := gin.Run(fmt.Sprintf("0.0.0.0:%d", config.Rest.Port))
					log.Error("gin fatal error", zap.Error(err))
					err = shutdowner.Shutdown()
					if err != nil {
						log.Fatal("failed to shutdown gracefully", zap.String("where", "gin"), zap.Error(err))
					}
				}()
				return nil
			},
		},
	)
}

var RestExport = fx.Options(fx.Provide(newEngine), fx.Invoke(startEngine))

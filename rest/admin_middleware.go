package rest

import (
	"anileha/config"
	"anileha/db"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

var UserKey = "user"

func forceLoginAsAdmin(config *config.Config, session sessions.Session) error {
	user := db.User{
		Login: config.Admin.Username,
		Admin: true,
	}
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

func AdminMiddleware(log *zap.Logger, config *config.Config) func(c *gin.Context) {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		entry := session.Get(UserKey)
		//if CheckLocalhostAdmin(config, session, entry, c) {
		//	return
		//}
		if entry == nil {
			log.Debug("user is not authorized")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorizedInst.Error()})
			return
		}
		user, _ := entry.(*db.AuthUser)
		if !user.Admin {
			log.Debug("user is not admin", zap.Uint("id", user.ID))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorizedInst.Error()})
			return
		}
		c.Next()
	}
}
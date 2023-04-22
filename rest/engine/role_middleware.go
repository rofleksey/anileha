package engine

import (
	"anileha/db"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

var UserKey = "user"

func AuthorizedMiddleware(log *zap.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		entry := session.Get(UserKey)
		if entry == nil {
			log.Debug("user is not authorized")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorizedInst.Error()})
			return
		}
		user, _ := entry.(*db.AuthUser)

		c.Set(UserKey, user)
		c.Next()
	}
}

func RoleMiddleware(log *zap.Logger, roles []string) func(c *gin.Context) {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		entry := session.Get(UserKey)
		if entry == nil {
			log.Debug("user is not authorized")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorizedInst.Error()})
			return
		}
		user, _ := entry.(*db.AuthUser)
		hasRole := false
	outerRoleLoop:
		for _, role := range roles {
			for _, userRole := range user.Roles {
				if role == userRole {
					hasRole = true
					break outerRoleLoop
				}
			}
		}
		if !hasRole {
			log.Debug("user doesn't have required roles",
				zap.Strings("roles", roles),
				zap.Uint("id", user.ID))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorizedInst.Error()})
			return
		}
		c.Set(UserKey, user)
		c.Next()
	}
}

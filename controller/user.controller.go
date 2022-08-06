package controller

import (
	"anileha/config"
	"anileha/dao"
	"anileha/db"
	"anileha/service"
	"anileha/util"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"net/http"
)

func registerUserController(
	engine *gin.Engine,
	config *config.Config,
	service *service.UserService,
) {
	userGroup := engine.Group("/user")
	userGroup.POST("/register", func(c *gin.Context) {
		var req dao.NewUserRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		confirmId, err := service.RequestRegistration(req.User, req.Pass, req.Email)
		if err != nil {
			_ = c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		link := config.Rest.BaseUrl + "/user/confirm/" + confirmId
		err = service.SendConfirmEmail(req.User, req.Email, link)
		if err != nil {
			_ = c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "OK")
	})
	// TODO: do redirect to login page here
	userGroup.GET("/confirm/:confirmId", func(c *gin.Context) {
		idString := c.Param("confirmId")
		err := service.ConfirmRegistration(idString)
		if err != nil {
			_ = c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "Registration success!")
	})
	userGroup.GET("/me", func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(UserKey)
		if user == nil {
			c.String(http.StatusInternalServerError, "Unauthorized")
		} else {
			c.String(http.StatusOK, user.(*db.AuthUser).Login)
		}
	})
	userGroup.POST("/login", func(c *gin.Context) {
		var req dao.AuthRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user, err := service.GetUserByLogin(req.User)
		if err != nil {
			_ = c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if !util.CheckPasswordHash(req.Pass, config.User.Salt, user.Hash) {
			_ = c.Error(util.ErrInvalidPassword)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.ErrInvalidPassword.Error()})
			return
		}
		session := sessions.Default(c)
		session.Set(UserKey, db.NewAuthUser(*user))
		if err := session.Save(); err != nil {
			_ = c.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": util.ErrSessionSavingFailed.Error()})
			return
		}
		c.String(http.StatusOK, "OK")
	})
	userGroup.POST("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set(UserKey, nil)
		if err := session.Save(); err != nil {
			_ = c.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": util.ErrSessionSavingFailed.Error()})
			return
		}
		c.String(http.StatusOK, "OK")
	})
}

var UserControllerExport = fx.Options(fx.Invoke(registerUserController))

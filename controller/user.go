package controller

import (
	"anileha/config"
	"anileha/dao"
	"anileha/db"
	"anileha/rest"
	"anileha/service"
	"anileha/util"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"net/http"
)

type LoginResponse struct {
	User    string `json:"user"`
	IsAdmin bool   `json:"isAdmin"`
}

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
		err := service.CheckExists(req.User, req.Email)
		if err != nil {
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
		user := session.Get(rest.UserKey)
		if user == nil {
			//if rest.CheckLocalhostAdmin(config, session, user, c) {
			//	c.JSON(http.StatusOK, LoginResponse{config.Admin.Username, true})
			//	return
			//}
			c.Error(rest.ErrUnauthorizedInst)
			return
		}
		authUser := user.(*db.AuthUser)
		c.JSON(http.StatusOK, LoginResponse{authUser.Login, authUser.Admin})
	})
	userGroup.POST("/login", func(c *gin.Context) {
		var req dao.AuthRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(rest.ErrBadRequest(err.Error()))
			return
		}
		user, err := service.GetByLogin(req.User)
		if err != nil {
			_ = c.Error(err)
			return
		}
		if !util.CheckPasswordHash(req.Pass, config.User.Salt, user.Hash) {
			_ = c.Error(rest.ErrInvalidPassword)
			return
		}
		session := sessions.Default(c)
		session.Options(sessions.Options{
			Path:     "/",
			SameSite: http.SameSiteNoneMode,
			Secure:   true,
		})
		session.Set(rest.UserKey, db.NewAuthUser(*user))
		if err := session.Save(); err != nil {
			_ = c.Error(rest.ErrSessionSavingFailed)
			return
		}
		c.JSON(http.StatusOK, LoginResponse{user.Login, user.Admin})
	})
	userGroup.POST("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set(rest.UserKey, nil)
		if err := session.Save(); err != nil {
			_ = c.Error(rest.ErrSessionSavingFailed)
			return
		}
		c.String(http.StatusOK, "OK")
	})
}

var UserControllerExport = fx.Options(fx.Invoke(registerUserController))

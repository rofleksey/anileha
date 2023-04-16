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
	"go.uber.org/zap"
	"net/http"
)

func mapUserToResponse(user db.User) dao.UserResponseDao {
	return dao.UserResponseDao{
		ID:    user.ID,
		Login: user.Login,
		Name:  user.Name,
		Email: user.Email,
		Roles: user.Roles,
		Thumb: user.Thumb.Url,
	}
}

func mapUsersToResponseSlice(users []db.User) []dao.UserResponseDao {
	res := make([]dao.UserResponseDao, 0, len(users))
	for _, u := range users {
		res = append(res, mapUserToResponse(u))
	}
	return res
}

func registerUserController(
	engine *gin.Engine,
	log *zap.Logger,
	config *config.Config,
	fileService *service.FileService,
	thumbService *service.ThumbService,
	userService *service.UserService,
) {
	userGroup := engine.Group("/user")
	userGroup.POST("/register", func(c *gin.Context) {
		var req dao.NewUserRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		err := userService.CheckExists(req.User, req.Email)
		if err != nil {
			c.Error(err)
			return
		}
		confirmId, err := userService.RequestRegistration(req.User, req.Pass, req.Email)
		if err != nil {
			c.Error(err)
			return
		}
		link := config.Rest.BaseUrl + "/user/confirm/" + confirmId
		err = userService.SendConfirmEmail(req.User, req.Email, link)
		if err != nil {
			c.Error(err)
			return
		}
		c.String(http.StatusOK, "OK")
	})
	// TODO: do redirect to login page here
	userGroup.GET("/confirm/:confirmId", func(c *gin.Context) {
		idString := c.Param("confirmId")
		err := userService.ConfirmRegistration(idString)
		if err != nil {
			c.Error(err)
			return
		}
		c.String(http.StatusOK, "Registration success!")
	})
	userGroup.GET("/me", func(c *gin.Context) {
		session := sessions.Default(c)
		sessionUser := session.Get(rest.UserKey)
		if sessionUser == nil {
			c.Error(rest.ErrUnauthorizedInst)
			return
		}
		authUser := sessionUser.(*db.AuthUser)
		user, err := userService.GetById(authUser.ID)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, mapUserToResponse(*user))
	})
	userGroup.POST("/login", func(c *gin.Context) {
		var req dao.AuthRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(rest.ErrBadRequest(err.Error()))
			return
		}
		user, err := userService.GetByLogin(req.User)
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
			Path: "/",
			//SameSite: http.SameSiteNoneMode,
			//Secure:   true,
		})
		session.Set(rest.UserKey, db.NewAuthUser(*user))
		if err := session.Save(); err != nil {
			_ = c.Error(rest.ErrSessionSavingFailed)
			return
		}
		c.JSON(http.StatusOK, mapUserToResponse(*user))
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

	authGroup := engine.Group("/user")
	authGroup.Use(rest.AuthorizedMiddleware(log))
	authGroup.POST("/modify", func(c *gin.Context) {
		var req dao.ModifyUserRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(err)
			return
		}

		authUser := c.MustGet(rest.UserKey).(*db.AuthUser)

		if err := userService.Modify(authUser.ID, req.Name, req.Pass, req.Email); err != nil {
			c.Error(err)
			return
		}

		c.String(http.StatusOK, "OK")
	})

	authGroup.POST("/avatar", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.Error(rest.ErrBadRequest(err.Error()))
			return
		}

		files := form.File["image"]
		if files == nil || len(files) != 1 {
			c.Error(rest.ErrBadRequest("invalid number of images sent"))
			return
		}

		file := files[0]

		tempDst, err := fileService.GenTempFilePath(file.Filename)
		if err != nil {
			c.Error(rest.ErrInternal(err.Error()))
			return
		}

		defer fileService.DeleteTempFileAsync(file.Filename)

		err = c.SaveUploadedFile(file, tempDst)
		if err != nil {
			c.Error(rest.ErrInternal(err.Error()))
			return
		}

		thumb, err := thumbService.CreateFromTempFile(tempDst)
		if err != nil {
			c.Error(err)
			return
		}

		authUser := c.MustGet(rest.UserKey).(*db.AuthUser)
		err = userService.SetThumb(authUser.ID, thumb)
		if err != nil {
			c.Error(err)
			return
		}

		c.String(http.StatusOK, thumb.Url)
	})

	ownerUserGroup := engine.Group("/owner/user")
	ownerUserGroup.Use(rest.RoleMiddleware(log, []string{"owner"}))
	ownerUserGroup.GET("", func(c *gin.Context) {
		userSlice, err := userService.GetAll()
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, mapUsersToResponseSlice(userSlice))
	})

	ownerUserGroup.POST("", func(c *gin.Context) {
		var req dao.OwnerCreateUserRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(err)
			return
		}

		if err := userService.CreateManually(req.Login, req.Pass, req.Email, req.Roles); err != nil {
			c.Error(err)
			return
		}

		c.String(http.StatusOK, "OK")
	})
}

var UserControllerExport = fx.Options(fx.Invoke(registerUserController))

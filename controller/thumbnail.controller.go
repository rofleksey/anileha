package controller

import (
	"anileha/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

func registerThumbController(engine *gin.Engine, fileService *service.FileService, thumbService *service.ThumbService) {
	// TODO: delete thumb
}

var ThumbControllerExport = fx.Options(fx.Invoke(registerThumbController))

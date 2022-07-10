package controller

import (
	"anileha/service"
	"anileha/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"net/http"
	"strconv"
)

func registerThumbnailController(engine *gin.Engine, tempService *service.FileService, thumbnailService *service.ThumbnailService) {
	engine.POST(util.ThumbRoute, func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		tempDst, err := tempService.GetTempFileDst(file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer tempService.DeleteFileAsync(file.Filename)
		err = c.SaveUploadedFile(file, tempDst)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		id, err := thumbnailService.AddThumbnail(file.Filename, tempDst)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, strconv.FormatUint(uint64(id), 10))
	})
}

var ThumbnailControllerExport = fx.Options(fx.Invoke(registerThumbnailController))

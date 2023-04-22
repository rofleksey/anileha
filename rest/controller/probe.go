package controller

import (
	"anileha/config"
	"anileha/db"
	"anileha/ffmpeg/analyze"
	"anileha/rest/dao"
	"anileha/rest/engine"
	"anileha/service"
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gopkg.in/vansante/go-ffprobe.v2"
	"net/http"
	"time"
)

func registerProbeController(
	log *zap.Logger,
	config *config.Config,
	ginEngine *gin.Engine,
	torrentService *service.TorrentService,
	analyzer *analyze.ProbeAnalyzer,
) {
	probeGroup := ginEngine.Group("/admin")
	probeGroup.Use(engine.RoleMiddleware(log, []string{"admin"}))

	probeGroup.POST("/probe", func(c *gin.Context) {
		var req dao.TorrentWithFileIndexRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(engine.ErrBadRequest(err.Error()))
			return
		}
		torrent, err := torrentService.GetById(req.Id)
		if err != nil {
			c.Error(err)
			return
		}
		file := torrent.Files[req.FileIndex]
		if file.Status != db.TorrentFileReady {
			c.Error(engine.ErrFileIsNotReadyToBeConverted)
			return
		}
		if file.ReadyPath == nil {
			c.Error(engine.ErrReadyFileNotFound)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		probe, err := ffprobe.ProbeURL(ctx, *file.ReadyPath)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, probe)
	})

	probeGroup.POST("/subText", func(c *gin.Context) {
		var req dao.TorrentWithFileAndStreamIndexRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(engine.ErrBadRequest(err.Error()))
			return
		}
		torrent, err := torrentService.GetById(req.Id)
		if err != nil {
			c.Error(err)
			return
		}
		file := torrent.Files[req.FileIndex]
		if file.Status != db.TorrentFileReady {
			c.Error(engine.ErrFileIsNotReadyToBeConverted)
			return
		}
		if file.ReadyPath == nil {
			c.Error(engine.ErrReadyFileNotFound)
			return
		}
		result, err := analyzer.ExtractSubText(*file.ReadyPath, req.Stream)
		if err != nil {
			c.Error(err)
			return
		}
		c.String(http.StatusOK, result)
	})
}

var ProbeExport = fx.Options(fx.Invoke(registerProbeController))

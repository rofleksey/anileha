package controller

import (
	"anileha/analyze"
	"anileha/dao"
	"anileha/db"
	"anileha/service"
	"anileha/util"
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gopkg.in/vansante/go-ffprobe.v2"
	"net/http"
	"time"
)

func registerProbeController(
	engine *gin.Engine,
	torrentService *service.TorrentService,
	analyzer *analyze.ProbeAnalyzer,
) {
	probeGroup := engine.Group("/admin/probe")
	probeGroup.Use(AdminMiddleware)
	probeGroup.POST("/", func(c *gin.Context) {
		var req dao.TorrentWithFileIndexRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		torrent, err := torrentService.GetTorrentById(req.TorrentId)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		file := torrent.Files[req.FileIndex]
		if file.Status != db.TORRENT_FILE_READY {
			c.Error(util.ErrFileIsNotReadyToBeConverted)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.ErrFileIsNotReadyToBeConverted.Error()})
			return
		}
		if file.ReadyPath == nil {
			c.Error(util.ErrReadyFileNotFound)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": util.ErrReadyFileNotFound.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		probe, err := ffprobe.ProbeURL(ctx, *file.ReadyPath)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, probe)
	})

	analyzeGroup := engine.Group("/admin/analyze")
	analyzeGroup.Use(AdminMiddleware)
	analyzeGroup.POST("/", func(c *gin.Context) {
		var req dao.TorrentWithFileIndexRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		torrent, err := torrentService.GetTorrentById(req.TorrentId)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		file := torrent.Files[req.FileIndex]
		if file.Status != db.TORRENT_FILE_READY {
			c.Error(util.ErrFileIsNotReadyToBeConverted)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.ErrFileIsNotReadyToBeConverted.Error()})
			return
		}
		if file.ReadyPath == nil {
			c.Error(util.ErrReadyFileNotFound)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": util.ErrReadyFileNotFound.Error()})
			return
		}
		result, err := analyzer.Analyze(*file.ReadyPath, true)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})
}

var ProbeControllerExport = fx.Options(fx.Invoke(registerProbeController))

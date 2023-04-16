package controller

import (
	"anileha/analyze"
	"anileha/config"
	"anileha/dao"
	"anileha/db"
	"anileha/rest"
	"anileha/service"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rofleksey/roflmeta"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gopkg.in/vansante/go-ffprobe.v2"
	"net/http"
	"time"
)

func registerProbeController(
	log *zap.Logger,
	config *config.Config,
	engine *gin.Engine,
	torrentService *service.TorrentService,
	analyzer *analyze.ProbeAnalyzer,
) {
	probeGroup := engine.Group("/admin/probe")
	probeGroup.Use(rest.RoleMiddleware(log, []string{"admin"}))
	probeGroup.POST("/", func(c *gin.Context) {
		var req dao.TorrentWithFileIndexRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(rest.ErrBadRequest(err.Error()))
			return
		}
		torrent, err := torrentService.GetById(req.Id)
		if err != nil {
			c.Error(err)
			return
		}
		file := torrent.Files[req.FileIndex]
		if file.Status != db.TorrentFileReady {
			c.Error(rest.ErrFileIsNotReadyToBeConverted)
			return
		}
		if file.ReadyPath == nil {
			c.Error(rest.ErrReadyFileNotFound)
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

	analyzeGroup := engine.Group("/admin/analyze")
	analyzeGroup.Use(rest.RoleMiddleware(log, []string{"admin"}))
	analyzeGroup.POST("/", func(c *gin.Context) {
		var req dao.TorrentWithFileIndexRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(rest.ErrBadRequest(err.Error()))
			return
		}
		torrent, err := torrentService.GetById(req.Id)
		if err != nil {
			c.Error(err)
			return
		}
		file := torrent.Files[req.FileIndex]
		if file.Status != db.TorrentFileReady {
			c.Error(rest.ErrFileIsNotReadyToBeConverted)
			return
		}
		if file.ReadyPath == nil {
			c.Error(rest.ErrReadyFileNotFound)
			return
		}
		result, err := analyzer.Probe(*file.ReadyPath)
		if err != nil {
			c.Error(err)
			return
		}
		metadata := roflmeta.ParseSingleEpisodeMetadata(file.TorrentPath)
		c.JSON(http.StatusOK, dao.AnalysisResponseDao{
			AnalysisResult: result,
			EpisodeMetadata: dao.EpisodeMetadata{
				Season:  metadata.Season,
				Episode: metadata.Episode,
			},
		})
	})
}

var ProbeControllerExport = fx.Options(fx.Invoke(registerProbeController))

package controller

import (
	"anileha/analyze"
	"anileha/dao"
	"anileha/db"
	"anileha/service"
	"anileha/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"net/http"
	"strconv"
)

// TODO: stop conversions on torrent/story deletion

func mapConversionToResponse(c db.Conversion) dao.ConversionResponseDao {
	return dao.ConversionResponseDao{
		ID:            c.ID,
		SeriesId:      c.SeriesId,
		TorrentFileId: c.TorrentFileId,
		EpisodeId:     c.EpisodeId,
		Name:          c.Name,
		FFmpegCommand: c.Command,
		Status:        c.Status,
		Eta:           c.Progress.Eta,
		Progress:      c.Progress.Progress,
		Elapsed:       c.Progress.Elapsed,
	}
}

func mapConversionsToResponseSlice(conversions []db.Conversion) []dao.ConversionResponseDao {
	res := make([]dao.ConversionResponseDao, 0, len(conversions))
	for _, t := range conversions {
		res = append(res, mapConversionToResponse(t))
	}
	return res
}

func registerConvertController(
	engine *gin.Engine,
	seriesService *service.SeriesService,
	torrentService *service.TorrentService,
	convertService *service.ConversionService,
	analyzer *analyze.ProbeAnalyzer,
) {
	engine.GET("/convert/:id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse id"})
			return
		}
		conversion, err := convertService.GetConversionById(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapConversionToResponse(*conversion))
	})
	engine.GET("/convert/series/:id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse id"})
			return
		}
		conversions, err := convertService.GetConversionsBySeriesId(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapConversionsToResponseSlice(conversions))
	})
	engine.POST("/convert/start", func(c *gin.Context) {
		var req dao.ConvertStartRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		series, err := seriesService.GetSeriesById(req.SeriesId)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		torrentFile, err := torrentService.GetTorrentFileById(req.TorrentFileId)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if torrentFile.Status != db.TORRENT_FILE_READY {
			c.Error(util.ErrFileIsNotReadyToBeConverted)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.ErrFileIsNotReadyToBeConverted.Error()})
			return
		}
		if torrentFile.ReadyPath == nil {
			c.Error(util.ErrFileStateIsCorrupted)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": util.ErrFileStateIsCorrupted.Error()})
			return
		}
		analysis, err := analyzer.Analyze(*torrentFile.ReadyPath, true)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		err = convertService.StartConversion(series, torrentFile, analysis)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "OK")
	})
	engine.POST("/convert/stop", func(c *gin.Context) {
		var req dao.ConvertIdRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := convertService.StopConversion(req.ConversionId)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "OK")
	})
}

var ConvertControllerExport = fx.Options(fx.Invoke(registerConvertController))

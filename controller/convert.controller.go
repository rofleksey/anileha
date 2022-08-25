package controller

import (
	"anileha/analyze"
	"anileha/config"
	"anileha/dao"
	"anileha/db"
	"anileha/service"
	"anileha/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"net/http"
	"strconv"
)

func mapConversionToResponse(c db.Conversion) dao.ConversionResponseDao {
	return dao.ConversionResponseDao{
		ID:            c.ID,
		SeriesId:      c.SeriesId,
		TorrentId:     c.TorrentId,
		TorrentFileId: c.TorrentFileId,
		EpisodeId:     c.EpisodeId,
		EpisodeName:   c.EpisodeName,
		Name:          c.Name,
		FFmpegCommand: c.Command,
		Status:        c.Status,
		Progress:      c.Progress,
		UpdatedAt:     c.UpdatedAt,
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
	config *config.Config,
	engine *gin.Engine,
	torrentService *service.TorrentService,
	convertService *service.ConversionService,
	analyzer *analyze.ProbeAnalyzer,
) {
	convertGroup := engine.Group("/admin/convert")
	convertGroup.Use(AdminMiddleware(config))
	convertGroup.GET("", func(c *gin.Context) {
		conversions, err := convertService.GetAllConversions()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapConversionsToResponseSlice(conversions))
	})
	convertGroup.GET("/:id", func(c *gin.Context) {
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
	convertGroup.GET("/:id/logs", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse id"})
			return
		}
		logs, err := convertService.GetLogsById(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, *logs)
	})
	convertGroup.GET("/series/:id", func(c *gin.Context) {
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
	convertGroup.POST("/start", func(c *gin.Context) {
		var req dao.TorrentWithFileIndicesRequestDao
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
		fileIndices, err := util.ParseFileIndices(req.FileIndices)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		torrentFiles := make([]db.TorrentFile, 0, fileIndices.Length())
		analysisArr := make([]*analyze.Result, 0, fileIndices.Length())
		for _, file := range torrent.Files {
			if fileIndices.Contains(file.EpisodeIndex) {
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
				analysis, err := analyzer.Analyze(*file.ReadyPath, true)
				if err != nil {
					c.Error(err)
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				torrentFiles = append(torrentFiles, file)
				analysisArr = append(analysisArr, analysis)
			}
		}
		err = convertService.StartConversion(*torrent, torrentFiles, analysisArr)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "OK")
	})
	convertGroup.POST("/stop", func(c *gin.Context) {
		var req dao.ConvertIdRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		conversion, err := convertService.GetConversionById(req.ConversionId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if conversion.Status == db.CONVERSION_ERROR || conversion.Status == db.CONVERSION_CANCELLED || conversion.Status == db.CONVERSION_READY {
			c.JSON(http.StatusBadRequest, gin.H{"error": util.ErrAlreadyStopped.Error()})
			return
		}
		err = convertService.StopConversion(req.ConversionId)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "OK")
	})
}

var ConvertControllerExport = fx.Options(fx.Invoke(registerConvertController))

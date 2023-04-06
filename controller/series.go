package controller

import (
	"anileha/config"
	"anileha/dao"
	"anileha/db"
	"anileha/rest"
	"anileha/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

func mapSeriesToResponse(series db.Series) dao.SeriesResponseDao {
	return dao.SeriesResponseDao{
		ID:         series.ID,
		Name:       series.Name,
		Query:      series.Query,
		LastUpdate: series.LastUpdate,
		Thumb:      series.Thumb.DownloadUrl,
	}
}

func mapSeriesToResponseSlice(series []db.Series) []dao.SeriesResponseDao {
	res := make([]dao.SeriesResponseDao, 0, len(series))
	for _, s := range series {
		res = append(res, mapSeriesToResponse(s))
	}
	return res
}

func registerSeriesController(
	config *config.Config,
	log *zap.Logger,
	engine *gin.Engine,
	fileService *service.FileService,
	thumbService *service.ThumbService,
	seriesService *service.SeriesService,
	torrentService *service.TorrentService,
	convertService *service.ConversionService,
	episodeService *service.EpisodeService,
) {
	// TODO: make this GET method
	engine.GET("/series", func(c *gin.Context) {
		seriesSlice, err := seriesService.GetAllSeries()
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, mapSeriesToResponseSlice(seriesSlice))
	})
	engine.POST("/series/search", func(c *gin.Context) {
		var req dao.QueryRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(rest.ErrBadRequest(err.Error()))
			return
		}
		series, err := seriesService.SearchSeries(req.Query)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, mapSeriesToResponseSlice(series))
	})
	engine.GET("/series/:id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.Error(rest.ErrBadRequest(fmt.Sprintf("failed to parse id: %s", err.Error())))
			return
		}
		series, err := seriesService.GetSeriesById(uint(id))
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, mapSeriesToResponse(*series))
	})

	adminSeriesGroup := engine.Group("/admin/series")
	adminSeriesGroup.Use(rest.AdminMiddleware(config))

	adminSeriesGroup.DELETE("/:id", func(c *gin.Context) {
		idString := c.Param("id")
		_, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.Error(rest.ErrBadRequest(fmt.Sprintf("failed to parse id: %s", err.Error())))
			return
		}
		log.Info("not implemented")
		// TODO: delete series
		c.String(http.StatusOK, "OK")
	})
	adminSeriesGroup.POST("/", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.Error(rest.ErrBadRequest(err.Error()))
			return
		}
		names := form.Value["name"]
		if names == nil || len(names) != 1 {
			c.Error(rest.ErrBadRequest("error getting series name"))
			return
		}
		name := names[0]
		trimmedName := strings.TrimSpace(name)
		if len(trimmedName) == 0 {
			c.Error(rest.ErrBadRequest("series name is blank"))
			return
		}
		files := form.File["thumb"]
		if files == nil || len(files) != 1 {
			c.Error(rest.ErrBadRequest("invalid number of files sent"))
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
		thumbId, err := thumbService.AddThumb(file.Filename, tempDst)
		if err != nil {
			c.Error(err)
			return
		}
		seriesId, err := seriesService.AddSeries(name, thumbId)
		if err != nil {
			c.Error(err)
			return
		}
		c.String(http.StatusOK, strconv.FormatUint(uint64(seriesId), 10))
	})
}

var SeriesControllerExport = fx.Options(fx.Invoke(registerSeriesController))

package controller

import (
	"anileha/config"
	"anileha/db"
	"anileha/rest/dao"
	"anileha/rest/engine"
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
	var queryValue *db.SeriesQuery

	if series.Query != nil {
		actualValue := series.Query.Data()
		queryValue = &actualValue
	}

	return dao.SeriesResponseDao{
		ID:         series.ID,
		Title:      series.Title,
		LastUpdate: series.LastUpdate,
		Thumb:      series.Thumb.Url,
		Query:      queryValue,
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
	ginEngine *gin.Engine,
	fileService *service.FileService,
	thumbService *service.ThumbService,
	seriesService *service.SeriesService,
	torrentService *service.TorrentService,
	convertService *service.ConversionService,
	episodeService *service.EpisodeService,
) {
	ginEngine.GET("/series", func(c *gin.Context) {
		seriesSlice, err := seriesService.GetAll()
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, mapSeriesToResponseSlice(seriesSlice))
	})

	ginEngine.POST("/series/search", func(c *gin.Context) {
		var req dao.QueryRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(engine.ErrBadRequest(err.Error()))
			return
		}
		series, err := seriesService.Search(req.Query)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, mapSeriesToResponseSlice(series))
	})

	ginEngine.GET("/series/:id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.Error(engine.ErrBadRequest(fmt.Sprintf("failed to parse id: %s", err.Error())))
			return
		}
		series, err := seriesService.GetById(uint(id))
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, mapSeriesToResponse(*series))
	})

	adminSeriesGroup := ginEngine.Group("/admin/series")
	adminSeriesGroup.Use(engine.RoleMiddleware(log, []string{"admin"}))

	adminSeriesGroup.DELETE("/:id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.Error(engine.ErrBadRequest(fmt.Sprintf("failed to parse id: %s", err.Error())))
			return
		}

		episodes, _ := episodeService.GetBySeriesId(uint(id))

		for _, ep := range episodes {
			_ = episodeService.DeleteById(ep.ID)
		}

		conversions, _ := convertService.GetBySeriesId(uint(id))

		for _, c := range conversions {
			_ = convertService.StopConversion(c.ID)
		}

		torrents, _ := torrentService.GetBySeriesId(uint(id))

		for _, t := range torrents {
			_ = torrentService.Stop(t)
			_ = torrentService.DeleteById(t.ID)
		}

		err = seriesService.DeleteById(uint(id))
		if err != nil {
			c.Error(err)
			return
		}

		c.String(http.StatusOK, "OK")
	})

	adminSeriesGroup.POST("/", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.Error(engine.ErrBadRequest(err.Error()))
			return
		}
		titles := form.Value["title"]
		if titles == nil || len(titles) != 1 {
			c.Error(engine.ErrBadRequest("error getting series title"))
			return
		}
		title := titles[0]
		trimmedTitle := strings.TrimSpace(title)
		if len(trimmedTitle) == 0 {
			c.Error(engine.ErrBadRequest("series title is blank"))
			return
		}
		files := form.File["thumb"]
		if files == nil || len(files) != 1 {
			c.Error(engine.ErrBadRequest("invalid number of files sent"))
			return
		}
		file := files[0]
		tempDst, err := fileService.GenTempFilePath(file.Filename)
		if err != nil {
			c.Error(engine.ErrInternal(err.Error()))
			return
		}
		defer fileService.DeleteTempFileAsync(file.Filename)
		err = c.SaveUploadedFile(file, tempDst)
		if err != nil {
			c.Error(engine.ErrInternal(err.Error()))
			return
		}
		thumb, err := thumbService.CreateFromTempFile(tempDst)
		if err != nil {
			c.Error(err)
			return
		}
		seriesId, err := seriesService.AddSeries(trimmedTitle, thumb)
		if err != nil {
			thumb.Delete()
			c.Error(err)
			return
		}
		c.String(http.StatusOK, strconv.FormatUint(uint64(seriesId), 10))
	})
}

var SeriesExport = fx.Options(fx.Invoke(registerSeriesController))

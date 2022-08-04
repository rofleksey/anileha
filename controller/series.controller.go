package controller

import (
	"anileha/dao"
	"anileha/db"
	"anileha/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"net/http"
	"strconv"
	"strings"
)

func mapSeriesToResponse(series db.Series) dao.SeriesResponseDao {
	return dao.SeriesResponseDao{
		ID:          series.ID,
		Name:        series.Name,
		Description: series.Description,
		Query:       series.Query,
		Thumb:       series.Thumbnail.DownloadUrl,
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
	engine *gin.Engine,
	fileService *service.FileService,
	thumbService *service.ThumbnailService,
	seriesService *service.SeriesService,
) {
	engine.GET("/series", func(c *gin.Context) {
		seriesSlice, err := seriesService.GetAllSeries()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapSeriesToResponseSlice(seriesSlice))
	})
	engine.POST("/series/search", func(c *gin.Context) {
		var req dao.QueryRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		series, err := seriesService.SearchSeries(req.Query)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapSeriesToResponseSlice(series))
	})
	engine.GET("/series/:id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse id"})
			return
		}
		series, err := seriesService.GetSeriesById(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapSeriesToResponse(*series))
	})
	engine.DELETE("/series/:id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse id"})
			return
		}
		err = seriesService.DeleteSeriesById(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "OK")
	})
	engine.POST("/series", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		names := form.Value["name"]
		if names == nil || len(names) != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "error getting series name"})
			return
		}
		name := names[0]
		// TODO: improve trim everywhere (e.g. use regexp)
		trimmedName := strings.Trim(name, " \n")
		if len(trimmedName) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "series name is blank"})
			return
		}
		files := form.File["thumb"]
		if files == nil || len(files) != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid number of files sent"})
			return
		}
		file := files[0]
		tempDst, err := fileService.GenTempFilePath(file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer fileService.DeleteTempFileAsync(file.Filename)
		err = c.SaveUploadedFile(file, tempDst)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		thumbnailId, err := thumbService.AddThumbnail(file.Filename, tempDst)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		seriesId, err := seriesService.AddSeries(name, thumbnailId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, strconv.FormatUint(uint64(seriesId), 10))
	})
}

var SeriesControllerExport = fx.Options(fx.Invoke(registerSeriesController))

package controller

import (
	"anileha/dao"
	"anileha/db"
	"anileha/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"net/http"
	"strconv"
)

func mapEpisodeToResponse(episode db.Episode) dao.EpisodeResponseDao {
	return dao.EpisodeResponseDao{
		ID:           episode.ID,
		ConversionId: episode.ConversionId,
		Name:         episode.Name,
		ThumbnailId:  episode.ThumbnailID,
		Length:       episode.Length,
		DurationSec:  episode.DurationSec,
		Url:          episode.Url,
	}
}

func mapEpisodesToResponseSlice(episodes []db.Episode) []dao.EpisodeResponseDao {
	res := make([]dao.EpisodeResponseDao, 0, len(episodes))
	for _, s := range episodes {
		res = append(res, mapEpisodeToResponse(s))
	}
	return res
}

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

func registerSeriesController(engine *gin.Engine, seriesService *service.SeriesService, episodeService *service.EpisodeService) {
	engine.GET("/series", func(c *gin.Context) {
		seriesSlice, err := seriesService.GetAllSeries()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapSeriesToResponseSlice(seriesSlice))
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
	engine.GET("/series/:id/episodes", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse id"})
			return
		}
		episodes, err := episodeService.GetEpisodesBySeriesId(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapEpisodesToResponseSlice(episodes))
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
		var req dao.SeriesRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id, err := seriesService.AddSeries(req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, strconv.FormatUint(uint64(id), 10))
	})
}

var SeriesControllerExport = fx.Options(fx.Invoke(registerSeriesController))

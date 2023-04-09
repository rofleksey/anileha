package controller

import (
	"anileha/config"
	"anileha/dao"
	"anileha/db"
	"anileha/rest"
	"anileha/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func mapEpisodeToResponse(episode db.Episode) dao.EpisodeResponseDao {
	return dao.EpisodeResponseDao{
		ID:          episode.ID,
		SeriesId:    episode.SeriesId,
		Title:       episode.Title,
		Episode:     episode.Episode,
		Season:      episode.Season,
		CreatedAt:   episode.CreatedAt,
		Thumb:       episode.Thumb.Url,
		Length:      episode.Length,
		DurationSec: episode.DurationSec,
		Url:         episode.Url,
	}
}

func mapEpisodesToResponseSlice(episodes []db.Episode) []dao.EpisodeResponseDao {
	res := make([]dao.EpisodeResponseDao, 0, len(episodes))
	for _, s := range episodes {
		res = append(res, mapEpisodeToResponse(s))
	}
	return res
}

func registerEpisodeController(
	log *zap.Logger,
	config *config.Config,
	engine *gin.Engine,
	episodeService *service.EpisodeService,
) {
	engine.GET("/episodes/series/:seriesId", func(c *gin.Context) {
		idString := c.Param("seriesId")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse id"})
			return
		}
		episodes, err := episodeService.GetBySeriesId(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapEpisodesToResponseSlice(episodes))
	})

	engine.GET("/episodes/:id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse id"})
			return
		}
		episode, err := episodeService.GetById(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapEpisodeToResponse(*episode))
	})

	episodeGroup := engine.Group("/admin/episodes")
	episodeGroup.Use(rest.AdminMiddleware(log, config))

	episodeGroup.POST("refreshThumb", func(c *gin.Context) {
		var req dao.IdRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(rest.ErrBadRequest(err.Error()))
			return
		}
		err := episodeService.RefreshThumb(req.Id)
		if err != nil {
			c.Error(err)
			return
		}
		c.String(http.StatusOK, "OK")
	})

	episodeGroup.DELETE("/:id", func(c *gin.Context) {
		episodeIdString := c.Param("id")
		episodeId, err := strconv.ParseUint(episodeIdString, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse episode id"})
			return
		}
		err = episodeService.DeleteById(uint(episodeId))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "OK")
	})
}

var EpisodeControllerExport = fx.Options(fx.Invoke(registerEpisodeController))

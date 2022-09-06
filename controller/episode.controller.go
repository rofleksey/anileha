package controller

import (
	"anileha/config"
	"anileha/dao"
	"anileha/db"
	"anileha/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"net/http"
	"strconv"
)

func mapEpisodeToResponse(episode db.Episode) dao.EpisodeResponseDao {
	var thumb *string
	if episode.Thumb != nil {
		thumb = &episode.Thumb.Path
	}
	return dao.EpisodeResponseDao{
		ID:           episode.ID,
		SeriesId:     episode.SeriesId,
		ConversionId: episode.ConversionId,
		Name:         episode.Name,
		CreatedAt:    episode.CreatedAt,
		Thumb:        thumb,
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

func registerEpisodeController(
	config *config.Config,
	engine *gin.Engine,
	episodeService *service.EpisodeService,
) {
	engine.GET("/episodes/:id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse id"})
			return
		}
		episode, err := episodeService.GetEpisodeById(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapEpisodeToResponse(*episode))
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

	episodeGroup := engine.Group("/admin/episodes")
	episodeGroup.Use(AdminMiddleware(config))
	episodeGroup.DELETE("/:id", func(c *gin.Context) {
		episodeIdString := c.Param("id")
		episodeId, err := strconv.ParseUint(episodeIdString, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse episode id"})
			return
		}
		err = episodeService.DeleteEpisodeById(uint(episodeId))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "OK")
	})
}

var EpisodeControllerExport = fx.Options(fx.Invoke(registerEpisodeController))

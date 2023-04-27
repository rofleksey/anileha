package controller

import (
	"anileha/config"
	"anileha/db"
	"anileha/rest/dao"
	"anileha/rest/engine"
	"anileha/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
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
	ginEngine *gin.Engine,
	fileService *service.FileService,
	episodeService *service.EpisodeService,
) {
	ginEngine.GET("/episodes", func(c *gin.Context) {
		pageString := c.Param("page")
		page, _ := strconv.Atoi(pageString)
		episodes, maxPages, err := episodeService.GetEpisodes(page)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, dao.GetEpisodesResponseDao{
			Episodes: mapEpisodesToResponseSlice(episodes),
			MaxPages: maxPages,
		})
	})

	ginEngine.GET("/episodes/series/:seriesId", func(c *gin.Context) {
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

	ginEngine.GET("/episodes/:id", func(c *gin.Context) {
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

	adminEpisodeGroup := ginEngine.Group("/admin/episodes")
	adminEpisodeGroup.Use(engine.RoleMiddleware(log, []string{"admin"}))

	adminEpisodeGroup.POST("/", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.Error(engine.ErrBadRequest(err.Error()))
			return
		}

		var seriesId *uint

		seriesIds := form.Value["seriesId"]
		if seriesIds != nil && len(seriesIds) == 1 {
			seriesIdStr := strings.TrimSpace(seriesIds[0])
			if len(seriesIdStr) != 0 {
				seriesIdTemp, err := strconv.ParseUint(seriesIdStr, 10, 64)
				if err == nil {
					seriesIdUint := uint(seriesIdTemp)
					seriesId = &seriesIdUint
				}
			}
		}

		var seasonStr string

		var episodeStr string

		titles := form.Value["title"]
		if titles == nil || len(titles) != 1 {
			c.Error(engine.ErrBadRequest("error getting title"))
			return
		}
		title := strings.TrimSpace(titles[0])
		if len(title) == 0 {
			c.Error(engine.ErrBadRequest("title is blank"))
			return
		}

		seasons := form.Value["season"]
		if seasons != nil && len(seasons) == 1 {
			seasonStr = strings.TrimSpace(seasons[0])
		}

		episodes := form.Value["episode"]
		if episodes != nil && len(episodes) == 1 {
			episodeStr = strings.TrimSpace(episodes[0])
		}

		files := form.File["file"]
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

		if _, err := episodeService.CreateManually(seriesId, tempDst, title, episodeStr, seasonStr); err != nil {
			c.Error(err)
			return
		}

		c.String(http.StatusOK, "OK")
	})

	adminEpisodeGroup.POST("refreshThumb", func(c *gin.Context) {
		var req dao.IdRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(engine.ErrBadRequest(err.Error()))
			return
		}
		err := episodeService.RefreshThumb(req.Id)
		if err != nil {
			c.Error(err)
			return
		}
		c.String(http.StatusOK, "OK")
	})

	adminEpisodeGroup.DELETE("/:id", func(c *gin.Context) {
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

var EpisodeExport = fx.Options(fx.Invoke(registerEpisodeController))

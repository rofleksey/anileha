package controller

import (
	"anileha/command"
	"anileha/config"
	"anileha/dao"
	"anileha/db"
	"anileha/rest"
	"anileha/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
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
		Command:       c.Command,
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
	log *zap.Logger,
	config *config.Config,
	engine *gin.Engine,
	torrentService *service.TorrentService,
	convertService *service.ConversionService,
) {
	convertGroup := engine.Group("/admin/convert")
	convertGroup.Use(rest.RoleMiddleware(log, []string{"admin"}))
	convertGroup.GET("", func(c *gin.Context) {
		conversions, err := convertService.GetAll()
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, mapConversionsToResponseSlice(conversions))
	})
	convertGroup.GET("/:id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.Error(rest.ErrBadRequest("failed to parse id"))
			return
		}
		conversion, err := convertService.GetById(uint(id))
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, mapConversionToResponse(*conversion))
	})
	convertGroup.GET("/:id/logs", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.Error(rest.ErrBadRequest("failed to parse id"))
			return
		}
		logs, err := convertService.GetLogsById(uint(id))
		if err != nil {
			c.Error(err)
			return
		}
		if logs == nil {
			c.Error(rest.ErrInternal("logs are nil"))
			return
		}
		c.String(http.StatusOK, string(logs))
	})
	convertGroup.GET("/series/:id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.Error(rest.ErrBadRequest("failed to parse id"))
			return
		}
		conversions, err := convertService.GetBySeriesId(uint(id))
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, mapConversionsToResponseSlice(conversions))
	})
	convertGroup.POST("/start", func(c *gin.Context) {
		var req dao.StartConversionRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(rest.ErrBadRequest(err.Error()))
			return
		}
		torrent, err := torrentService.GetByIdWithSeries(req.TorrentId)
		if err != nil {
			c.Error(err)
			return
		}
		torrentFiles := make([]db.TorrentFile, 0, 32)
		prefsArr := make([]command.Preferences, 0, 32)
		for _, file := range torrent.Files {
			reqIndex := slices.IndexFunc(req.Files, func(data dao.StartConversionFilePrefData) bool {
				return data.Index == file.ClientIndex
			})
			if reqIndex >= 0 {
				reqFile := req.Files[reqIndex]
				if file.Status != db.TorrentFileReady {
					c.Error(rest.ErrFileIsNotReadyToBeConverted)
					return
				}
				if file.ReadyPath == nil {
					c.Error(rest.ErrReadyFileNotFound)
					return
				}
				torrentFiles = append(torrentFiles, file)
				prefsArr = append(prefsArr, command.Preferences{
					Audio: command.PreferencesData{
						Disable:      reqFile.Audio.Disable,
						ExternalFile: reqFile.Audio.File,
						StreamIndex:  reqFile.Audio.Stream,
						Lang:         reqFile.Audio.Lang,
					},
					Sub: command.PreferencesData{
						Disable:      reqFile.Sub.Disable,
						ExternalFile: reqFile.Sub.File,
						StreamIndex:  reqFile.Sub.Stream,
						Lang:         reqFile.Sub.Lang,
					},
					Episode: reqFile.Episode,
					Season:  reqFile.Season,
				})
			}
		}
		err = convertService.StartConversion(*torrent, torrentFiles, prefsArr)
		if err != nil {
			c.Error(err)
			return
		}
		c.String(http.StatusOK, "OK")
	})
	convertGroup.POST("/stop", func(c *gin.Context) {
		var req dao.IdRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(rest.ErrBadRequest(err.Error()))
			return
		}
		conversion, err := convertService.GetById(req.Id)
		if err != nil {
			c.Error(err)
			return
		}
		if conversion.Status == db.ConversionError || conversion.Status == db.ConversionCancelled || conversion.Status == db.ConversionReady {
			c.Error(rest.ErrAlreadyStopped)
			return
		}
		err = convertService.StopConversion(req.Id)
		if err != nil {
			c.Error(err)
			return
		}
		c.String(http.StatusOK, "OK")
	})
}

var ConvertControllerExport = fx.Options(fx.Invoke(registerConvertController))

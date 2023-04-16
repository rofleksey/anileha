package controller

import (
	"anileha/config"
	"anileha/dao"
	"anileha/db"
	"anileha/rest"
	"anileha/service"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func mapTorrentFilesToResponse(torrentFiles []db.TorrentFile) []dao.TorrentFileResponseDao {
	res := make([]dao.TorrentFileResponseDao, 0, len(torrentFiles))
	for _, f := range torrentFiles {
		res = append(res, dao.TorrentFileResponseDao{
			Path:        f.TorrentPath,
			Status:      f.Status,
			Selected:    f.Selected,
			Length:      f.Length,
			ClientIndex: f.ClientIndex,
		})
	}
	return res
}

func mapTorrentToResponse(torrent db.Torrent) dao.TorrentResponseDao {
	return dao.TorrentResponseDao{
		ID:                  torrent.ID,
		Name:                torrent.Name,
		Status:              torrent.Status,
		Source:              torrent.Source,
		TotalLength:         torrent.TotalLength,
		TotalDownloadLength: torrent.TotalDownloadLength,
		Progress:            torrent.Progress,
		BytesRead:           torrent.BytesRead,
		Files:               mapTorrentFilesToResponse(torrent.Files),
		UpdatedAt:           torrent.UpdatedAt,
	}
}

func mapTorrentWithoutFilesToResponse(torrent db.Torrent) dao.TorrentResponseWithoutFilesDao {
	return dao.TorrentResponseWithoutFilesDao{
		ID:                  torrent.ID,
		Name:                torrent.Name,
		Status:              torrent.Status,
		Source:              torrent.Source,
		TotalLength:         torrent.TotalLength,
		TotalDownloadLength: torrent.TotalDownloadLength,
		Progress:            torrent.Progress,
		BytesRead:           torrent.BytesRead,
		UpdatedAt:           torrent.UpdatedAt,
	}
}

func mapTorrentsWithoutFilesToResponseSlice(torrents []db.Torrent) []dao.TorrentResponseWithoutFilesDao {
	res := make([]dao.TorrentResponseWithoutFilesDao, 0, len(torrents))
	for _, t := range torrents {
		res = append(res, mapTorrentWithoutFilesToResponse(t))
	}
	return res
}

func registerTorrentController(
	log *zap.Logger,
	config *config.Config,
	engine *gin.Engine,
	fileService *service.FileService,
	torrentService *service.TorrentService,
) {
	torrentGroup := engine.Group("/admin/torrent")
	torrentGroup.Use(rest.RoleMiddleware(log, []string{"admin"}))

	torrentGroup.GET("", func(c *gin.Context) {
		torrentsSlice, err := torrentService.GetAll()
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, mapTorrentsWithoutFilesToResponseSlice(torrentsSlice))
	})
	torrentGroup.GET(":id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.Error(rest.ErrBadRequest(err.Error()))
			return
		}
		torrent, err := torrentService.GetById(uint(id))
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, mapTorrentToResponse(*torrent))
	})
	torrentGroup.GET("series/:id", func(c *gin.Context) {
		seriesIdString := c.Param("id")
		id, err := strconv.ParseUint(seriesIdString, 10, 64)
		if err != nil {
			c.Error(rest.ErrBadRequest(err.Error()))
			return
		}
		torrents, err := torrentService.GetBySeriesId(uint(id))
		if err != nil {
			log.Info("error occurred")
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, mapTorrentsWithoutFilesToResponseSlice(torrents))
	})
	torrentGroup.POST("start", func(c *gin.Context) {
		var req dao.StartTorrentRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(rest.ErrBadRequest(err.Error()))
			return
		}
		torrent, err := torrentService.GetById(req.Id)
		if err != nil {
			c.Error(err)
			return
		}
		if torrent.Status == db.TorrentDownloading {
			c.Error(rest.ErrAlreadyStarted)
			return
		}
		err = torrentService.Start(*torrent, req.FileIndices)
		if err != nil {
			c.Error(err)
			return
		}
		c.String(http.StatusOK, "OK")
	})
	torrentGroup.POST("stop", func(c *gin.Context) {
		var req dao.IdRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(rest.ErrBadRequest(err.Error()))
			return
		}
		torrent, err := torrentService.GetById(req.Id)
		if err != nil {
			c.Error(err)
			return
		}
		if torrent.Status != db.TorrentDownloading {
			c.String(http.StatusOK, "Already stopped")
			return
		}
		err = torrentService.Stop(*torrent)
		if err != nil {
			c.Error(err)
			return
		}
		c.String(http.StatusOK, "OK")
	})
	torrentGroup.DELETE(":id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.Error(rest.ErrBadRequest(err.Error()))
			return
		}
		err = torrentService.DeleteById(uint(id))
		if err != nil {
			c.Error(err)
			return
		}
		c.String(http.StatusOK, "OK")
	})
	torrentGroup.POST("", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.Error(rest.ErrBadRequest(err.Error()))
			return
		}

		seriesIdStrArr := form.Value["seriesId"]
		if seriesIdStrArr == nil || len(seriesIdStrArr) != 1 {
			c.Error(rest.ErrBadRequest("error getting seriesId"))
			return
		}
		seriesId, err := strconv.ParseUint(seriesIdStrArr[0], 10, 64)
		if err != nil {
			c.Error(rest.ErrBadRequest("failed to parse seriesId"))
			return
		}

		var auto *db.AutoTorrent

		autoArr := form.Value["auto"]
		if autoArr != nil && len(autoArr) == 1 {
			if err := json.Unmarshal([]byte(autoArr[0]), &auto); err == nil {
				if auto.AudioLang == "" || auto.SubLang == "" {
					c.Error(rest.ErrBadRequest("invalid auto JSON"))
					return
				}
			}
		}

		files := form.File["file"]
		if files == nil || len(files) == 0 {
			c.Error(rest.ErrBadRequest("no files sent"))
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
		err = torrentService.AddFromFile(uint(seriesId), tempDst, auto)
		if err != nil {
			c.Error(err)
			return
		}
		c.String(http.StatusOK, "OK")
	})
}

var TorrentControllerExport = fx.Options(fx.Invoke(registerTorrentController))

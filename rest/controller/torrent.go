package controller

import (
	"anileha/config"
	"anileha/db"
	"anileha/rest/dao"
	"anileha/rest/engine"
	"anileha/search/nyaa"
	"anileha/service"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strconv"
)

func mapTorrentFilesToResponse(torrentFiles []db.TorrentFile) []dao.TorrentFileResponseDao {
	res := make([]dao.TorrentFileResponseDao, 0, len(torrentFiles))
	for _, f := range torrentFiles {
		res = append(res, dao.TorrentFileResponseDao{
			Path:              f.TorrentPath,
			Status:            f.Status,
			Selected:          f.Selected,
			Length:            f.Length,
			ClientIndex:       f.ClientIndex,
			Type:              f.Type,
			SuggestedMetadata: f.SuggestedMetadata.Data(),
			Analysis:          f.Analysis.Data(),
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
	ginEngine *gin.Engine,
	fileService *service.FileService,
	nyaaService *nyaa.Service,
	torrentService *service.TorrentService,
) {
	torrentGroup := ginEngine.Group("/admin/torrent")
	torrentGroup.Use(engine.RoleMiddleware(log, []string{"admin"}))

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
			c.Error(engine.ErrBadRequest(err.Error()))
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
			c.Error(engine.ErrBadRequest(err.Error()))
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
			c.Error(engine.ErrBadRequest(err.Error()))
			return
		}
		torrent, err := torrentService.GetById(req.Id)
		if err != nil {
			c.Error(err)
			return
		}
		if torrent.Status == db.TorrentDownload || torrent.Status == db.TorrentAnalysis {
			c.Error(engine.ErrAlreadyStarted)
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
			c.Error(engine.ErrBadRequest(err.Error()))
			return
		}
		torrent, err := torrentService.GetById(req.Id)
		if err != nil {
			c.Error(err)
			return
		}
		if torrent.Status != db.TorrentDownload {
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
			c.Error(engine.ErrBadRequest(err.Error()))
			return
		}
		err = torrentService.DeleteById(uint(id))
		if err != nil {
			c.Error(err)
			return
		}
		c.String(http.StatusOK, "OK")
	})

	torrentGroup.POST("/fromFile", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.Error(engine.ErrBadRequest(err.Error()))
			return
		}

		seriesIdStrArr := form.Value["seriesId"]
		if seriesIdStrArr == nil || len(seriesIdStrArr) != 1 {
			c.Error(engine.ErrBadRequest("error getting seriesId"))
			return
		}
		seriesId, err := strconv.ParseUint(seriesIdStrArr[0], 10, 64)
		if err != nil {
			c.Error(engine.ErrBadRequest("failed to parse seriesId"))
			return
		}

		var auto *db.AutoTorrent

		autoArr := form.Value["auto"]
		if autoArr != nil && len(autoArr) == 1 {
			if err := json.Unmarshal([]byte(autoArr[0]), &auto); err == nil {
				if auto.AudioLang == "" || auto.SubLang == "" {
					c.Error(engine.ErrBadRequest("invalid auto JSON"))
					return
				}
			}
		}

		files := form.File["file"]
		if files == nil || len(files) == 0 {
			c.Error(engine.ErrBadRequest("no files sent"))
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
		err = torrentService.AddFromFile(uint(seriesId), tempDst, auto)
		if err != nil {
			c.Error(err)
			return
		}
		c.String(http.StatusOK, "OK")
	})

	torrentGroup.POST("/fromSearch", func(c *gin.Context) {
		var req dao.AddTorrentFromSearchRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(engine.ErrBadRequest(err.Error()))
			return
		}

		tempDst, err := fileService.GenTempFilePath("new.torrent")
		if err != nil {
			c.Error(engine.ErrInternal(err.Error()))
			return
		}

		bytes, err := nyaaService.DownloadById(c.Request.Context(), req.TorrentID)
		if err != nil {
			c.Error(engine.ErrInternal(err.Error()))
			return
		}

		err = os.WriteFile(tempDst, bytes, 0644)
		if err != nil {
			c.Error(engine.ErrInternal(err.Error()))
			return
		}

		err = torrentService.AddFromFile(req.SeriesID, tempDst, req.Auto)
		if err != nil {
			c.Error(err)
			return
		}

		c.String(http.StatusOK, "OK")
	})
}

var TorrentExport = fx.Options(fx.Invoke(registerTorrentController))

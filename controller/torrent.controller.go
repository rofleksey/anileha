package controller

import (
	"anileha/config"
	"anileha/dao"
	"anileha/db"
	"anileha/service"
	"anileha/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"net/http"
	"strconv"
)

func mapTorrentFilesToResponse(torrentFiles []db.TorrentFile) []dao.TorrentFileResponseDao {
	res := make([]dao.TorrentFileResponseDao, 0, len(torrentFiles))
	for _, f := range torrentFiles {
		res = append(res, dao.TorrentFileResponseDao{
			Path:         f.TorrentPath,
			Status:       f.Status,
			Selected:     f.Selected,
			Length:       f.Length,
			Episode:      f.Episode,
			EpisodeIndex: f.EpisodeIndex,
			Season:       f.Season,
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
		Auto:                torrent.Auto,
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
		Auto:                torrent.Auto,
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
	config *config.Config,
	engine *gin.Engine,
	fileService *service.FileService,
	torrentService *service.TorrentService,
	pipelineFacade *service.PipelineFacade,
) {
	torrentGroup := engine.Group("/admin/torrent")
	torrentGroup.Use(AdminMiddleware(config))

	torrentGroup.GET("", func(c *gin.Context) {
		torrentsSlice, err := torrentService.GetAllTorrents()
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapTorrentsWithoutFilesToResponseSlice(torrentsSlice))
	})
	torrentGroup.GET(":id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		torrent, err := torrentService.GetTorrentById(uint(id))
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapTorrentToResponse(*torrent))
	})
	torrentGroup.GET("series/:id", func(c *gin.Context) {
		seriesIdString := c.Param("id")
		id, err := strconv.ParseUint(seriesIdString, 10, 64)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		torrents, err := torrentService.GetTorrentsBySeriesId(uint(id))
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapTorrentsWithoutFilesToResponseSlice(torrents))
	})
	torrentGroup.POST("start", func(c *gin.Context) {
		var req dao.TorrentWithFileIndicesRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fileIndices, err := util.ParseFileIndices(req.FileIndices)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		torrent, err := torrentService.GetTorrentById(req.TorrentId)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if torrent.Status == db.TORRENT_DOWNLOADING {
			c.Error(util.ErrAlreadyStarted)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.ErrAlreadyStarted.Error()})
			return
		}
		err = torrentService.StartTorrent(*torrent, fileIndices)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "OK")
	})
	torrentGroup.POST("stop", func(c *gin.Context) {
		var req dao.TorrentIdRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		torrent, err := torrentService.GetTorrentById(req.TorrentId)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if torrent.Status != db.TORRENT_DOWNLOADING {
			c.Error(util.ErrAlreadyStopped)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.ErrAlreadyStopped.Error()})
			return
		}
		err = torrentService.StopTorrent(*torrent)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "OK")
	})
	torrentGroup.DELETE(":id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		resultChan := make(chan error, 1)
		pipelineFacade.Channel <- service.PipelineMessageDeleteTorrent{
			TorrentId: uint(id),
			Result:    resultChan,
		}
		err = <-resultChan
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "OK")
	})
	torrentGroup.POST("", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		seriesIdStrArr := form.Value["seriesId"]
		if seriesIdStrArr == nil || len(seriesIdStrArr) != 1 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "error getting seriesId"})
			return
		}
		auto := false
		autoStr := form.Value["auto"]
		if autoStr != nil && len(autoStr) == 1 {
			auto, _ = strconv.ParseBool(autoStr[0])
		}
		seriesId, err := strconv.ParseUint(seriesIdStrArr[0], 10, 64)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to parse seriesId"})
			return
		}
		files := form.File["files"]
		if files == nil || len(files) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no files sent"})
			return
		}
		for _, file := range files {
			tempDst, err := fileService.GenTempFilePath(file.Filename)
			if err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			// it's okay :>
			defer fileService.DeleteTempFileAsync(file.Filename)
			err = c.SaveUploadedFile(file, tempDst)
			if err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			id, err := torrentService.AddTorrentFromFile(uint(seriesId), tempDst, auto)
			if err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if auto {
				torrent, err := torrentService.GetTorrentById(id)
				if err != nil {
					c.Error(err)
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				allIndices, _ := util.ParseFileIndices("*")
				err = torrentService.StartTorrent(*torrent, allIndices)
				if err != nil {
					c.Error(err)
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
			}
		}
		c.String(http.StatusOK, "OK")
	})
}

var TorrentControllerExport = fx.Options(fx.Invoke(registerTorrentController))

package controller

import (
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
		Files:               mapTorrentFilesToResponse(torrent.Files),
	}
}

func mapTorrentsToResponseSlice(torrents []db.Torrent) []dao.TorrentResponseDao {
	res := make([]dao.TorrentResponseDao, 0, len(torrents))
	for _, t := range torrents {
		res = append(res, mapTorrentToResponse(t))
	}
	return res
}

func registerTorrentController(engine *gin.Engine, fileService *service.FileService, torrentService *service.TorrentService) {
	engine.GET("/torrent", func(c *gin.Context) {
		torrentsSlice, err := torrentService.GetAllTorrents()
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapTorrentsToResponseSlice(torrentsSlice))
	})
	engine.GET("/torrent/:id", func(c *gin.Context) {
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
	engine.POST("/torrent/start", func(c *gin.Context) {
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
	engine.POST("/torrent/stop", func(c *gin.Context) {
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
	engine.DELETE("/torrent/:id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = torrentService.DeleteTorrentById(uint(id))
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "OK")
	})
	engine.POST("/torrent", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		seriesIdStrArr := form.Value["seriesId"]
		if seriesIdStrArr == nil || len(seriesIdStrArr) != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "error getting seriesId"})
			return
		}
		seriesId, err := strconv.ParseUint(seriesIdStrArr[0], 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse seriesId"})
			return
		}
		files := form.File["file"]
		if files == nil || len(files) != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid number of files sent"})
			return
		}
		file := files[0]
		tempDst, err := fileService.GenTempFilePath(file.Filename)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer fileService.DeleteTempFileAsync(file.Filename)
		err = c.SaveUploadedFile(file, tempDst)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id, err := torrentService.AddTorrentFromFile(uint(seriesId), tempDst)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, strconv.FormatUint(uint64(id), 10))
	})
}

var TorrentControllerExport = fx.Options(fx.Invoke(registerTorrentController))

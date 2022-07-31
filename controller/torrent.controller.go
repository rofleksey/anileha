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

func mapTorrentFilesToResponse(torrentFiles []*db.TorrentFile) []dao.TorrentFileResponseDao {
	res := make([]dao.TorrentFileResponseDao, 0, len(torrentFiles))
	for _, f := range torrentFiles {
		res = append(res, dao.TorrentFileResponseDao{
			Path:     f.TorrentPath,
			Status:   f.Status,
			Selected: f.Selected,
			Length:   f.Length,
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
		Files:               mapTorrentFilesToResponse(torrent.Files),
	}
}

func mapTorrentWithProgressToResponse(torrent db.TorrentWithProgress) dao.TorrentResponseDao {
	return dao.TorrentResponseDao{
		ID:                  torrent.ID,
		Name:                torrent.Name,
		Status:              torrent.Status,
		Source:              torrent.Source,
		TotalLength:         torrent.TotalLength,
		TotalDownloadLength: torrent.TotalDownloadLength,
		Progress:            &torrent.Progress,
		BytesRead:           &torrent.BytesRead,
		BytesMissing:        &torrent.BytesMissing,
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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapTorrentsToResponseSlice(torrentsSlice))
	})
	engine.GET("/torrent/:id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse id"})
			return
		}
		torrent, err := torrentService.GetTorrentById(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, mapTorrentWithProgressToResponse(*torrent))
	})
	engine.DELETE("/torrent/:id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse id"})
			return
		}
		err = torrentService.DeleteTorrentById(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "OK")
	})
	engine.POST("/torrent/file", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		fileIndicesArr := form.Value["fileIndices"]
		if fileIndicesArr == nil || len(fileIndicesArr) != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "error getting fileIndices"})
			return
		}
		fileIndices, err := util.ParseFileIndices(fileIndicesArr[0])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer fileService.DeleteTempFileAsync(file.Filename)
		err = c.SaveUploadedFile(file, tempDst)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		id, err := torrentService.AddTorrentFromFile(uint(seriesId), tempDst, fileIndices)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, strconv.FormatUint(uint64(id), 10))
	})
}

var TorrentControllerExport = fx.Options(fx.Invoke(registerTorrentController))

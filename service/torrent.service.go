package service

import (
	"anileha/config"
	"anileha/db"
	"anileha/util"
	"context"
	"errors"
	torrentLib "github.com/anacrolix/torrent"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type TorrentService struct {
	db              *gorm.DB
	client          *torrentLib.Client
	cTorrentMap     sync.Map // cTorrentMap Stores torrentLib.Client torrent entries [uint -> *torrentLib.Torrent]
	fileService     *FileService
	log             *zap.Logger
	infoFolder      string
	downloadsFolder string
	readyFolder     string
}

func createDirs(config *config.Config) (string, string, string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", "", "", err
	}
	infoFolder := path.Join(workingDir, config.Data.Dir, util.TorrentInfoSubDir)
	err = os.MkdirAll(infoFolder, os.ModePerm)
	if err != nil {
		return "", "", "", err
	}

	downloadsFolder := path.Join(workingDir, config.Data.Dir, util.TorrentDownloadsSubDir)
	err = os.MkdirAll(downloadsFolder, os.ModePerm)
	if err != nil {
		return "", "", "", err
	}

	readyFolder := path.Join(workingDir, config.Data.Dir, util.TorrentReadySubDir)
	err = os.MkdirAll(readyFolder, os.ModePerm)
	if err != nil {
		return "", "", "", err
	}

	return infoFolder, downloadsFolder, readyFolder, nil
}

func NewTorrentService(lifecycle fx.Lifecycle, db *gorm.DB, log *zap.Logger, config *config.Config, fileService *FileService) (*TorrentService, error) {
	infoFolder, downloadsFolder, readyFolder, err := createDirs(config)
	if err != nil {
		return nil, err
	}
	clientConfig := torrentLib.NewDefaultClientConfig()
	clientConfig.DataDir = downloadsFolder
	client, err := torrentLib.NewClient(clientConfig)
	lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			client.Close()
			return nil
		},
	})
	return &TorrentService{
		db:              db,
		client:          client,
		fileService:     fileService,
		log:             log,
		infoFolder:      infoFolder,
		downloadsFolder: downloadsFolder,
		readyFolder:     readyFolder,
	}, nil
}

func (s *TorrentService) GetTorrentById(id uint) (*db.TorrentWithProgress, error) {
	var torrent db.Torrent
	queryResult := s.db.Preload("Files").First(&torrent, "id = ?", id)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, util.ErrNotFound
	}
	mapValue, exists := s.cTorrentMap.Load(torrent.ID)
	if !exists {
		return &db.TorrentWithProgress{
			Torrent:      torrent,
			Progress:     0,
			BytesRead:    0,
			BytesMissing: 0,
		}, nil
	}
	cTorrent := mapValue.(*torrentLib.Torrent)
	bytesTotal := torrent.TotalDownloadLength
	cFiles := cTorrent.Files()
	bytesRead := int64(0)
	for i, file := range torrent.Files {
		cFile := cFiles[i]
		if file.Selected {
			bytesRead += cFile.BytesCompleted()
		}
	}
	bytesMissing := bytesTotal - bytesRead
	var progress float64
	if bytesTotal > 0 {
		progress = float64(bytesRead) / float64(bytesTotal)
	} else {
		progress = 0
	}
	return &db.TorrentWithProgress{
		Torrent:      torrent,
		Progress:     progress,
		BytesRead:    bytesRead,
		BytesMissing: bytesMissing,
	}, nil
}

func (s *TorrentService) GetTorrentFileById(id uint) (*db.TorrentFile, error) {
	var file db.TorrentFile
	queryResult := s.db.First(&file, "id = ?", id)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, util.ErrNotFound
	}
	return &file, nil
}

func (s *TorrentService) GetAllTorrents() ([]db.Torrent, error) {
	var torrentArr []db.Torrent
	queryResult := s.db.Find(&torrentArr)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return torrentArr, nil
}

func (s *TorrentService) DeleteTorrentById(id uint) error {
	var torrent db.Torrent
	queryResult := s.db.Preload("Files").First(&torrent, "id = ?", id)
	if queryResult.Error != nil {
		return queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return util.ErrNotFound
	}
	if torrent.Status == db.TORRENT_DOWNLOADING {
		return util.ErrDeleteStartedTorrent
	}
	queryResult = s.db.Delete(&db.Torrent{}, id)
	if queryResult.Error != nil {
		return queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return util.ErrNotFound
	}
	go s.cleanUpTorrent(torrent)
	return nil
}

// cleanUpTorrent Drops cTorrent, removes all torrent files
func (s *TorrentService) cleanUpTorrent(torrent db.Torrent) {
	mapValue, exists := s.cTorrentMap.LoadAndDelete(torrent.ID)

	if exists {
		cTorrent := mapValue.(*torrentLib.Torrent)
		cTorrent.Drop()
		<-cTorrent.Closed()
		for _, cFile := range cTorrent.Files() {
			cFilePath := path.Join(s.downloadsFolder, cFile.DisplayPath())
			err := os.RemoveAll(cFilePath)
			if err != nil {
				s.log.Warn("failed to cleanup torrent file", zap.String("path", cFilePath), zap.Error(err))
			}
		}
	}

	if torrent.InfoType == db.TORRENT_INFO_FILE {
		deleteErr := os.Remove(torrent.InfoPath)
		if deleteErr != nil {
			s.log.Warn("failed to cleanup torrent info", zap.String("path", torrent.InfoPath), zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name), zap.Error(deleteErr))
		}
	}
}

// onFailedImport Logs error, deletes torrent from DB, cleans up torrent files
func (s *TorrentService) onFailedImport(torrent db.Torrent, err error) {
	s.log.Warn("failed to import torrent", zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name), zap.Error(err))
	deleteErr := s.DeleteTorrentById(torrent.ID)
	if deleteErr != nil {
		s.log.Warn("failed to delete torrent DB entry", zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name), zap.Error(deleteErr))
	}
	s.cleanUpTorrent(torrent)
}

// onTorrentCompletion Creates READY folder, moves torrent files into it, updates DB entries
func (s *TorrentService) onTorrentCompletion(id uint) {
	var torrent db.Torrent
	queryResult := s.db.Preload("Files").First(&torrent, "id = ?", id)
	if queryResult.Error != nil {
		s.log.Error("failed to complete torrent", zap.Uint("torrentId", torrent.ID), zap.Error(queryResult.Error))
		return
	}
	if queryResult.RowsAffected == 0 {
		s.log.Error("failed to complete torrent", zap.Uint("torrentId", torrent.ID), zap.Error(errors.New("len(rows) == 0")))
		return
	}

	torrentIdStr := strconv.FormatUint(uint64(torrent.ID), 10)
	torrentReadyRootFolder := path.Join(s.readyFolder, torrentIdStr)

	for _, file := range torrent.Files {
		// ignore already downloaded files
		if file.ReadyPath != nil {
			continue
		}

		oldPath := path.Join(s.downloadsFolder, torrent.Name, file.TorrentPath)

		// if file is not selected - delete it and continue
		if !file.Selected {
			_ = os.Remove(oldPath)
			//if deleteErr != nil {
			//	s.log.Warn("failed to delete unselected file", zap.String("path", oldPath), zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name), zap.Error(deleteErr))
			//}
			continue
		}

		torrentFileName := filepath.Base(file.TorrentPath)
		newPath, err := s.fileService.GenFilePath(torrentReadyRootFolder, torrentFileName)
		if err != nil {
			s.log.Error("failed to acquire temp file", zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name), zap.Error(err))
			return
		}

		newPathDir := filepath.Dir(newPath)
		createDirErr := os.MkdirAll(newPathDir, os.ModePerm)
		if createDirErr != nil {
			s.log.Error("failed to create dirs", zap.String("path", newPathDir), zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name), zap.Error(err))
			return
		}

		err = os.Rename(oldPath, newPath)
		if err != nil {
			s.log.Error("failed to move ready torrent file", zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name), zap.Error(err))
			return
		}
		file.ReadyPath = &newPath
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// update torrent status
		if err := tx.Model(&db.Torrent{}).Where("id = ?", id).Updates(db.Torrent{Status: db.TORRENT_READY}).Error; err != nil {
			return err
		}

		// update downloaded files' statuses
		for _, file := range torrent.Files {
			if file.Selected && file.Status != db.TORRENT_FILE_READY {
				if err := tx.Model(&db.TorrentFile{}).Where("id = ?", file.ID).Updates(db.TorrentFile{Status: db.TORRENT_FILE_READY, ReadyPath: file.ReadyPath}).Error; err != nil {
					return err
				}
			}
		}

		// apply transaction
		return nil
	})
	if err != nil {
		s.log.Error("transaction on torrent completion failed", zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name), zap.Error(err))
		return
	}

	s.log.Info("torrent finished", zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name))
}

// torrentCompletionWatcher Polls for torrent's completion, calls onTorrentCompletion
func (s *TorrentService) torrentCompletionWatcher(id uint, name string, files []db.TorrentFile, totalDownloadLength int64, cTorrent *torrentLib.Torrent) {
	ticker := time.NewTicker(3 * time.Second)
	cFiles := cTorrent.Files()
	for {
		select {
		case <-cTorrent.Closed():
			s.log.Info("torrentCompletionWatcher exited", zap.Uint("torrentId", id), zap.String("torrentName", name), zap.String("cause", "closed"))
			return
		case <-ticker.C:
			bytesRead := int64(0)
			for i, file := range files {
				cFile := cFiles[i]
				if file.Selected {
					bytesRead += cFile.BytesCompleted()
				}
			}
			if bytesRead < totalDownloadLength {
				continue
			}
			ticker.Stop()

			cTorrent.Drop()
			s.cTorrentMap.Delete(id)
			<-cTorrent.Closed()

			s.onTorrentCompletion(id)
			return
		}
	}
}

func (s *TorrentService) initTorrent(torrent db.Torrent, fileIndices map[uint]struct{}) {
	var cTorrent *torrentLib.Torrent
	var err error
	switch torrent.InfoType {
	case db.TORRENT_INFO_FILE:
		cTorrent, err = s.client.AddTorrentFromFile(torrent.InfoPath)
	case db.TORRENT_INFO_MAGNET:
		cTorrent, err = s.client.AddMagnet(torrent.InfoPath)
	default:
		err = util.ErrInvalidInfoType
	}
	if err != nil {
		s.onFailedImport(torrent, err)
		return
	}

	<-cTorrent.GotInfo()
	info := cTorrent.Info()

	torrent.Name = info.BestName()
	torrent.Status = db.TORRENT_DOWNLOADING
	s.log.Info("torrent received info", zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name))

	files := make([]db.TorrentFile, 0, len(info.Files))
	for i, metaFile := range info.Files {
		_, isSelected := fileIndices[uint(i)]
		file := db.NewTorrentFile(torrent.ID, uint(i), metaFile.DisplayPath(info), isSelected, uint(metaFile.Length))
		files = append(files, file)
	}

	totalLength := int64(0)
	for _, cFile := range cTorrent.Files() {
		cFile.SetPriority(torrentLib.PiecePriorityNone)
		totalLength += cFile.Length()
	}
	torrent.TotalLength = totalLength

	downloadLength := int64(0)
	for i, cFile := range cTorrent.Files() {
		if _, isSelected := fileIndices[uint(i)]; isSelected {
			cFile.SetPriority(torrentLib.PiecePriorityNormal)
			downloadLength += cFile.Length()
		}
	}
	torrent.TotalDownloadLength = downloadLength

	queryResult := s.db.Create(&files)
	if queryResult.Error != nil {
		s.onFailedImport(torrent, queryResult.Error)
		return
	}
	if queryResult.RowsAffected == 0 {
		s.onFailedImport(torrent, util.ErrFileMapping)
		return
	}

	s.cTorrentMap.Store(torrent.ID, cTorrent)

	queryResult = s.db.Save(&torrent)
	if queryResult.Error != nil {
		s.onFailedImport(torrent, queryResult.Error)
		return
	}
	if queryResult.RowsAffected == 0 {
		s.onFailedImport(torrent, util.ErrNotFound)
		return
	}
	go s.torrentCompletionWatcher(torrent.ID, torrent.Name, files, downloadLength, cTorrent)
	s.log.Info("torrent initialized", zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name))
}

func (s *TorrentService) AddTorrentFromFile(seriesId uint, tempPath string, fileIndices map[uint]struct{}) (uint, error) {
	newPath, err := s.fileService.GenFilePath(s.infoFolder, tempPath)
	if err != nil {
		return 0, err
	}
	err = os.Rename(tempPath, newPath)
	if err != nil {
		return 0, err
	}
	torrent := db.NewTorrent(seriesId, newPath, db.TORRENT_INFO_FILE)
	queryResult := s.db.Create(&torrent)
	if queryResult.Error != nil {
		deleteErr := os.Remove(newPath)
		if deleteErr != nil {
			s.log.Warn("error deleting torrent on add error", zap.Error(deleteErr))
		}
		return 0, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		deleteErr := os.Remove(newPath)
		if deleteErr != nil {
			s.log.Warn("error deleting torrent on add error", zap.Error(deleteErr))
		}
		return 0, util.ErrCreationFailed
	}
	s.log.Info("added new torrent", zap.Uint("seriesId", seriesId), zap.Uint("torrentId", torrent.ID))
	go s.initTorrent(torrent, fileIndices)
	return torrent.ID, nil
}

var TorrentServiceExport = fx.Options(fx.Provide(NewTorrentService))

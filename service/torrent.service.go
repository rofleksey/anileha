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
	cTorrentMap     map[uint]*torrentLib.Torrent // TODO: use concurrent map
	cTorrentLock    sync.RWMutex
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
		cTorrentMap:     make(map[uint]*torrentLib.Torrent),
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
	s.cTorrentLock.RLock()
	cTorrent := s.cTorrentMap[torrent.ID]
	s.cTorrentLock.RUnlock()
	cFiles := cTorrent.Files()
	bytesRead := int64(0)
	for i, file := range torrent.Files {
		cFile := cFiles[i]
		if file.Selected {
			bytesRead += cFile.BytesCompleted()
		}
	}
	bytesTotal := torrent.TotalDownloadLength
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
	if torrent.Status == db.TORRENT_DOWNLOADING || torrent.Status == db.TORRENT_POSTPROCESSING {
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
	s.cTorrentLock.Lock()
	cTorrent := s.cTorrentMap[torrent.ID]
	delete(s.cTorrentMap, torrent.ID)
	s.cTorrentLock.Unlock()

	if cTorrent != nil {
		cTorrent.Drop()
		<-cTorrent.Closed()
		s.cTorrentLock.Lock()
		delete(s.cTorrentMap, torrent.ID)
		s.cTorrentLock.Unlock()
	}

	for _, cFile := range cTorrent.Files() {
		cFilePath := path.Join(s.downloadsFolder, cFile.DisplayPath())
		err := os.RemoveAll(cFilePath)
		if err != nil {
			s.log.Warn("failed to cleanup torrent file", zap.String("path", cFilePath), zap.Error(err))
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

	for _, file := range torrent.Files {
		// ignore already downloaded files
		if file.ReadyPath != nil {
			continue
		}

		oldPath := path.Join(s.downloadsFolder, torrent.Name, file.TorrentPath)

		// if file is not selected - delete it and continue
		if !file.Selected {
			deleteErr := os.Remove(oldPath)
			if deleteErr != nil {
				s.log.Warn("failed to delete unselected file", zap.String("path", oldPath), zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name), zap.Error(deleteErr))
			}
			continue
		}

		torrentFileName := filepath.Base(file.TorrentPath)
		newPath, err := s.fileService.GetFileDst(torrentReadyRootFolder, torrentFileName)
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
			<-cTorrent.Closed()
			s.cTorrentLock.Lock()
			delete(s.cTorrentMap, id)
			s.cTorrentLock.Unlock()

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
		s.onFailedImport(torrent, errors.New("file mapping failed"))
		return
	}

	s.cTorrentLock.Lock()
	s.cTorrentMap[torrent.ID] = cTorrent
	s.cTorrentLock.Unlock()

	queryResult = s.db.Save(&torrent)
	if queryResult.Error != nil {
		s.onFailedImport(torrent, queryResult.Error)
		return
	}
	if queryResult.RowsAffected == 0 {
		s.onFailedImport(torrent, errors.New("can't find torrent to update"))
		return
	}
	go s.torrentCompletionWatcher(torrent.ID, torrent.Name, files, downloadLength, cTorrent)
	s.log.Info("torrent initialized", zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name))
}

func (s *TorrentService) AddTorrentFromFile(seriesId uint, tempPath string, fileIndices map[uint]struct{}) (uint, error) {
	newPath, err := s.fileService.GetFileDst(s.infoFolder, tempPath)
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

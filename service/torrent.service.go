package service

import (
	"anileha/config"
	"anileha/db"
	"anileha/util"
	"context"
	"errors"
	torrentLib "github.com/anacrolix/torrent"
	"github.com/rofleksey/roflmeta"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
	"path"
	"path/filepath"
	"sort"
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

func NewTorrentService(
	lifecycle fx.Lifecycle,
	database *gorm.DB,
	log *zap.Logger,
	config *config.Config,
	fileService *FileService,
) (*TorrentService, error) {
	if err := database.Model(&db.Torrent{}).Where("status = ? or status = ?", db.TORRENT_DOWNLOADING, db.TORRENT_CREATING).Updates(db.Torrent{Status: db.TORRENT_ERROR}).Error; err != nil {
		return nil, err
	}
	if err := database.Model(&db.TorrentFile{}).Where("status = ?", db.TORRENT_FILE_DOWNLOADING).Updates(map[string]interface{}{"status": db.TORRENT_FILE_ERROR, "selected": false}).Error; err != nil {
		return nil, err
	}
	infoFolder, downloadsFolder, readyFolder, err := createDirs(config)
	if err != nil {
		return nil, err
	}
	clientConfig := torrentLib.NewDefaultClientConfig()
	clientConfig.DataDir = downloadsFolder
	client, err := torrentLib.NewClient(clientConfig)
	if err != nil {
		return nil, err
	}
	retainedTorrents := client.Torrents()
	for _, t := range retainedTorrents {
		t.Drop()
	}
	lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			client.Close()
			return nil
		},
	})
	return &TorrentService{
		db:              database,
		client:          client,
		fileService:     fileService,
		log:             log,
		infoFolder:      infoFolder,
		downloadsFolder: downloadsFolder,
		readyFolder:     readyFolder,
	}, nil
}

func (s *TorrentService) GetTorrentByIdSimple(id uint) (*db.Torrent, error) {
	var torrent db.Torrent
	queryResult := s.db.Preload("Files", func(db *gorm.DB) *gorm.DB {
		return db.Order("torrent_files.episode_index ASC")
	}).First(&torrent, "id = ?", id)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, util.ErrNotFound
	}
	return &torrent, nil
}

func (s *TorrentService) GetTorrentById(id uint) (*db.TorrentWithProgress, error) {
	var torrent db.Torrent
	queryResult := s.db.Preload("Files", func(db *gorm.DB) *gorm.DB {
		return db.Order("torrent_files.episode_index ASC")
	}).First(&torrent, "id = ?", id)
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
	for i := range torrent.Files {
		if torrent.Files[i].Selected {
			cFile := cFiles[torrent.Files[i].TorrentIndex]
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
	queryResult := s.db.Find(&torrentArr, func(db *gorm.DB) *gorm.DB {
		return db.Order("torrents.created_at ASC")
	})
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return torrentArr, nil
}

func (s *TorrentService) DeleteTorrentById(id uint) error {
	var torrent db.Torrent
	queryResult := s.db.Preload("Files", func(db *gorm.DB) *gorm.DB {
		return db.Order("torrent_files.episode_index ASC")
	}).First(&torrent, "id = ?", id)
	if queryResult.Error != nil {
		return queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return util.ErrNotFound
	}
	if torrent.Status == db.TORRENT_DOWNLOADING {
		err := s.StopTorrent(torrent)
		if err != nil {
			return err
		}
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

	deleteErr := os.Remove(torrent.FilePath)
	if deleteErr != nil {
		s.log.Warn("failed to cleanup torrent info", zap.String("path", torrent.FilePath), zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name), zap.Error(deleteErr))
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
	queryResult := s.db.Preload("Files", func(db *gorm.DB) *gorm.DB {
		return db.Order("torrent_files.episode_index ASC")
	}).First(&torrent, "id = ?", id)
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

	for i := range torrent.Files {
		// ignore already downloaded files
		if torrent.Files[i].ReadyPath != nil {
			continue
		}

		oldPath := path.Join(s.downloadsFolder, torrent.Name, torrent.Files[i].TorrentPath)

		// if file is not selected - delete it and continue
		if !torrent.Files[i].Selected {
			_ = os.Remove(oldPath)
			//if deleteErr != nil {
			//	s.log.Warn("failed to delete unselected file", zap.String("path", oldPath), zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name), zap.Error(deleteErr))
			//}
			continue
		}

		torrentFileName := filepath.Base(torrent.Files[i].TorrentPath)
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
		torrent.Files[i].ReadyPath = &newPath
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

// TODO: torrent eta + speed

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
			for _, file := range files {
				if file.Selected {
					cFile := cFiles[file.TorrentIndex]
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

// TODO: restore cTorrent on startup

func (s *TorrentService) initTorrent(torrent db.Torrent) error {
	cTorrent, err := s.client.AddTorrentFromFile(torrent.FilePath)
	if err != nil {
		s.onFailedImport(torrent, err)
		return err
	}

	s.cTorrentMap.Store(torrent.ID, cTorrent)
	<-cTorrent.GotInfo()
	info := cTorrent.Info()

	torrent.Name = info.BestName()
	torrent.Status = db.TORRENT_IDLE
	s.log.Info("torrent received info", zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name))

	files := make([]db.TorrentFile, 0, len(info.Files))
	filenames := make([]string, 0, len(info.Files))
	for i, metaFile := range info.Files {
		file := db.NewTorrentFile(torrent.ID, uint(i), metaFile.DisplayPath(info), false, uint(metaFile.Length))
		files = append(files, file)
		filenames = append(filenames, file.TorrentPath)
	}

	// sort by season/episode
	episodeMetadata, multipleSuccess := roflmeta.ParseMultipleEpisodeMetadata(filenames)
	if !multipleSuccess {
		s.log.Warn("failed to apply multiple episode metadata extraction", zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name))
	}
	for i := range files {
		files[i].Episode = episodeMetadata[i].Episode
		files[i].Season = episodeMetadata[i].Season
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].Season == files[j].Season {
			return files[i].Episode < files[j].Episode
		}
		return files[i].Season < files[j].Season
	})
	for i := range files {
		files[i].EpisodeIndex = uint(i)
	}

	totalLength := int64(0)
	for _, cFile := range cTorrent.Files() {
		cFile.SetPriority(torrentLib.PiecePriorityNone)
		totalLength += cFile.Length()
	}
	torrent.TotalLength = totalLength

	queryResult := s.db.Create(files)
	if queryResult.Error != nil {
		s.onFailedImport(torrent, queryResult.Error)
		return queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		s.onFailedImport(torrent, util.ErrFileMapping)
		return util.ErrFileMapping
	}

	queryResult = s.db.Save(&torrent)
	if queryResult.Error != nil {
		s.onFailedImport(torrent, queryResult.Error)
		return queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		s.onFailedImport(torrent, util.ErrNotFound)
		return util.ErrNotFound
	}
	cTorrent.Drop()
	s.cTorrentMap.Delete(torrent.ID)
	s.log.Info("torrent initialized", zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name))
	return nil
}

func (s *TorrentService) AddTorrentFromFile(seriesId uint, tempPath string) (uint, error) {
	newPath, err := s.fileService.GenFilePath(s.infoFolder, tempPath)
	if err != nil {
		return 0, err
	}
	err = os.Rename(tempPath, newPath)
	if err != nil {
		return 0, err
	}
	torrent := db.NewTorrent(seriesId, newPath)
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
	err = s.initTorrent(torrent)
	if err != nil {
		return 0, err
	}
	s.log.Info("added new torrent", zap.Uint("seriesId", seriesId), zap.Uint("torrentId", torrent.ID))
	return torrent.ID, nil
}

func (s *TorrentService) StartTorrent(torrent db.Torrent, fileIndices map[uint]struct{}) error {
	mapEntry, exists := s.cTorrentMap.Load(torrent.ID)
	if exists {
		cTorrent, castOk := mapEntry.(*torrentLib.Torrent)
		if castOk {
			cTorrent.Drop()
			<-cTorrent.Closed()
		}
	}
	cTorrent, err := s.client.AddTorrentFromFile(torrent.FilePath)
	if err != nil {
		return err
	}
	s.cTorrentMap.Store(torrent.ID, cTorrent)
	<-cTorrent.GotInfo()
	downloadLength := int64(0)
	for i := range torrent.Files {
		cFile := cTorrent.Files()[torrent.Files[i].TorrentIndex]
		cFile.SetPriority(torrentLib.PiecePriorityNone)
		torrent.Files[i].Selected = false
		torrent.Files[i].Status = db.TORRENT_FILE_IDLE
	}
	for i := range torrent.Files {
		if _, isSelected := fileIndices[uint(i)]; isSelected {
			cFile := cTorrent.Files()[torrent.Files[i].TorrentIndex]
			cFile.SetPriority(torrentLib.PiecePriorityNormal)
			downloadLength += cFile.Length()
			torrent.Files[i].Selected = true
			torrent.Files[i].Status = db.TORRENT_FILE_DOWNLOADING
		}
	}
	err = s.db.Model(&db.Torrent{}).Where("id = ?", torrent.ID).Updates(db.Torrent{Status: db.TORRENT_DOWNLOADING, TotalDownloadLength: downloadLength}).Error
	if err != nil {
		return err
	}
	queryResult := s.db.Save(&torrent.Files)
	if queryResult.Error != nil {
		return queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return util.ErrFileMapping
	}
	go s.torrentCompletionWatcher(torrent.ID, torrent.Name, torrent.Files, downloadLength, cTorrent)
	return nil
}

func (s *TorrentService) StopTorrent(torrent db.Torrent) error {
	mapEntry, exists := s.cTorrentMap.Load(torrent.ID)
	if !exists {
		return util.ErrCTorrentNotFound
	}
	cTorrent, castOk := mapEntry.(*torrentLib.Torrent)
	if !castOk {
		return util.ErrCTorrentCorrupted
	}
	downloadLength := int64(0)
	for i := range torrent.Files {
		torrent.Files[i].Selected = false
		torrent.Files[i].Status = db.TORRENT_FILE_IDLE
	}
	cTorrent.Drop()
	<-cTorrent.Closed()
	err := s.db.Model(&db.Torrent{}).Where("id = ?", torrent.ID).Updates(db.Torrent{Status: db.TORRENT_DOWNLOADING, TotalDownloadLength: downloadLength}).Error
	if err != nil {
		return err
	}
	queryResult := s.db.Save(&torrent.Files)
	if queryResult.Error != nil {
		return queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return util.ErrFileMapping
	}
	return nil
}

var TorrentServiceExport = fx.Options(fx.Provide(NewTorrentService))

package service

import (
	"anileha/config"
	"anileha/db"
	"anileha/rest"
	"anileha/util"
	"context"
	"errors"
	"fmt"
	torrentLib "github.com/anacrolix/torrent"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
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
		return "", "", "", fmt.Errorf("failed to get working dir")
	}
	infoFolder := path.Join(workingDir, config.Data.Dir, util.TorrentInfoSubDir)
	err = os.MkdirAll(infoFolder, os.ModePerm)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create info dir")
	}

	downloadsFolder := path.Join(workingDir, config.Data.Dir, util.TorrentDownloadsSubDir)
	err = os.MkdirAll(downloadsFolder, os.ModePerm)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create downloads dir")
	}

	readyFolder := path.Join(workingDir, config.Data.Dir, util.TorrentReadySubDir)
	err = os.MkdirAll(readyFolder, os.ModePerm)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create ready dir")
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
	if err := database.Model(&db.Torrent{}).Where("status = ? or status = ?", db.TorrentDownloading, db.TorrentCreating).Updates(db.Torrent{Status: db.TorrentError}).Error; err != nil {
		return nil, err
	}
	if err := database.Model(&db.TorrentFile{}).Where("status = ?", db.TorrentFileDownloading).Updates(map[string]interface{}{"status": db.TorrentFileError, "selected": false}).Error; err != nil {
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

// cleanUpTorrent Drops cTorrent, removes all torrent files
func (s *TorrentService) cleanUpTorrent(torrent db.Torrent) {
	mapValue, exists := s.cTorrentMap.LoadAndDelete(torrent.ID)

	if exists {
		cTorrent := mapValue.(*torrentLib.Torrent)
		cTorrent.Drop()
		<-cTorrent.Closed()
		for _, file := range torrent.Files {
			cFile := cTorrent.Files()[file.TorrentIndex]
			var cFilePath string
			if len(torrent.Files) > 1 {
				cFilePath = path.Join(s.downloadsFolder, torrent.Name, cFile.DisplayPath())
			} else {
				cFilePath = path.Join(s.downloadsFolder, cFile.DisplayPath())
			}
			_ = os.RemoveAll(cFilePath)
		}
	}

	for _, file := range torrent.Files {
		if file.ReadyPath != nil {
			err := os.RemoveAll(*file.ReadyPath)
			if err != nil {
				s.log.Error("failed to cleanup torrent file", zap.String("path", *file.ReadyPath), zap.Error(err))
			}
		}
	}

	deleteErr := os.Remove(torrent.FilePath)
	if deleteErr != nil {
		s.log.Warn("failed to cleanup torrent info", zap.String("path", torrent.FilePath), zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name), zap.Error(deleteErr))
	}
}

func (s *TorrentService) GetTorrentById(id uint) (*db.Torrent, error) {
	var torrent db.Torrent
	queryResult := s.db.Preload("Files", func(db *gorm.DB) *gorm.DB {
		return db.Order("torrent_files.client_index ASC")
	}).First(&torrent, "id = ?", id)
	if queryResult.Error != nil {
		return nil, rest.ErrInternal(queryResult.Error.Error())
	}
	if queryResult.RowsAffected == 0 {
		return nil, rest.ErrNotFoundInst
	}
	return &torrent, nil
}

func (s *TorrentService) GetTorrentFileById(id uint) (*db.TorrentFile, error) {
	var file db.TorrentFile
	queryResult := s.db.First(&file, "id = ?", id)
	if queryResult.Error != nil {
		return nil, rest.ErrInternal(queryResult.Error.Error())
	}
	if queryResult.RowsAffected == 0 {
		return nil, rest.ErrNotFoundInst
	}
	return &file, nil
}

func (s *TorrentService) GetAllTorrents() ([]db.Torrent, error) {
	var torrentArr []db.Torrent
	queryResult := s.db.Order("torrents.created_at DESC").Find(&torrentArr)
	if queryResult.Error != nil {
		return nil, rest.ErrInternal(queryResult.Error.Error())
	}
	return torrentArr, nil
}

func (s *TorrentService) GetTorrentsBySeriesId(seriesId uint) ([]db.Torrent, error) {
	var torrentArr []db.Torrent
	queryResult := s.db.Where("series_id = ?", seriesId).Order("torrents.created_at DESC").Find(&torrentArr)
	if queryResult.Error != nil {
		return nil, rest.ErrInternal(queryResult.Error.Error())
	}
	return torrentArr, nil
}

func (s *TorrentService) DeleteTorrentById(id uint) error {
	var torrent db.Torrent
	queryResult := s.db.Preload("Files", func(db *gorm.DB) *gorm.DB {
		return db.Order("torrent_files.client_index ASC")
	}).First(&torrent, "id = ?", id)
	if queryResult.Error != nil {
		return rest.ErrInternal(queryResult.Error.Error())
	}
	if queryResult.RowsAffected == 0 {
		return rest.ErrNotFoundInst
	}
	if torrent.Status == db.TorrentDownloading {
		err := s.StopTorrent(torrent)
		if err != nil {
			return rest.ErrInternal(err.Error())
		}
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&db.TorrentFile{}, "torrent_id = ?", id).Error; err != nil {
			return rest.ErrInternal(err.Error())
		}
		queryResult = tx.Delete(&db.Torrent{}, id)
		if queryResult.Error != nil {
			return rest.ErrInternal(queryResult.Error.Error())
		}
		if queryResult.RowsAffected == 0 {
			return rest.ErrNotFoundInst
		}
		go s.cleanUpTorrent(torrent)
		return nil
	})
	return err
}

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
		return db.Order("torrent_files.client_index ASC")
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

		var oldPath string

		if len(torrent.Files) > 1 {
			oldPath = path.Join(s.downloadsFolder, torrent.Name, torrent.Files[i].TorrentPath)
		} else {
			oldPath = path.Join(s.downloadsFolder, torrent.Files[i].TorrentPath)
		}

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
		if err := tx.Model(&db.Torrent{}).Where("id = ?", id).Updates(db.Torrent{Status: db.TorrentReady}).Error; err != nil {
			return err
		}

		// update downloaded files' statuses
		for _, file := range torrent.Files {
			if file.Selected && file.Status != db.TorrentFileReady {
				if err := tx.Model(&db.TorrentFile{}).Where("id = ?", file.ID).Updates(db.TorrentFile{Status: db.TorrentFileReady, ReadyPath: file.ReadyPath}).Error; err != nil {
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
func (s *TorrentService) torrentCompletionWatcher(id uint, name string, files []db.TorrentFile, totalDownloadLength uint, cTorrent *torrentLib.Torrent) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	cFiles := cTorrent.Files()
	etaCalc := util.NewEtaCalculator(0, float64(totalDownloadLength))
	etaCalc.Start()
	for {
		select {
		case <-cTorrent.Closed():
			s.log.Info("torrentCompletionWatcher exited", zap.Uint("torrentId", id), zap.String("torrentName", name), zap.String("cause", "closed"))
			return
		case <-ticker.C:
			bytesRead := uint(0)
			for _, file := range files {
				if file.Selected {
					cFile := cFiles[file.TorrentIndex]
					bytesRead += uint(cFile.BytesCompleted())
				}
			}
			etaCalc.Update(float64(bytesRead))
			progress := etaCalc.GetProgress()
			go func() {
				if err := s.db.Model(&db.Torrent{}).Where("id = ?", id).Updates(db.Torrent{Progress: progress, BytesRead: bytesRead}).Error; err != nil {
					s.log.Error("failed to update db on torrent progress", zap.Uint("torrentId", id), zap.String("torrentName", name), zap.Error(err))
				}
			}()
			if bytesRead < totalDownloadLength {
				continue
			}

			cTorrent.Drop()
			s.cTorrentMap.Delete(id)
			<-cTorrent.Closed()

			s.onTorrentCompletion(id)
			return
		}
	}
}

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
	torrent.Status = db.TorrentIdle
	s.log.Info("torrent received info", zap.Int("fileCount", len(cTorrent.Files())), zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name))

	files := make([]db.TorrentFile, 0, len(info.Files))
	filenames := make([]string, 0, len(info.Files))
	for i, torrentFile := range cTorrent.Files() {
		file := db.TorrentFile{
			TorrentId:    torrent.ID,
			TorrentIndex: i,
			TorrentPath:  torrentFile.DisplayPath(),
			Length:       uint(torrentFile.Length()),
		}
		files = append(files, file)
		filenames = append(filenames, file.TorrentPath)
	}

	// sort by episode / season
	sort.Slice(files, func(i, j int) bool {
		return files[i].TorrentPath < files[j].TorrentPath
	})
	for i := range files {
		files[i].ClientIndex = i
	}

	totalLength := uint(0)
	for _, cFile := range cTorrent.Files() {
		cFile.SetPriority(torrentLib.PiecePriorityNone)
		totalLength += uint(cFile.Length())
	}
	torrent.TotalLength = totalLength

	err = s.db.Transaction(func(tx *gorm.DB) error {
		queryResult := tx.Create(files)
		if queryResult.Error != nil {
			s.onFailedImport(torrent, queryResult.Error)
			return queryResult.Error
		}
		if queryResult.RowsAffected == 0 {
			s.onFailedImport(torrent, util.ErrFileMapping)
			return util.ErrFileMapping
		}

		queryResult = tx.Save(&torrent)
		if queryResult.Error != nil {
			s.onFailedImport(torrent, queryResult.Error)
			return queryResult.Error
		}
		if queryResult.RowsAffected == 0 {
			s.onFailedImport(torrent, rest.ErrNotFoundInst)
			return rest.ErrNotFoundInst
		}

		return nil
	})

	if err != nil {
		return err
	}

	cTorrent.Drop()
	s.cTorrentMap.Delete(torrent.ID)
	s.log.Info("torrent initialized", zap.Uint("torrentId", torrent.ID), zap.String("torrentName", torrent.Name))
	return nil
}

func (s *TorrentService) AddTorrentFromFile(seriesId uint, tempPath string) error {
	newPath, err := s.fileService.GenFilePath(s.infoFolder, tempPath)
	if err != nil {
		return rest.ErrInternal(err.Error())
	}
	err = os.Rename(tempPath, newPath)
	if err != nil {
		return rest.ErrInternal(err.Error())
	}
	torrent := db.Torrent{
		SeriesId: seriesId,
		FilePath: newPath,
	}
	queryResult := s.db.Create(&torrent)
	if queryResult.Error != nil {
		deleteErr := os.Remove(newPath)
		if deleteErr != nil {
			s.log.Warn("error deleting torrent on add error", zap.Error(deleteErr))
		}
		return rest.ErrInternal(queryResult.Error.Error())
	}
	if queryResult.RowsAffected == 0 {
		deleteErr := os.Remove(newPath)
		if deleteErr != nil {
			s.log.Warn("error deleting torrent on add error", zap.Error(deleteErr))
		}
		return rest.ErrCreationFailed
	}
	go func() {
		err = s.initTorrent(torrent)
		if err != nil {
			s.log.Error("failed to init torrent", zap.Error(err))
		}
	}()
	s.log.Info("adding new torrent in the background", zap.Uint("seriesId", seriesId), zap.Uint("torrentId", torrent.ID))
	return nil
}

func (s *TorrentService) StartTorrent(torrent db.Torrent, fileIndices []int) error {
	mapEntry, exists := s.cTorrentMap.Load(torrent.ID)
	if exists {
		cTorrent := mapEntry.(*torrentLib.Torrent)
		cTorrent.Drop()
		<-cTorrent.Closed()
	}
	cTorrent, err := s.client.AddTorrentFromFile(torrent.FilePath)
	if err != nil {
		return rest.ErrInternal(fmt.Sprintf("failed to add torrent from file: %s", err.Error()))
	}
	s.cTorrentMap.Store(torrent.ID, cTorrent)
	<-cTorrent.GotInfo()
	downloadLength := uint(0)
	for i := range torrent.Files {
		cFile := cTorrent.Files()[torrent.Files[i].TorrentIndex]
		cFile.SetPriority(torrentLib.PiecePriorityNone)
		torrent.Files[i].Selected = false
		torrent.Files[i].Status = db.TorrentFileIdle
	}
	unselectedFiles := make([]uint, 0, len(torrent.Files))
	selectedFiles := make([]uint, 0, len(torrent.Files))
	for i := range torrent.Files {
		if fileIndices == nil || slices.Contains(fileIndices, torrent.Files[i].ClientIndex) {
			cFile := cTorrent.Files()[torrent.Files[i].TorrentIndex]
			cFile.SetPriority(torrentLib.PiecePriorityNormal)
			downloadLength += uint(cFile.Length())
			torrent.Files[i].Selected = true
			torrent.Files[i].Status = db.TorrentFileDownloading
			selectedFiles = append(selectedFiles, torrent.Files[i].ID)
		} else {
			unselectedFiles = append(unselectedFiles, torrent.Files[i].ID)
		}
	}
	err = s.db.Model(&db.Torrent{}).Where("id = ?", torrent.ID).Updates(db.Torrent{Status: db.TorrentDownloading, TotalDownloadLength: downloadLength}).Error
	if err != nil {
		return rest.ErrInternal(err.Error())
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, id := range unselectedFiles {
			err = s.db.Model(&db.TorrentFile{}).Where("id = ?", id).Updates(map[string]interface{}{"status": db.TorrentFileIdle, "selected": false}).Error
			if err != nil {
				return err
			}
		}
		for _, id := range selectedFiles {
			err = s.db.Model(&db.TorrentFile{}).Where("id = ?", id).Updates(map[string]interface{}{"status": db.TorrentFileDownloading, "selected": true}).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return rest.ErrInternal(err.Error())
	}
	go s.torrentCompletionWatcher(torrent.ID, torrent.Name, torrent.Files, downloadLength, cTorrent)
	return nil
}

func (s *TorrentService) StopTorrent(torrent db.Torrent) error {
	mapEntry, exists := s.cTorrentMap.Load(torrent.ID)
	if !exists {
		return rest.ErrNotFoundInst
	}
	cTorrent := mapEntry.(*torrentLib.Torrent)
	downloadLength := int64(0)
	for i := range torrent.Files {
		torrent.Files[i].Selected = false
		torrent.Files[i].Status = db.TorrentFileIdle
	}
	cTorrent.Drop()
	<-cTorrent.Closed()
	err := s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&db.Torrent{}).Where("id = ?", torrent.ID).Updates(map[string]interface{}{"status": db.TorrentIdle, "total_download_length": downloadLength}).Error
		if err != nil {
			return rest.ErrInternal(err.Error())
		}
		err = tx.Model(&db.TorrentFile{}).Where("torrent_id = ?", torrent.ID).Updates(map[string]interface{}{"status": db.TorrentFileIdle, "selected": false}).Error
		if err != nil {
			return rest.ErrInternal(err.Error())
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

var TorrentServiceExport = fx.Options(fx.Provide(NewTorrentService))

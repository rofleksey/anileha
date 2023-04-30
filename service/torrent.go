package service

import (
	"anileha/config"
	"anileha/db"
	"anileha/db/repo"
	"anileha/ffmpeg/analyze"
	"anileha/ffmpeg/command"
	"anileha/rest/engine"
	"anileha/util"
	"anileha/util/meta"
	"context"
	"errors"
	"fmt"
	torrentLib "github.com/anacrolix/torrent"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"golang.org/x/time/rate"
	"gorm.io/datatypes"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"
)

type TorrentService struct {
	torrentRepo     *repo.TorrentRepo
	client          *torrentLib.Client
	cTorrentMap     sync.Map // cTorrentMap Stores torrentLib.Client torrent entries [uint -> *torrentLib.Torrent]
	fileService     *FileService
	analyzer        *analyze.ProbeAnalyzer
	convertService  *ConversionService
	fontService     *FontService
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
	torrentRepo *repo.TorrentRepo,
	log *zap.Logger,
	config *config.Config,
	fileService *FileService,
	analyzer *analyze.ProbeAnalyzer,
	convertService *ConversionService,
	fontService *FontService,
) (*TorrentService, error) {
	if err := torrentRepo.ResetDownloadStatus(); err != nil {
		return nil, fmt.Errorf("failed to reset download status: %w", err)
	}
	infoFolder, downloadsFolder, readyFolder, err := createDirs(config)
	if err != nil {
		return nil, err
	}
	clientConfig := torrentLib.NewDefaultClientConfig()
	clientConfig.DataDir = downloadsFolder
	downloadRate := rate.Every(time.Second / time.Duration(config.Data.DownloadBpsLimit))
	uploadRate := rate.Every(time.Second / time.Duration(config.Data.UploadBpsLimit))
	clientConfig.DownloadRateLimiter = rate.NewLimiter(downloadRate, config.Data.DownloadBpsLimit)
	clientConfig.UploadRateLimiter = rate.NewLimiter(uploadRate, config.Data.UploadBpsLimit)
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
			<-client.Closed()
			return nil
		},
	})
	return &TorrentService{
		torrentRepo:     torrentRepo,
		client:          client,
		fileService:     fileService,
		analyzer:        analyzer,
		convertService:  convertService,
		fontService:     fontService,
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
			_ = os.RemoveAll(*file.ReadyPath)
		}
	}

	_ = os.Remove(torrent.FilePath)

	torrentIdStr := strconv.FormatUint(uint64(torrent.ID), 10)
	torrentReadyRootFolder := path.Join(s.readyFolder, torrentIdStr)

	_ = os.Remove(torrentReadyRootFolder)

	torrentDownloadRootFolder := path.Join(s.downloadsFolder, torrent.Name)
	_ = os.Remove(torrentDownloadRootFolder)
}

func (s *TorrentService) GetById(id uint) (*db.Torrent, error) {
	torrent, err := s.torrentRepo.GetById(id, false)
	if err != nil {
		return nil, engine.ErrInternal(err.Error())
	}
	if torrent == nil {
		return nil, engine.ErrNotFoundInst
	}
	return torrent, nil
}

func (s *TorrentService) GetByIdWithSeries(id uint) (*db.Torrent, error) {
	torrent, err := s.torrentRepo.GetById(id, true)
	if err != nil {
		return nil, engine.ErrInternal(err.Error())
	}
	if torrent == nil {
		return nil, engine.ErrNotFoundInst
	}
	return torrent, nil
}

func (s *TorrentService) GetAll() ([]db.Torrent, error) {
	torrentArr, err := s.torrentRepo.GetAll()
	if err != nil {
		return nil, engine.ErrInternal(err.Error())
	}
	return torrentArr, nil
}

func (s *TorrentService) GetBySeriesId(seriesId uint) ([]db.Torrent, error) {
	torrentArr, err := s.torrentRepo.GetBySeriesId(seriesId)
	if err != nil {
		return nil, engine.ErrInternal(err.Error())
	}
	return torrentArr, nil
}

func (s *TorrentService) DeleteById(id uint) error {
	torrent, err := s.torrentRepo.GetById(id, false)
	if err != nil {
		return engine.ErrInternal(err.Error())
	}
	if torrent == nil {
		return engine.ErrNotFoundInst
	}

	if torrent.Status == db.TorrentDownload {
		err := s.Stop(*torrent)
		if err != nil {
			return engine.ErrInternal(err.Error())
		}
	}

	err = s.torrentRepo.DeleteById(id)
	if err != nil {
		return engine.ErrInternal(err.Error())
	}

	go s.cleanUpTorrent(*torrent)

	return nil
}

func (s *TorrentService) onFailedImport(torrent db.Torrent, err error) {
	s.log.Warn("failed to import torrent",
		zap.Uint("torrentId", torrent.ID),
		zap.String("torrentName", torrent.Name),
		zap.Error(err))
	deleteErr := s.DeleteById(torrent.ID)
	if deleteErr != nil {
		s.log.Warn("failed to delete torrent DB entry",
			zap.Uint("torrentId", torrent.ID),
			zap.String("torrentName", torrent.Name),
			zap.Error(deleteErr))
	}
	s.cleanUpTorrent(torrent)
}

// prepareForAnalysis Creates READY folder, moves torrent files into it, updates DB entries
func (s *TorrentService) prepareForAnalysis(id uint) {
	torrent, err := s.torrentRepo.GetById(id, false)
	if err != nil {
		s.log.Error("failed to complete torrent",
			zap.Uint("torrentId", torrent.ID),
			zap.Error(err))
		return
	}
	if torrent == nil {
		s.log.Error("failed to complete torrent",
			zap.Uint("torrentId", torrent.ID),
			zap.Error(errors.New("torrent not found")))
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

		oldPath = path.Join(s.downloadsFolder, torrent.Name, torrent.Files[i].TorrentPath)
		if _, err := os.Stat(oldPath); err != nil {
			oldPath = path.Join(s.downloadsFolder, torrent.Files[i].TorrentPath)
		}

		// if file is not selected - delete it and continue
		if !torrent.Files[i].Selected {
			_ = os.Remove(oldPath)
			continue
		}

		torrentFileName := filepath.Base(torrent.Files[i].TorrentPath)
		newPath, err := s.fileService.GenFilePath(torrentReadyRootFolder, torrentFileName)
		if err != nil {
			s.log.Error("failed to acquire temp file",
				zap.Uint("torrentId", torrent.ID),
				zap.String("torrentName", torrent.Name),
				zap.Error(err))
			return
		}

		newPathDir := filepath.Dir(newPath)
		createDirErr := os.MkdirAll(newPathDir, os.ModePerm)
		if createDirErr != nil {
			s.log.Error("failed to create dirs",
				zap.String("path", newPathDir),
				zap.Uint("torrentId", torrent.ID),
				zap.String("torrentName", torrent.Name),
				zap.Error(err))
			return
		}

		err = os.Rename(oldPath, newPath)
		if err != nil {
			s.log.Error("failed to move ready torrent file",
				zap.Uint("torrentId", torrent.ID),
				zap.String("torrentName", torrent.Name),
				zap.String("from", oldPath),
				zap.String("to", newPath),
				zap.Error(err))
			return
		}
		torrent.Files[i].ReadyPath = &newPath
	}

	s.fontService.LoadFonts(context.Background(), torrent.Files)

	err = s.torrentRepo.SetPreAnalysis(*torrent)
	if err != nil {
		s.log.Error("transaction on torrent completion failed",
			zap.Uint("torrentId", torrent.ID),
			zap.String("torrentName", torrent.Name),
			zap.Error(err))
		return
	}

	s.log.Info("torrent prepared for analysis",
		zap.Uint("torrentId", torrent.ID),
		zap.String("torrentName", torrent.Name))
}

func (s *TorrentService) performAnalysis(id uint, etaCalc *util.EtaCalculator) {
	torrent, err := s.torrentRepo.GetById(id, false)
	if err != nil {
		s.log.Error("failed to complete torrent",
			zap.Uint("torrentId", torrent.ID),
			zap.Error(err))
		return
	}
	if torrent == nil {
		s.log.Error("failed to complete torrent",
			zap.Uint("torrentId", torrent.ID),
			zap.Error(errors.New("torrent not found")))
		return
	}

	filesToAnalyze := 0

	for i := range torrent.Files {
		// ignore already analyzed files
		if torrent.Files[i].Status != db.TorrentFileAnalysis {
			continue
		}
		filesToAnalyze++
	}

	// doesn't reset elapsed this way
	etaCalc.ContinueWithNewValues(0, float64(filesToAnalyze))

	err = s.torrentRepo.UpdateProgress(id, etaCalc.GetProgress())
	if err != nil {
		s.log.Error("failed to update torrent progress",
			zap.Uint("torrentId", torrent.ID),
			zap.String("torrentName", torrent.Name),
			zap.Error(err))
		return
	}

	filesAnalyzed := 0
	for i := range torrent.Files {
		// ignore already analyzed files
		if torrent.Files[i].Status != db.TorrentFileAnalysis {
			continue
		}

		// ignore non-video files
		if torrent.Files[i].Type != util.FileTypeVideo {
			continue
		}

		result, err := s.analyzer.Probe(*torrent.Files[i].ReadyPath)
		if err != nil {
			s.log.Error("failed to analyze torrent file",
				zap.Uint("torrentId", torrent.ID),
				zap.Uint("fileId", torrent.Files[i].ID),
				zap.String("torrentName", torrent.Name),
				zap.Error(err))
			return
		}

		err = s.torrentRepo.SetFileAnalysis(torrent.Files[i].ID, *result)
		if err != nil {
			s.log.Error("failed to set torrent file analysis",
				zap.Uint("torrentId", torrent.ID),
				zap.Uint("fileId", torrent.Files[i].ID),
				zap.String("torrentName", torrent.Name),
				zap.Error(err))
			return
		}

		filesAnalyzed++
		etaCalc.Update(float64(filesAnalyzed))
		progress := etaCalc.GetProgress()
		progress.Speed = 0
		err = s.torrentRepo.UpdateProgress(id, progress)
		if err != nil {
			s.log.Error("failed to update torrent progress",
				zap.Uint("torrentId", torrent.ID),
				zap.String("torrentName", torrent.Name),
				zap.Error(err))
			return
		}
	}

	err = s.torrentRepo.SetReady(*torrent)
	if err != nil {
		s.log.Error("transaction on torrent completion failed",
			zap.Uint("torrentId", torrent.ID),
			zap.String("torrentName", torrent.Name),
			zap.Error(err))
		return
	}

	if torrent.Auto.Data() != nil {
		go s.startAutoConvert(id)
	}

	s.log.Info("torrent finished",
		zap.Uint("torrentId", torrent.ID),
		zap.String("torrentName", torrent.Name))
}

func (s *TorrentService) startAutoDownload(id uint) {
	torrent, err := s.torrentRepo.GetById(id, false)
	if err != nil {
		s.log.Error("failed to get torrent",
			zap.Uint("torrentId", torrent.ID),
			zap.Error(err))
		return
	}
	fileIndices := make([]int, 0, len(torrent.Files))
	for _, file := range torrent.Files {
		fileIndices = append(fileIndices, file.ClientIndex)
	}
	err = s.Start(*torrent, fileIndices)
	if err != nil {
		s.log.Error("failed to autostart download",
			zap.Uint("torrentId", torrent.ID),
			zap.String("torrentName", torrent.Name),
			zap.Error(err))
	}
}

func (s *TorrentService) startAutoConvert(id uint) {
	torrent, err := s.torrentRepo.GetById(id, true)
	if err != nil {
		s.log.Error("failed to get torrent",
			zap.Uint("torrentId", torrent.ID),
			zap.Error(err))
		return
	}

	torrentFiles := make([]db.TorrentFile, 0, len(torrent.Files))
	prefsArr := make([]command.Preferences, 0, len(torrent.Files))

	for _, file := range torrent.Files {
		if file.Status != db.TorrentFileReady || file.ReadyPath == nil {
			continue
		}
		torrentFiles = append(torrentFiles, file)
		prefsArr = append(prefsArr, command.Preferences{
			Audio: command.PreferencesData{
				Lang: torrent.Auto.Data().AudioLang,
			},
			Sub: command.PreferencesData{
				Lang: torrent.Auto.Data().SubLang,
			},
			Episode: file.SuggestedMetadata.Data().Episode,
			Season:  file.SuggestedMetadata.Data().Season,
		})
	}

	err = s.convertService.StartConversion(*torrent, torrentFiles, prefsArr)
	if err != nil {
		s.log.Error("failed to autostart conversion",
			zap.Uint("torrentId", torrent.ID),
			zap.String("torrentName", torrent.Name),
			zap.Error(err))
	}
}

// torrentCompletionWatcher Polls for torrent's completion, calls prepareForAnalysis
func (s *TorrentService) torrentCompletionWatcher(id uint, name string, files []db.TorrentFile,
	totalDownloadLength uint, cTorrent *torrentLib.Torrent) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	cFiles := cTorrent.Files()
	etaCalc := util.NewEtaCalculator(0, float64(totalDownloadLength))
	etaCalc.Start()
	for {
		select {
		case <-cTorrent.Closed():
			s.log.Info("torrentCompletionWatcher exited",
				zap.Uint("torrentId", id),
				zap.String("torrentName", name),
				zap.String("cause", "closed"))
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
				if err := s.torrentRepo.UpdateProgressAndBytesRead(id, progress, bytesRead); err != nil {
					s.log.Error("failed to update db on torrent progress",
						zap.Uint("torrentId", id),
						zap.String("torrentName", name),
						zap.Error(err))
				}
			}()
			if bytesRead < totalDownloadLength {
				continue
			}

			cTorrent.Drop()
			s.cTorrentMap.Delete(id)
			<-cTorrent.Closed()

			s.prepareForAnalysis(id)
			s.performAnalysis(id, etaCalc)
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
	s.log.Info("torrent received info",
		zap.Int("fileCount", len(cTorrent.Files())),
		zap.Uint("torrentId", torrent.ID),
		zap.String("torrentName", torrent.Name))

	files := make([]db.TorrentFile, 0, len(info.Files))
	filenames := make([]string, 0, len(info.Files))

	for i, torrentFile := range cTorrent.Files() {
		torrentPath := torrentFile.DisplayPath()
		episodeMeta := meta.GuessEpisodeMetadata(torrentPath)
		file := db.TorrentFile{
			TorrentId:         torrent.ID,
			TorrentIndex:      i,
			TorrentPath:       torrentPath,
			Length:            uint(torrentFile.Length()),
			Type:              util.GetFileType(torrentPath),
			SuggestedMetadata: datatypes.NewJSONType(episodeMeta),
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

	err = s.torrentRepo.InitFiles(torrent, files)
	if err != nil {
		s.onFailedImport(torrent, engine.ErrNotFoundInst)
		return err
	}

	cTorrent.Drop()
	s.cTorrentMap.Delete(torrent.ID)
	<-cTorrent.Closed()

	if torrent.Auto.Data() != nil {
		go s.startAutoDownload(torrent.ID)
	}

	s.log.Info("torrent initialized",
		zap.Uint("torrentId", torrent.ID),
		zap.String("torrentName", torrent.Name))
	return nil
}

func (s *TorrentService) AddFromFile(seriesId uint, tempPath string, auto *db.AutoTorrent) error {
	newPath, err := s.fileService.GenFilePath(s.infoFolder, tempPath)
	if err != nil {
		return engine.ErrInternal(err.Error())
	}
	err = os.Rename(tempPath, newPath)
	if err != nil {
		return engine.ErrInternal(err.Error())
	}
	torrent := db.Torrent{
		SeriesId: &seriesId,
		FilePath: newPath,
		Auto:     datatypes.NewJSONType(auto),
	}
	_, err = s.torrentRepo.Create(&torrent)
	if err != nil {
		deleteErr := os.Remove(newPath)
		if deleteErr != nil {
			s.log.Warn("error deleting torrent on add error", zap.Error(deleteErr))
		}
		return engine.ErrInternal(err.Error())
	}
	s.log.Info("adding new torrent in the background",
		zap.Uint("seriesId", seriesId),
		zap.Uint("torrentId", torrent.ID))
	go func() {
		err = s.initTorrent(torrent)
		if err != nil {
			s.log.Error("failed to init torrent", zap.Error(err))
		}
	}()
	return nil
}

func (s *TorrentService) Start(torrent db.Torrent, fileIndices []int) error {
	mapEntry, exists := s.cTorrentMap.Load(torrent.ID)
	if exists {
		cTorrent := mapEntry.(*torrentLib.Torrent)
		cTorrent.Drop()
		<-cTorrent.Closed()
	}

	cTorrent, err := s.client.AddTorrentFromFile(torrent.FilePath)
	if err != nil {
		return engine.ErrInternal(fmt.Sprintf("failed to add torrent from file: %s", err.Error()))
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
			torrent.Files[i].Status = db.TorrentFileDownload
			selectedFiles = append(selectedFiles, torrent.Files[i].ID)
		} else {
			unselectedFiles = append(unselectedFiles, torrent.Files[i].ID)
		}
	}

	err = s.torrentRepo.StartTorrent(torrent.ID, unselectedFiles, selectedFiles, downloadLength)
	if err != nil {
		return engine.ErrInternal(err.Error())
	}

	go s.torrentCompletionWatcher(torrent.ID, torrent.Name, torrent.Files, downloadLength, cTorrent)

	return nil
}

func (s *TorrentService) Stop(torrent db.Torrent) error {
	mapEntry, exists := s.cTorrentMap.Load(torrent.ID)
	if !exists {
		return engine.ErrNotFoundInst
	}

	cTorrent := mapEntry.(*torrentLib.Torrent)

	for i := range torrent.Files {
		torrent.Files[i].Selected = false
		torrent.Files[i].Status = db.TorrentFileIdle
	}

	cTorrent.Drop()
	<-cTorrent.Closed()

	err := s.torrentRepo.StopTorrent(torrent.ID)
	if err != nil {
		return err
	}

	return nil
}

var TorrentExport = fx.Options(fx.Provide(NewTorrentService))

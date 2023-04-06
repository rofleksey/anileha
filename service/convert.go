package service

import (
	"anileha/analyze"
	"anileha/command"
	"anileha/config"
	"anileha/db"
	"anileha/ffmpeg"
	"anileha/rest"
	"anileha/util"
	"fmt"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
	"path"
	"path/filepath"
)

type ConversionService struct {
	db               *gorm.DB
	log              *zap.Logger
	queue            *ffmpeg.Queue
	queueChan        chan ffmpeg.OutputMessage
	probeAnalyzer    *analyze.ProbeAnalyzer
	cmdProducer      *command.Producer
	fileService      *FileService
	seriesService    *SeriesService
	episodeService   *EpisodeService
	conversionFolder string
}

func NewConversionService(
	database *gorm.DB,
	probeAnalyzer *analyze.ProbeAnalyzer,
	cmdProducer *command.Producer,
	fileService *FileService,
	seriesService *SeriesService,
	episodeService *EpisodeService,
	log *zap.Logger,
	config *config.Config,
) (*ConversionService, error) {
	if err := database.Model(&db.Conversion{}).Where("status = ? OR status = ?", db.ConversionProcessing, db.ConversionCreated).Updates(db.Conversion{Status: db.ConversionError}).Error; err != nil {
		return nil, err
	}
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	conversionFolder := path.Join(workingDir, config.Data.Dir, util.ConversionSubDir)
	err = os.MkdirAll(conversionFolder, os.ModePerm)
	if err != nil {
		return nil, err
	}
	queueChan := make(chan ffmpeg.OutputMessage, 128)
	queue, err := ffmpeg.NewQueue(queueChan, log)
	if err != nil {
		return nil, err
	}
	queue.Start()
	service := &ConversionService{
		db:               database,
		probeAnalyzer:    probeAnalyzer,
		cmdProducer:      cmdProducer,
		fileService:      fileService,
		seriesService:    seriesService,
		episodeService:   episodeService,
		log:              log,
		queue:            queue,
		queueChan:        queueChan,
		conversionFolder: conversionFolder,
	}
	return service, nil
}

func (s *ConversionService) cleanUpConversion(conversion db.Conversion) {
	s.queue.Cancel(conversion.ID)
	// TODO: repeat deletion for several seconds until successful
	_ = os.Remove(conversion.VideoPath)
	_ = os.Remove(conversion.LogPath)
	if err := os.RemoveAll(conversion.OutputDir); err != nil {
		s.log.Error("failed to remove conversion output dir", zap.Uint("conversionId", conversion.ID), zap.String("file", conversion.OutputDir), zap.Error(err))
	}
}

func (s *ConversionService) GetConversionById(id uint) (*db.Conversion, error) {
	var conversion db.Conversion
	queryResult := s.db.First(&conversion, "id = ?", id)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, rest.ErrNotFoundInst
	}
	return &conversion, nil
}

func (s *ConversionService) GetAllConversions() ([]db.Conversion, error) {
	var conversions []db.Conversion
	queryResult := s.db.Find(&conversions)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return conversions, nil
}

func (s *ConversionService) GetLogsById(id uint) (*string, error) {
	var conversion db.Conversion
	queryResult := s.db.First(&conversion, "id = ?", id)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, rest.ErrNotFoundInst
	}
	logBytes, err := os.ReadFile(conversion.LogPath)
	if err != nil {
		return nil, err
	}
	logString := string(logBytes)
	return &logString, nil
}

func (s *ConversionService) GetConversionsBySeriesId(seriesId uint) ([]db.Conversion, error) {
	var conversions []db.Conversion
	queryResult := s.db.Where("series_id = ?", seriesId).Order("updated_at DESC").Find(&conversions)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return conversions, nil
}

func (s *ConversionService) GetConversionsByTorrentId(torrentId uint) ([]db.Conversion, error) {
	var conversions []db.Conversion
	queryResult := s.db.Where("torrent_id = ?", torrentId).Order("updated_at DESC").Find(&conversions)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return conversions, nil
}

func (s *ConversionService) DeleteConversionById(id uint) error {
	var conversion db.Conversion
	queryResult := s.db.First(&conversion, "id = ?", id)
	if queryResult.Error != nil {
		return queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return rest.ErrNotFoundInst
	}
	queryResult = s.db.Delete(&db.Conversion{}, id)
	if queryResult.Error != nil {
		return queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return rest.ErrNotFoundInst
	}
	go s.cleanUpConversion(conversion)
	return nil
}

func (s *ConversionService) queueWorker() {
	for update := range s.queueChan {
		switch msg := update.Msg.(type) {
		case ffmpeg.QueueSignalStarted:
			if err := s.db.Model(&db.Conversion{}).Where("id = ?", update.ID).Updates(db.Conversion{Status: db.ConversionProcessing}).Error; err != nil {
				s.log.Error("failed to update db on conversion start", zap.Uint("conversionId", update.ID), zap.Error(err))
				continue
			}
		case string:
			s.log.Info(msg, zap.Uint("conversionId", update.ID))
		case util.Progress:
			if err := s.db.Model(&db.Conversion{}).Where("id = ?", update.ID).Updates(db.Conversion{Progress: msg}).Error; err != nil {
				s.log.Error("failed to update db on conversion progress", zap.Uint("conversionId", update.ID), zap.Error(err))
				continue
			}
			//s.log.Info("conversion progress", zap.Uint("conversionId", update.ID), zap.Float64("progress", msg.Progress), zap.Float64("eta", msg.Eta), zap.Float64("elapsed", msg.Elapsed))
		case ffmpeg.CommandSignalEnd:
			if msg.Err == nil {
				finishedConversionId := update.ID
				go func() {
					conversion, err := s.GetConversionById(finishedConversionId)
					if err != nil {
						s.log.Error("failed to get conversion by id", zap.Uint("conversionId", finishedConversionId), zap.Error(err))
						return
					}
					episode, err := s.episodeService.CreateEpisodeFromConversion(conversion)
					if err != nil {
						s.log.Error("failed to create episode", zap.Uint("conversionId", finishedConversionId), zap.Error(err))
						return
					}
					if err := s.db.Model(&db.Conversion{}).Where("id = ?", finishedConversionId).Updates(db.Conversion{Status: db.ConversionReady, EpisodeId: &episode.ID, Progress: util.Progress{Progress: 100}}).Error; err != nil {
						s.log.Error("failed to update db on conversion finish", zap.Uint("conversionId", finishedConversionId), zap.Error(err))
						return
					}
				}()
			} else {
				var newStatus db.ConversionStatus
				if msg.Err == util.ErrCancelled {
					newStatus = db.ConversionCancelled
				} else {
					newStatus = db.ConversionError
				}
				if err := s.db.Model(&db.Conversion{}).Where("id = ?", update.ID).Updates(db.Conversion{Status: newStatus}).Error; err != nil {
					s.log.Error("failed to update db on conversion error", zap.Uint("conversionId", update.ID), zap.Error(err))
					continue
				}
			}
		}
	}
}

func (s *ConversionService) prepareConversion(
	torrent db.Torrent,
	torrentFile db.TorrentFile,
	outputDir string,
	videoPath string,
	logsPath string,
	command *ffmpeg.Command,
	durationSec int,
) (*db.Conversion, error) {
	conversionName := fmt.Sprintf("%s - %s", torrent.Name, torrentFile.TorrentPath)
	episodeName := torrentFile.TorrentPath
	conversion := db.Conversion{
		SeriesId:         torrent.SeriesId,
		TorrentId:        torrent.ID,
		Name:             conversionName,
		EpisodeName:      episodeName,
		OutputDir:        outputDir,
		VideoPath:        videoPath,
		LogPath:          logsPath,
		Command:          command.String(),
		VideoDurationSec: durationSec,
	}
	queryResult := s.db.Create(&conversion)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, rest.ErrCreationFailed
	}
	return &conversion, nil
}

func (s *ConversionService) StartConversion(torrent db.Torrent, torrentFiles []db.TorrentFile, prefsArr []command.Preferences) error {
	for i := range torrentFiles {
		folder, err := s.fileService.GenFolderPath(s.conversionFolder)
		if err != nil {
			return err
		}

		videoPath := filepath.Join(folder, "video.mp4")
		logsPath := filepath.Join(folder, "log.txt")

		probe, err := s.probeAnalyzer.Probe(*torrentFiles[i].ReadyPath)
		if err != nil {
			return rest.ErrInternal(fmt.Sprintf("probe of file %s failed: %s", *torrentFiles[i].ReadyPath, err.Error()))
		}

		ffmpegCmd, err := s.cmdProducer.GetFFmpegCommand(*torrentFiles[i].ReadyPath, videoPath, logsPath, probe, prefsArr[i])
		if err != nil {
			return rest.ErrInternal(fmt.Sprintf("failed to get ffmpeg command for file %s: %s", *torrentFiles[i].ReadyPath, err.Error()))
		}

		conversion, err := s.prepareConversion(torrent, torrentFiles[i], folder, videoPath, logsPath, ffmpegCmd, probe.Video.DurationSec)
		if err != nil {
			return rest.ErrInternal(fmt.Sprintf("failed to prepare conversion for file %s: %s", *torrentFiles[i].ReadyPath, err.Error()))
		}

		err = os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			return rest.ErrInternal(fmt.Sprintf("failed to create folder for file %s: %s", *torrentFiles[i].ReadyPath, err.Error()))
		}

		s.queue.Enqueue(conversion.ID, ffmpegCmd)
	}
	return nil
}

func (s *ConversionService) StopConversion(conversionId uint) error {
	s.queue.Cancel(conversionId)
	err := s.db.Model(&db.Conversion{}).Where("id = ?", conversionId).Updates(db.Conversion{Status: db.ConversionCancelled}).Error
	if err != nil {
		return err
	}
	return nil
}

func startQueueWorker(service *ConversionService) {
	go service.queueWorker()
}

var ConversionServiceExport = fx.Options(fx.Provide(NewConversionService), fx.Invoke(startQueueWorker))

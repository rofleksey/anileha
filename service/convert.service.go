package service

import (
	"anileha/analyze"
	"anileha/config"
	"anileha/db"
	"anileha/ffmpeg"
	"anileha/util"
	"fmt"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
	"path"
	"path/filepath"
)

// TODO: delete conversion video on failure or on startup error check

type ConversionService struct {
	db               *gorm.DB
	log              *zap.Logger
	queue            *ffmpeg.Queue
	queueChan        chan ffmpeg.OutputMessage
	fileService      *FileService
	seriesService    *SeriesService
	episodeService   *EpisodeService
	conversionFolder string
}

func NewConversionService(
	lifecycle fx.Lifecycle,
	database *gorm.DB,
	fileService *FileService,
	seriesService *SeriesService,
	episodeService *EpisodeService,
	log *zap.Logger,
	config *config.Config,
) (*ConversionService, error) {
	if err := database.Model(&db.Conversion{}).Where("status = ? OR status = ?", db.CONVERSION_PROCESSING, db.CONVERSION_CREATED).Updates(db.Conversion{Status: db.CONVERSION_ERROR}).Error; err != nil {
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
	queue, err := ffmpeg.NewQueue(config.Conversion.Parallelism, queueChan)
	if err != nil {
		return nil, err
	}
	queue.Start()
	service := &ConversionService{
		db:               database,
		fileService:      fileService,
		seriesService:    seriesService,
		episodeService:   episodeService,
		log:              log,
		queue:            queue,
		queueChan:        queueChan,
		conversionFolder: conversionFolder,
	}
	go service.queueWorker()
	return service, nil
}

func (s *ConversionService) GetConversionById(id uint) (*db.Conversion, error) {
	var conversion db.Conversion
	queryResult := s.db.First(&conversion, "id = ?", id)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, util.ErrNotFound
	}
	return &conversion, nil
}

func (s *ConversionService) GetConversionsBySeriesId(seriesId uint) ([]db.Conversion, error) {
	var conversions []db.Conversion
	queryResult := s.db.Where("series_id = ?", seriesId).Order("updated_at DESC").Find(&conversions)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, util.ErrNotFound
	}
	return conversions, nil
}

func (s *ConversionService) queueWorker() {
	for update := range s.queueChan {
		switch msg := update.Msg.(type) {
		case ffmpeg.QueueSignalStarted:
			if err := s.db.Model(&db.Conversion{}).Where("id = ?", update.ID).Updates(db.Conversion{Status: db.CONVERSION_PROCESSING}).Error; err != nil {
				s.log.Error("failed to update db on conversion start", zap.Uint("conversionId", update.ID), zap.Error(err))
				continue
			}
		case string:
			s.log.Info(msg, zap.Uint("conversionId", update.ID))
		case db.Progress:
			if err := s.db.Model(&db.Conversion{}).Where("id = ?", update.ID).Updates(db.Conversion{Progress: msg}).Error; err != nil {
				s.log.Error("failed to update db on conversion progress", zap.Uint("conversionId", update.ID), zap.Error(err))
				continue
			}
			//s.log.Info("conversion progress", zap.Uint("conversionId", update.ID), zap.Float64("progress", msg.Progress), zap.Float64("eta", msg.Eta), zap.Float64("elapsed", msg.Elapsed))
		case ffmpeg.CommandSignalEnd:
			if msg.Err == nil {
				conversion, err := s.GetConversionById(update.ID)
				if err != nil {
					s.log.Error("failed to get conversion by id", zap.Uint("conversionId", update.ID), zap.Error(err))
					continue
				}
				episode, err := s.episodeService.CreateEpisodeFromConversion(conversion)
				if err != nil {
					s.log.Error("failed to create episode", zap.Uint("conversionId", update.ID), zap.Error(err))
					continue
				}
				if err := s.db.Model(&db.Conversion{}).Where("id = ?", update.ID).Updates(db.Conversion{Status: db.CONVERSION_READY, EpisodeId: &episode.ID, Progress: db.Progress{Progress: 1}}).Error; err != nil {
					s.log.Error("failed to update db on conversion finish", zap.Uint("conversionId", update.ID), zap.Error(err))
					continue
				}
			} else {
				if err := s.db.Model(&db.Conversion{}).Where("id = ?", update.ID).Updates(db.Conversion{Status: db.CONVERSION_ERROR}).Error; err != nil {
					s.log.Error("failed to update db on conversion error", zap.Uint("conversionId", update.ID), zap.Error(err))
					continue
				}
			}
		}
	}
}

func (s *ConversionService) prepareCommand(inputFile string, outputPath string, logsPath string, analysis *analyze.Result) (*ffmpeg.Command, error) {
	// TODO: research settings
	command := ffmpeg.NewCommand(inputFile, analysis.Video.DurationSec, outputPath)
	command.AddKeyValue("-acodec", "aac", ffmpeg.OptionOutput)
	command.AddKeyValue("-b:a", "196k", ffmpeg.OptionOutput)
	command.AddKeyValue("-ac", "2", ffmpeg.OptionOutput)
	command.AddKeyValue("-vcodec", "libx264", ffmpeg.OptionOutput)
	command.AddKeyValue("-crf", "18", ffmpeg.OptionOutput)
	command.AddKeyValue("-tune", "animation", ffmpeg.OptionOutput)  // this is bad?
	command.AddKeyValue("-pix_fmt", "yuv420p", ffmpeg.OptionOutput) // yuv420p10le?
	command.AddKeyValue("-preset", "slow", ffmpeg.OptionOutput)
	command.AddKeyValue("-f", "mp4", ffmpeg.OptionOutput)
	command.AddKeyValue("-movflags", "+faststart", ffmpeg.OptionPostOutput)
	if analysis.Sub != nil {
		switch analysis.Sub.Type {
		case analyze.SubsText:
			command.AddKeyValue("-filter_complex", fmt.Sprintf("[0:v]subtitles=f='%s':si=%d[vo]", inputFile, analysis.Sub.RelativeIndex), ffmpeg.OptionOutput)
			command.AddKeyValue("-map", "[vo]", ffmpeg.OptionPostOutput)
		case analyze.SubsPicture:
			command.AddKeyValue("-filter_complex", fmt.Sprintf("[0:v][0:s:%d]overlay[vo]", analysis.Sub.RelativeIndex), ffmpeg.OptionOutput)
			command.AddKeyValue("-map", "[vo]", ffmpeg.OptionPostOutput)
		default:
			return nil, util.ErrUnsupportedSubs
		}
	}
	if analysis.Audio != nil {
		command.AddKeyValue("-map", fmt.Sprintf("0:a:%d", analysis.Audio.RelativeIndex), ffmpeg.OptionOutput)
	}
	command.WriteLogsTo(logsPath)
	return command, nil
}

func (s *ConversionService) prepareConversion(seriesId uint, torrentName string, torrentFile db.TorrentFile, videoPath string, logsPath string, command *ffmpeg.Command, durationSec uint64) (*db.Conversion, error) {
	var conversionName string
	if torrentFile.Season != "" {
		conversionName = fmt.Sprintf("%s - %s - %s", torrentName, torrentFile.Season, torrentFile.Episode)
	} else {
		conversionName = fmt.Sprintf("%s - %s", torrentName, torrentFile.Episode)
	}

	var episodeName string
	if torrentFile.Season != "" {
		episodeName = fmt.Sprintf("%s ??? %s", torrentFile.Season, torrentFile.Episode)
	} else {
		episodeName = torrentFile.Episode
	}

	conversion := db.NewConversion(seriesId, torrentFile.ID, conversionName, episodeName, videoPath, logsPath, command.String(), durationSec)
	queryResult := s.db.Create(&conversion)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, util.ErrCreationFailed
	}
	return &conversion, nil
}

func (s *ConversionService) StartConversion(seriesID uint, torrentName string, torrentFiles []db.TorrentFile, analysisArr []*analyze.Result) error {
	// TODO: parse name
	for i := range torrentFiles {
		folder, err := s.fileService.GenFolderPath(s.conversionFolder)
		if err != nil {
			return err
		}
		videoPath := filepath.Join(folder, "video.mp4")
		logsPath := filepath.Join(folder, "log.txt")
		command, err := s.prepareCommand(*torrentFiles[i].ReadyPath, videoPath, logsPath, analysisArr[i])
		if err != nil {
			return err
		}
		conversion, err := s.prepareConversion(seriesID, torrentName, torrentFiles[i], videoPath, logsPath, command, analysisArr[i].Video.DurationSec)
		if err != nil {
			return err
		}
		err = os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			return err
		}
		s.queue.Enqueue(conversion.ID, command)
	}
	return nil
}

func (s *ConversionService) StopConversion(conversionId uint) error {
	s.queue.Cancel(conversionId)
	err := s.db.Model(&db.Conversion{}).Where("id = ?", conversionId).Updates(db.Conversion{Status: db.CONVERSION_CANCELLED}).Error
	if err != nil {
		return err
	}
	return nil
}

var ConversionServiceExport = fx.Options(fx.Provide(NewConversionService))

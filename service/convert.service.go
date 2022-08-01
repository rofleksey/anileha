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
	db *gorm.DB,
	fileService *FileService,
	seriesService *SeriesService,
	episodeService *EpisodeService,
	log *zap.Logger,
	config *config.Config,
) (*ConversionService, error) {
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
		db:               db,
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
	queryResult := s.db.Find(&conversions, "seriesId = ?", seriesId)
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
			s.log.Info("conversion progress", zap.Uint("conversionId", update.ID), zap.Float64("progress", msg.Progress), zap.Float64("eta", msg.Eta), zap.Float64("elapsed", msg.Elapsed))
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

func (s *ConversionService) prepareConversion(seriesId uint, torrentFile *db.TorrentFile, videoPath string, logsPath string, command *ffmpeg.Command, durationSec uint64) (*db.Conversion, error) {
	conversion := db.NewConversion(seriesId, torrentFile.ID, filepath.Base(torrentFile.TorrentPath), videoPath, logsPath, command.String(), durationSec)
	queryResult := s.db.Create(&conversion)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, util.ErrCreationFailed
	}
	return &conversion, nil
}

func (s *ConversionService) StartConversion(series *db.Series, torrentFile *db.TorrentFile, analysis *analyze.Result) error {
	// TODO: parse name
	folder, err := s.fileService.GenFolderPath(s.conversionFolder)
	if err != nil {
		return err
	}
	videoPath := filepath.Join(folder, "video.mp4")
	logsPath := filepath.Join(folder, "log.txt")
	command, err := s.prepareCommand(*torrentFile.ReadyPath, videoPath, logsPath, analysis)
	if err != nil {
		return err
	}
	conversion, err := s.prepareConversion(series.ID, torrentFile, videoPath, logsPath, command, analysis.Video.DurationSec)
	if err != nil {
		return err
	}
	err = os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		return err
	}
	s.queue.Enqueue(conversion.ID, command)
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

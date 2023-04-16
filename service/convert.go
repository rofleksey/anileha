package service

import (
	"anileha/analyze"
	"anileha/command"
	"anileha/config"
	"anileha/db"
	"anileha/db/repo"
	"anileha/ffmpeg"
	"anileha/rest"
	"anileha/util"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type ConversionService struct {
	conversionRepo   *repo.ConversionRepo
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
	conversionRepo *repo.ConversionRepo,
	probeAnalyzer *analyze.ProbeAnalyzer,
	cmdProducer *command.Producer,
	fileService *FileService,
	seriesService *SeriesService,
	episodeService *EpisodeService,
	log *zap.Logger,
	config *config.Config,
) (*ConversionService, error) {
	if err := conversionRepo.ResetProcessing(); err != nil {
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
		conversionRepo:   conversionRepo,
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

func (s *ConversionService) GetById(id uint) (*db.Conversion, error) {
	conversion, err := s.conversionRepo.GetById(id)
	if err != nil {
		return nil, rest.ErrInternal(err.Error())
	}
	if conversion == nil {
		return nil, rest.ErrNotFoundInst
	}
	return conversion, nil
}

func (s *ConversionService) GetAll() ([]db.Conversion, error) {
	conversions, err := s.conversionRepo.GetAll()
	if err != nil {
		return nil, rest.ErrInternal(err.Error())
	}
	return conversions, nil
}

func (s *ConversionService) GetLogsById(id uint) ([]byte, error) {
	conversion, err := s.conversionRepo.GetById(id)
	if err != nil {
		return nil, rest.ErrInternal(err.Error())
	}
	if conversion == nil {
		return nil, rest.ErrNotFoundInst
	}
	logBytes, err := os.ReadFile(conversion.LogPath)
	if err != nil {
		return nil, err
	}
	return logBytes, nil
}

func (s *ConversionService) GetBySeriesId(seriesId uint) ([]db.Conversion, error) {
	conversions, err := s.conversionRepo.GetBySeriesId(seriesId)
	if err != nil {
		return nil, rest.ErrInternal(err.Error())
	}
	return conversions, nil
}

func (s *ConversionService) GetByTorrentId(torrentId uint) ([]db.Conversion, error) {
	conversions, err := s.conversionRepo.GetByTorrentId(torrentId)
	if err != nil {
		return nil, rest.ErrInternal(err.Error())
	}
	return conversions, nil
}

func (s *ConversionService) DeleteConversionById(id uint) error {
	conversion, err := s.conversionRepo.GetById(id)
	if err != nil {
		return rest.ErrInternal(err.Error())
	}
	if conversion == nil {
		return rest.ErrNotFoundInst
	}
	count, err := s.conversionRepo.DeleteById(id)
	if err != nil {
		return rest.ErrInternal(err.Error())
	}
	if count == 0 {
		return rest.ErrDeletionFailed
	}
	go s.cleanUpConversion(*conversion)
	return nil
}

func (s *ConversionService) queueWorker() {
	for update := range s.queueChan {
		switch msg := update.Msg.(type) {
		case ffmpeg.QueueSignalStarted:
			if err := s.conversionRepo.SetStatus(update.ID, db.ConversionProcessing); err != nil {
				s.log.Error("failed to update db on conversion start", zap.Uint("conversionId", update.ID), zap.Error(err))
				continue
			}
		case string:
			s.log.Info(msg, zap.Uint("conversionId", update.ID))
		case util.Progress:
			if err := s.conversionRepo.SetProgress(update.ID, msg); err != nil {
				s.log.Error("failed to update db on conversion progress", zap.Uint("conversionId", update.ID), zap.Error(err))
				continue
			}
			//s.log.Info("conversion progress", zap.Uint("conversionId", update.ID), zap.Float64("progress", msg.Progress), zap.Float64("eta", msg.Eta), zap.Float64("elapsed", msg.Elapsed))
		case ffmpeg.CommandSignalEnd:
			if msg.Err == nil {
				finishedConversionId := update.ID
				go func() {
					conversion, err := s.GetById(finishedConversionId)
					if err != nil {
						s.log.Error("failed to get conversion by id", zap.Uint("conversionId", finishedConversionId), zap.Error(err))
						return
					}
					episode, err := s.episodeService.CreateFromConversion(conversion)
					if err != nil {
						s.log.Error("failed to create episode", zap.Uint("conversionId", finishedConversionId), zap.Error(err))
						return
					}
					if err := s.conversionRepo.SetFinish(finishedConversionId, episode.ID); err != nil {
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
				if err := s.conversionRepo.SetStatus(update.ID, newStatus); err != nil {
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
	episode string,
	season string,
	outputDir string,
	videoPath string,
	logsPath string,
	command *ffmpeg.Command,
	durationSec int,
) (*db.Conversion, error) {
	conversionName := fmt.Sprintf("%s - %s", torrent.Name, torrentFile.TorrentPath)
	episodeNameSlice := make([]string, 0, 3)

	if torrent.Series != nil {
		episodeNameSlice = append(episodeNameSlice, torrent.Series.Title)
	}

	if season != "" {
		if _, err := strconv.Atoi(season); err == nil {
			episodeNameSlice = append(episodeNameSlice, "S"+season)
		} else {
			episodeNameSlice = append(episodeNameSlice, season)
		}
	}

	episodeNameSlice = append(episodeNameSlice, episode)
	episodeName := strings.Join(episodeNameSlice, " - ")

	conversion := db.Conversion{
		SeriesId:         torrent.SeriesId,
		TorrentId:        &torrent.ID,
		TorrentFileId:    &torrentFile.ID,
		Name:             conversionName,
		EpisodeName:      episodeName,
		EpisodeString:    episode,
		SeasonString:     season,
		OutputDir:        outputDir,
		VideoPath:        videoPath,
		LogPath:          logsPath,
		Command:          command.String(),
		Status:           db.ConversionCreated,
		VideoDurationSec: durationSec,
	}
	_, err := s.conversionRepo.Create(&conversion)
	if err != nil {
		return nil, err
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
		prefs := prefsArr[i]

		if prefs.Sub.ExternalFile != "" {
			index := pie.FindFirstUsing(torrent.Files, func(file db.TorrentFile) bool {
				return file.TorrentPath == prefs.Sub.ExternalFile
			})
			prefs.Sub.ExternalFile = *torrent.Files[index].ReadyPath
		}

		if prefs.Audio.ExternalFile != "" {
			index := pie.FindFirstUsing(torrent.Files, func(file db.TorrentFile) bool {
				return file.TorrentPath == prefs.Audio.ExternalFile
			})
			prefs.Audio.ExternalFile = *torrent.Files[index].ReadyPath
		}

		probe, err := s.probeAnalyzer.Probe(*torrentFiles[i].ReadyPath)
		if err != nil {
			return rest.ErrInternal(fmt.Sprintf("probe of file %s failed: %s", *torrentFiles[i].ReadyPath, err.Error()))
		}

		ffmpegCmd, err := s.cmdProducer.GetFFmpegCommand(*torrentFiles[i].ReadyPath, videoPath, logsPath, probe, prefs)
		if err != nil {
			return rest.ErrInternal(fmt.Sprintf("failed to get ffmpeg command for file %s: %s", *torrentFiles[i].ReadyPath, err.Error()))
		}

		conversion, err := s.prepareConversion(torrent, torrentFiles[i], prefs.Episode, prefs.Season, folder, videoPath, logsPath, ffmpegCmd, probe.Video.DurationSec)
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

	return s.conversionRepo.SetStatus(conversionId, db.ConversionCancelled)
}

func startQueueWorker(service *ConversionService) {
	go service.queueWorker()
}

var ConversionServiceExport = fx.Options(fx.Provide(NewConversionService), fx.Invoke(startQueueWorker))

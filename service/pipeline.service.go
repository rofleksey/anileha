package service

import (
	"anileha/analyze"
	"anileha/config"
	"anileha/db"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PipelineFacade struct {
	Channel chan interface{}
}

type PipelineMessageDeleteSeries struct {
	SeriesId uint
	Result   chan error
}

type PipelineMessageDeleteTorrent struct {
	TorrentId uint
	Result    chan error
}

type PipelineMessageTorrentFinished struct {
	TorrentId uint
}

type PipelineService struct {
	db                *gorm.DB
	log               *zap.Logger
	fileService       *FileService
	seriesService     *SeriesService
	torrentService    *TorrentService
	conversionService *ConversionService
	episodeService    *EpisodeService
	thumbService      *ThumbService
	analyzer          *analyze.ProbeAnalyzer
}

func NewPipelineService(
	lifecycle fx.Lifecycle,
	log *zap.Logger,
	config *config.Config,
	database *gorm.DB,
	fileService *FileService,
	seriesService *SeriesService,
	torrentService *TorrentService,
	conversionService *ConversionService,
	episodeService *EpisodeService,
	thumbService *ThumbService,
	analyzer *analyze.ProbeAnalyzer,
) *PipelineService {
	service := PipelineService{
		db:                database,
		log:               log,
		fileService:       fileService,
		seriesService:     seriesService,
		torrentService:    torrentService,
		conversionService: conversionService,
		episodeService:    episodeService,
		thumbService:      thumbService,
		analyzer:          analyzer,
	}
	return &service
}

func NewPipelineFacade() *PipelineFacade {
	return &PipelineFacade{
		Channel: make(chan interface{}, 16),
	}
}

func (s *PipelineService) messagesWorker(facade *PipelineFacade) {
	for {
		msg := <-facade.Channel
		switch casted := msg.(type) {
		case PipelineMessageDeleteSeries:
			go s.pipelineDeleteSeries(casted)
		case PipelineMessageDeleteTorrent:
			go s.pipelineDeleteTorrent(casted)
		case PipelineMessageTorrentFinished:
			go s.pipelineTorrentFinished(casted)
		default:
			s.log.Error("invalid message type received")
		}
	}
}

func (s *PipelineService) pipelineDeleteSeries(msg PipelineMessageDeleteSeries) {
	s.log.Info("delete series pipeline started", zap.Uint("seriesId", msg.SeriesId))
	conversions, err := s.conversionService.GetConversionsBySeriesId(msg.SeriesId)
	if err != nil {
		s.log.Warn("failed to delete conversions for series", zap.Uint("seriesId", msg.SeriesId))
		msg.Result <- err
		return
	} else {
		for _, c := range conversions {
			cId := c.ID
			go func() {
				if err := s.conversionService.DeleteConversionById(cId); err != nil {
					s.log.Warn("failed to delete conversion for series", zap.Uint("conversionId", cId), zap.Uint("seriesId", msg.SeriesId), zap.Error(err))
				}
			}()
		}
	}
	episodes, err := s.episodeService.GetEpisodesBySeriesId(msg.SeriesId)
	if err != nil {
		s.log.Warn("failed to delete episodes for series", zap.Uint("seriesId", msg.SeriesId))
		msg.Result <- err
		return
	} else {
		for _, c := range episodes {
			cId := c.ID
			go func() {
				if err := s.episodeService.DeleteEpisodeById(cId); err != nil {
					s.log.Warn("failed to delete episode for series", zap.Uint("episodeId", cId), zap.Uint("seriesId", msg.SeriesId), zap.Error(err))
				}
			}()
		}
	}
	torrents, err := s.torrentService.GetTorrentsBySeriesId(msg.SeriesId)
	if err != nil {
		s.log.Warn("failed to delete torrents for series", zap.Uint("seriesId", msg.SeriesId), zap.Error(err))
		msg.Result <- err
		return
	} else {
		for _, t := range torrents {
			tId := t.ID
			go func() {
				if err := s.torrentService.DeleteTorrentById(tId); err != nil {
					s.log.Warn("failed to delete torrent for series", zap.Uint("torrentId", tId), zap.Uint("seriesId", msg.SeriesId), zap.Error(err))
				}
			}()
		}
	}
	if err := s.seriesService.DeleteSeriesById(msg.SeriesId); err != nil {
		s.log.Error("delete series pipeline failed", zap.Uint("seriesId", msg.SeriesId), zap.Error(err))
		msg.Result <- err
		return
	}
	msg.Result <- nil
	s.log.Info("delete series pipeline success", zap.Uint("seriesId", msg.SeriesId))
}

func (s *PipelineService) pipelineDeleteTorrent(msg PipelineMessageDeleteTorrent) {
	s.log.Info("delete torrent pipeline started", zap.Uint("torrentId", msg.TorrentId))
	conversions, err := s.conversionService.GetConversionsByTorrentId(msg.TorrentId)
	if err != nil {
		s.log.Warn("failed to stop conversions for torrent", zap.Uint("torrentId", msg.TorrentId))
		msg.Result <- err
		return
	} else {
		for _, c := range conversions {
			cId := c.ID
			go func() {
				if err := s.conversionService.DeleteConversionById(cId); err != nil {
					s.log.Warn("failed to delete conversion for torrent", zap.Uint("conversionId", cId), zap.Uint("torrentId", msg.TorrentId), zap.Error(err))
				}
			}()
		}
	}
	if err := s.torrentService.DeleteTorrentById(msg.TorrentId); err != nil {
		s.log.Error("delete torrent pipeline failed", zap.Uint("torrentId", msg.TorrentId), zap.Error(err))
		msg.Result <- err
		return
	}
	msg.Result <- nil
	s.log.Info("delete torrent pipeline success", zap.Uint("torrentId", msg.TorrentId))
}

func (s *PipelineService) pipelineTorrentFinished(msg PipelineMessageTorrentFinished) {
	s.log.Info("torrent finished pipeline started", zap.Uint("torrentId", msg.TorrentId))
	torrent, err := s.torrentService.GetTorrentById(msg.TorrentId)
	if err != nil {
		s.log.Warn("torrent finished pipeline failed", zap.Uint("torrentId", msg.TorrentId), zap.Error(err))
		return
	}
	torrentFiles := make([]db.TorrentFile, 0, len(torrent.Files))
	analysisArr := make([]*analyze.Result, 0, len(torrent.Files))
	for _, file := range torrent.Files {
		if file.Selected {
			if file.Status != db.TORRENT_FILE_READY {
				continue
			}
			if file.ReadyPath == nil {
				continue
			}
			analysis, err := s.analyzer.Analyze(*file.ReadyPath, true)
			if err != nil {
				s.log.Warn("analysis error", zap.Uint("torrentId", msg.TorrentId), zap.Uint("fileId", file.ID), zap.Error(err))
				continue
			}
			torrentFiles = append(torrentFiles, file)
			analysisArr = append(analysisArr, analysis)
		}
	}
	if len(torrentFiles) > 0 {
		err = s.conversionService.StartConversion(*torrent, torrentFiles, analysisArr)
		if err != nil {
			s.log.Warn("torrent finished pipeline failed", zap.Uint("torrentId", msg.TorrentId), zap.Error(err))
			return
		}
	}
	s.log.Info("torrent finished pipeline success", zap.Uint("torrentId", msg.TorrentId))
}

func startMessagesWorker(service *PipelineService, facade *PipelineFacade) {
	go service.messagesWorker(facade)
}

var PipelineServiceExport = fx.Options(fx.Provide(NewPipelineFacade), fx.Provide(NewPipelineService), fx.Invoke(startMessagesWorker))
package service

import (
	"anileha/config"
	"anileha/db"
	"anileha/db/repo"
	"anileha/ffmpeg/analyze"
	"anileha/rest/engine"
	"anileha/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"os"
	"path"
	"path/filepath"
)

type EpisodeService struct {
	episodeRepo   *repo.EpisodeRepo
	log           *zap.Logger
	analyzer      *analyze.ProbeAnalyzer
	fileService   *FileService
	thumbService  *ThumbService
	episodeFolder string
}

func NewEpisodeService(
	lifecycle fx.Lifecycle,
	episodeRepo *repo.EpisodeRepo,
	fileService *FileService,
	thumbService *ThumbService,
	analyzer *analyze.ProbeAnalyzer,
	log *zap.Logger,
	config *config.Config,
) (*EpisodeService, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	episodeFolder := path.Join(workingDir, config.Data.Dir, util.EpisodeSubDir)
	err = os.MkdirAll(episodeFolder, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return &EpisodeService{
		episodeRepo:   episodeRepo,
		analyzer:      analyzer,
		fileService:   fileService,
		thumbService:  thumbService,
		log:           log,
		episodeFolder: episodeFolder,
	}, nil
}

func (s *EpisodeService) cleanUpEpisode(episode db.Episode) {
	episode.Thumb.Delete()
	if err := os.Remove(episode.Path); err != nil {
		s.log.Error("failed to remove episode file", zap.Uint("episodeId", episode.ID), zap.String("file", episode.Path), zap.Error(err))
	}
}

func (s *EpisodeService) CreateFromConversion(conversion *db.Conversion) (*db.Episode, error) {
	// TODO: gen thumbnail
	episodePath, err := s.fileService.GenFilePath(s.episodeFolder, conversion.VideoPath)
	if err != nil {
		return nil, err
	}
	err = os.Rename(conversion.VideoPath, episodePath)
	if err != nil {
		return nil, err
	}
	stat, err := os.Stat(episodePath)
	if err != nil {
		return nil, err
	}
	thumb, err := s.thumbService.CreateForVideo(episodePath, conversion.VideoDurationSec)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/%s", util.EpisodeRoute, filepath.Base(episodePath))
	episode := db.Episode{
		SeriesId:    conversion.SeriesId,
		Title:       conversion.EpisodeName,
		Episode:     conversion.EpisodeString,
		Season:      conversion.SeasonString,
		Length:      uint64(stat.Size()),
		DurationSec: conversion.VideoDurationSec,
		Path:        episodePath,
		Thumb:       thumb,
		Url:         url,
	}

	id, err := s.episodeRepo.Create(&episode)
	if err != nil {
		thumb.Delete()
		return nil, engine.ErrInternal(err.Error())
	}

	episode.ID = id

	return &episode, nil
}

func (s *EpisodeService) CreateManually(seriesId *uint, tempFilePath string, title string, episodeStr string, seasonStr string) (*db.Episode, error) {
	// TODO: gen thumbnail
	episodePath, err := s.fileService.GenFilePath(s.episodeFolder, tempFilePath)
	if err != nil {
		return nil, err
	}
	err = os.Rename(tempFilePath, episodePath)
	if err != nil {
		return nil, err
	}
	stat, err := os.Stat(episodePath)
	if err != nil {
		return nil, err
	}
	durationSec, err := s.analyzer.GetVideoDurationSec(episodePath)
	if err != nil {
		return nil, err
	}
	thumb, err := s.thumbService.CreateForVideo(episodePath, durationSec)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/%s", util.EpisodeRoute, filepath.Base(episodePath))

	episode := db.Episode{
		SeriesId:    seriesId,
		Title:       title,
		Episode:     episodeStr,
		Season:      seasonStr,
		Length:      uint64(stat.Size()),
		DurationSec: durationSec,
		Path:        episodePath,
		Thumb:       thumb,
		Url:         url,
	}

	id, err := s.episodeRepo.Create(&episode)
	if err != nil {
		thumb.Delete()
		return nil, engine.ErrInternal(err.Error())
	}

	episode.ID = id

	return &episode, nil
}

func (s *EpisodeService) GetById(id uint) (*db.Episode, error) {
	episode, err := s.episodeRepo.GetById(id)
	if err != nil {
		return nil, engine.ErrInternal(err.Error())
	}
	if episode == nil {
		return nil, engine.ErrNotFoundInst
	}
	return episode, nil
}

func (s *EpisodeService) GetBySeriesId(seriesId uint) ([]db.Episode, error) {
	episodes, err := s.episodeRepo.GetBySeriesId(seriesId)
	if err != nil {
		return nil, engine.ErrInternal(err.Error())
	}
	return episodes, nil
}

func (s *EpisodeService) RefreshThumb(id uint) error {
	episode, err := s.episodeRepo.GetById(id)
	if err != nil {
		return engine.ErrInternal(err.Error())
	}
	if episode == nil {
		return engine.ErrNotFoundInst
	}

	oldThumb := episode.Thumb

	newThumb, err := s.thumbService.CreateForVideo(episode.Path, episode.DurationSec)
	if err != nil {
		return engine.ErrInternal(err.Error())
	}

	if err := s.episodeRepo.SetThumb(id, newThumb); err != nil {
		return engine.ErrInternal(err.Error())
	}

	oldThumb.Delete()

	return nil
}

func (s *EpisodeService) DeleteById(id uint) error {
	episode, err := s.episodeRepo.GetById(id)
	if err != nil {
		return engine.ErrInternal(err.Error())
	}
	if episode == nil {
		return engine.ErrNotFoundInst
	}

	rows, err := s.episodeRepo.DeleteById(id)
	if err != nil {
		return engine.ErrInternal(err.Error())
	}
	if rows == 0 {
		return engine.ErrNotFoundInst
	}

	go s.cleanUpEpisode(*episode)
	return nil
}

func registerStaticEpisodes(engine *gin.Engine, config *config.Config) {
	engine.Static(util.EpisodeRoute, path.Join(config.Data.Dir, util.EpisodeSubDir))
}

var EpisodeExport = fx.Options(fx.Provide(NewEpisodeService), fx.Invoke(registerStaticEpisodes))

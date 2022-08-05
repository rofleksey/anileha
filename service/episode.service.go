package service

import (
	"anileha/config"
	"anileha/db"
	"anileha/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
	"path"
	"path/filepath"
)

type EpisodeService struct {
	db            *gorm.DB
	log           *zap.Logger
	fileService   *FileService
	episodeFolder string
}

func NewEpisodeService(
	lifecycle fx.Lifecycle,
	db *gorm.DB,
	fileService *FileService,
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
		db:            db,
		fileService:   fileService,
		log:           log,
		episodeFolder: episodeFolder,
	}, nil
}

func (s *EpisodeService) cleanUpEpisode(episode db.Episode) {
	if err := os.Remove(episode.Path); err != nil {
		s.log.Error("failed to remove episode file", zap.Uint("episodeId", episode.ID), zap.String("file", episode.Path), zap.Error(err))
	}
}

func (s *EpisodeService) CreateEpisodeFromConversion(conversion *db.Conversion) (*db.Episode, error) {
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
	url := fmt.Sprintf("%s/%s", util.EpisodeRoute, filepath.Base(episodePath))
	episode := db.NewEpisode(conversion.SeriesId, conversion.ID, conversion.EpisodeName, nil, uint64(stat.Size()), conversion.VideoDurationSec, episodePath, url)
	queryResult := s.db.Create(&episode)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, util.ErrCreationFailed
	}
	return &episode, nil
}

func (s *EpisodeService) GetEpisodeById(id uint) (*db.Episode, error) {
	var episode db.Episode
	queryResult := s.db.First(&episode, "id = ?", id)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, util.ErrNotFound
	}
	return &episode, nil
}

func (s *EpisodeService) GetEpisodesBySeriesId(seriesId uint) ([]db.Episode, error) {
	var episodes []db.Episode
	queryResult := s.db.Where("series_id = ?", seriesId).Order("episodes.name ASC").Preload("Thumb").Find(&episodes)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return episodes, nil
}

func (s *EpisodeService) DeleteEpisodeById(id uint) error {
	var episode db.Episode
	queryResult := s.db.First(&episode, "id = ?", id)
	if queryResult.Error != nil {
		return queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return util.ErrNotFound
	}
	queryResult = s.db.Delete(&db.Episode{}, id)
	if queryResult.Error != nil {
		return queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return util.ErrNotFound
	}
	go s.cleanUpEpisode(episode)
	return nil
}

func registerStaticEpisodes(engine *gin.Engine, config *config.Config) {
	engine.Static(util.EpisodeRoute, path.Join(config.Data.Dir, util.EpisodeSubDir))
}

var EpisodeServiceExport = fx.Options(fx.Provide(NewEpisodeService), fx.Invoke(registerStaticEpisodes))

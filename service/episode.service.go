package service

import (
	"anileha/config"
	"anileha/db"
	"anileha/util"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
	"path"
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
	episodeFolder := path.Join(workingDir, config.Data.Dir, util.ConversionSubDir)
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

func (s *ConversionService) GetEpisodeById(id uint) (*db.Episode, error) {
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

func (s *ConversionService) GetEpisodesBySeriesId(seriesId uint) ([]db.Episode, error) {
	var episodes []db.Episode
	queryResult := s.db.Find(&episodes, "seriesId = ?", seriesId)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, util.ErrNotFound
	}
	return episodes, nil
}

var EpisodeServiceExport = fx.Options(fx.Provide(NewEpisodeService))

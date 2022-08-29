package service

import (
	"anileha/db"
	"anileha/util"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SeriesService struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewSeriesService(db *gorm.DB, log *zap.Logger) *SeriesService {
	return &SeriesService{
		db, log,
	}
}

func (s *SeriesService) GetSeriesById(id uint) (*db.Series, error) {
	var series db.Series
	queryResult := s.db.Preload("Thumb").First(&series, "id = ?", id)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, util.ErrNotFound
	}
	return &series, nil
}

func (s *SeriesService) GetAllSeries() ([]db.Series, error) {
	var seriesArr []db.Series
	queryResult := s.db.Order(s.db.Order("series.created_at ASC")).Preload("Thumb").Find(&seriesArr)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return seriesArr, nil
}

func (s *SeriesService) SearchSeries(query string) ([]db.Series, error) {
	var seriesArr []db.Series
	queryResult := s.db.Where("name ILIKE '%' || ? || '%'", query).Order("series.created_at ASC").Preload("Thumb").Find(&seriesArr)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return seriesArr, nil
}

func (s *SeriesService) DeleteSeriesById(id uint) error {
	queryResult := s.db.Delete(&db.Series{}, id)
	if queryResult.Error != nil {
		return queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return util.ErrNotFound
	}
	return nil
}

func (s *SeriesService) AddSeries(name string, thumbId uint) (uint, error) {
	series := db.NewSeries(name, nil, &thumbId)
	queryResult := s.db.Create(&series)
	if queryResult.Error != nil {
		return 0, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return 0, util.ErrCreationFailed
	}
	s.log.Info("created series", zap.Uint("seriesId", series.ID), zap.String("seriesName", name))
	return series.ID, nil
}

var SeriesServiceExport = fx.Options(fx.Provide(NewSeriesService))

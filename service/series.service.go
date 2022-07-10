package service

import (
	"anileha/dao"
	"anileha/db"
	"errors"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type SeriesService struct {
	db *gorm.DB
}

func NewSeriesService(db *gorm.DB) *SeriesService {
	return &SeriesService{
		db,
	}
}

func (s *SeriesService) GetSeriesById(id uint) (*db.Series, error) {
	var series db.Series
	queryResult := s.db.Where("id = ?", id).First(&series)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, errors.New("not found")
	}
	return &series, nil
}

func (s *SeriesService) GetAllSeries() ([]db.Series, error) {
	var seriesArr []db.Series
	queryResult := s.db.Find(&seriesArr)
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
		return errors.New("not found")
	}
	return nil
}

func (s *SeriesService) AddSeries(req dao.SeriesDao) (uint, error) {
	series := db.NewSeries(req.Name, req.Description, req.Query, req.ThumbnailId)
	queryResult := s.db.Create(&series)
	if queryResult.Error != nil {
		return 0, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return 0, errors.New("creation failed")
	}
	return series.ID, nil
}

var SeriesServiceExport = fx.Options(fx.Provide(NewSeriesService))

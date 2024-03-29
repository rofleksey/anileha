package service

import (
	"anileha/db"
	"anileha/db/repo"
	"anileha/rest/engine"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type SeriesService struct {
	seriesRepo *repo.SeriesRepo
	log        *zap.Logger
}

func NewSeriesService(seriesRepo *repo.SeriesRepo, log *zap.Logger) *SeriesService {
	return &SeriesService{
		seriesRepo, log,
	}
}

func (s *SeriesService) GetById(id uint) (*db.Series, error) {
	series, err := s.seriesRepo.GetById(id)
	if err != nil {
		return nil, engine.ErrInternal(err.Error())
	}
	if series == nil {
		return nil, engine.ErrNotFoundInst
	}
	return series, nil
}

func (s *SeriesService) GetAll() ([]db.Series, error) {
	seriesArr, err := s.seriesRepo.GetAll()
	if err != nil {
		return nil, engine.ErrInternal(err.Error())
	}
	return seriesArr, nil
}

func (s *SeriesService) Search(query string) ([]db.Series, error) {
	seriesArr, err := s.seriesRepo.Search(query)
	if err != nil {
		return nil, engine.ErrInternal(err.Error())
	}
	return seriesArr, nil
}

func (s *SeriesService) DeleteById(id uint) error {
	series, err := s.seriesRepo.GetById(id)
	if err != nil {
		return engine.ErrInternal(err.Error())
	}
	if series == nil {
		return engine.ErrNotFoundInst
	}
	rows, err := s.seriesRepo.DeleteById(id)
	if err != nil {
		return engine.ErrInternal(err.Error())
	}
	if rows == 0 {
		return engine.ErrNotFoundInst
	}
	series.Thumb.Delete()
	return nil
}

func (s *SeriesService) SetQuery(id uint, query *db.SeriesQuery) error {
	if err := s.seriesRepo.SetQuery(id, query); err != nil {
		return engine.ErrInternal(err.Error())
	}
	return nil
}

func (s *SeriesService) AddSeries(name string, thumb db.Thumb) (uint, error) {
	series := db.Series{
		Title: name,
		Thumb: thumb,
	}
	id, err := s.seriesRepo.Create(&series)
	if err != nil {
		return 0, engine.ErrInternal(err.Error())
	}
	s.log.Info("created series", zap.Uint("seriesId", id), zap.String("seriesName", name))
	return id, nil
}

var SeriesExport = fx.Options(fx.Provide(NewSeriesService))

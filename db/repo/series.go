package repo

import (
	"anileha/db"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SeriesRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewSeriesRepo(db *gorm.DB, log *zap.Logger) *SeriesRepo {
	return &SeriesRepo{
		db:  db,
		log: log,
	}
}

func (r *SeriesRepo) GetById(id uint) (*db.Series, error) {
	var series db.Series
	queryResult := r.db.First(&series, "id = ?", id)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, nil
	}
	return &series, nil
}

func (r *SeriesRepo) DeleteById(id uint) (int64, error) {
	queryResult := r.db.Delete(&db.Series{}, id)
	if queryResult.Error != nil {
		return 0, queryResult.Error
	}
	return queryResult.RowsAffected, nil
}

func (r *SeriesRepo) GetAll() ([]db.Series, error) {
	var seriesArr []db.Series
	queryResult := r.db.
		Order("series.last_update DESC").
		Find(&seriesArr)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return seriesArr, nil
}

func (r *SeriesRepo) Search(query string) ([]db.Series, error) {
	var seriesArr []db.Series
	queryResult := r.db.Where("name ILIKE '%' || ? || '%'", query).
		Order("series.last_update DESC").
		Find(&seriesArr)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return seriesArr, nil
}

func (r *SeriesRepo) Create(series *db.Series) (uint, error) {
	queryResult := r.db.Create(series)
	if queryResult.Error != nil {
		return 0, queryResult.Error
	}
	return series.ID, nil
}

var SeriesRepoExport = fx.Options(fx.Provide(NewSeriesRepo))

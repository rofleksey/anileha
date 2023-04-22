package repo

import (
	"anileha/db"
	"anileha/util"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ConversionRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewConversionRepo(db *gorm.DB, log *zap.Logger) *ConversionRepo {
	return &ConversionRepo{
		db:  db,
		log: log,
	}
}

func (r *ConversionRepo) GetById(id uint) (*db.Conversion, error) {
	var conversion db.Conversion
	queryResult := r.db.First(&conversion, "id = ?", id)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, nil
	}
	return &conversion, nil
}

func (r *ConversionRepo) ResetProcessing() error {
	return r.db.Model(&db.Conversion{}).
		Where("status = ? OR status = ? OR status = ?", "", db.ConversionProcessing, db.ConversionCreated).
		Updates(db.Conversion{Status: db.ConversionError}).Error
}

func (r *ConversionRepo) SetStatus(id uint, status db.ConversionStatus) error {
	return r.db.Model(&db.Conversion{}).
		Where("id = ?", id).
		Updates(db.Conversion{Status: status}).Error
}

func (r *ConversionRepo) SetProgress(id uint, progress util.Progress) error {
	return r.db.Model(&db.Conversion{}).
		Where("id = ?", id).
		Updates(db.Conversion{Progress: progress}).Error
}

func (r *ConversionRepo) SetFinish(id uint, episodeId uint) error {
	return r.db.Model(&db.Conversion{}).
		Where("id = ?", id).
		Updates(db.Conversion{
			Status:    db.ConversionReady,
			EpisodeId: &episodeId,
			Progress:  util.Progress{Progress: 100},
		}).Error
}

func (r *ConversionRepo) DeleteById(id uint) (int64, error) {
	queryResult := r.db.Delete(&db.Conversion{}, id)
	if queryResult.Error != nil {
		return 0, queryResult.Error
	}
	return queryResult.RowsAffected, nil
}

func (r *ConversionRepo) GetAll() ([]db.Conversion, error) {
	var conversionArr []db.Conversion
	queryResult := r.db.
		Order("conversions.updated_at DESC").
		Find(&conversionArr)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return conversionArr, nil
}

func (r *ConversionRepo) GetBySeriesId(seriesId uint) ([]db.Conversion, error) {
	var conversions []db.Conversion
	queryResult := r.db.Where("series_id = ?", seriesId).
		Order("conversions.updated_at DESC").
		Find(&conversions)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return conversions, nil
}

func (r *ConversionRepo) GetByTorrentId(seriesId uint) ([]db.Conversion, error) {
	var conversions []db.Conversion
	queryResult := r.db.Where("torrent_id = ?", seriesId).
		Order("conversions.updated_at DESC").
		Find(&conversions)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return conversions, nil
}

func (r *ConversionRepo) Create(conversion *db.Conversion) (uint, error) {
	queryResult := r.db.Create(conversion)
	if queryResult.Error != nil {
		return 0, queryResult.Error
	}
	return conversion.ID, nil
}

var ConversionExport = fx.Options(fx.Provide(NewConversionRepo))

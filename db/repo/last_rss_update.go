package repo

import (
	"anileha/db"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type LastRSSRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewLastRSSRepo(db *gorm.DB, log *zap.Logger) *LastRSSRepo {
	return &LastRSSRepo{
		db:  db,
		log: log,
	}
}

func (r *LastRSSRepo) GetLast() (db.LastRSSUpdate, error) {
	entry := db.LastRSSUpdate{
		ID: 1,
	}
	if err := r.db.FirstOrCreate(&entry, db.LastRSSUpdate{
		ID: 1,
	}).Error; err != nil {
		return db.LastRSSUpdate{}, err
	}
	return entry, nil
}

func (r *LastRSSRepo) SetLast(newEntry db.LastRSSUpdate) error {
	return r.db.Model(&db.LastRSSUpdate{}).
		Where("id = ?", 1).
		Updates(db.LastRSSUpdate{
			Timestamp: newEntry.Timestamp,
			RssId:     newEntry.RssId,
		}).Error
}

var LastRSSExport = fx.Options(fx.Provide(NewLastRSSRepo))

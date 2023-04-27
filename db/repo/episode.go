package repo

import (
	"anileha/db"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type EpisodeRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewEpisodeRepo(db *gorm.DB, log *zap.Logger) *EpisodeRepo {
	return &EpisodeRepo{
		db:  db,
		log: log,
	}
}

func (r *EpisodeRepo) GetById(id uint) (*db.Episode, error) {
	var episode db.Episode
	queryResult := r.db.First(&episode, "id = ?", id)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, nil
	}
	return &episode, nil
}

func (r *EpisodeRepo) Get(offset int, limit int) ([]db.Episode, error) {
	var episodes []db.Episode
	queryResult := r.db.
		Offset(offset).
		Limit(limit).
		Order("episodes.created_at DESC").
		Find(&episodes)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return episodes, nil
}

func (r *EpisodeRepo) Count() (int64, error) {
	var count int64

	if err := r.db.Table("episodes").Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *EpisodeRepo) GetBySeriesId(seriesId uint) ([]db.Episode, error) {
	var episodes []db.Episode
	queryResult := r.db.Where("series_id = ?", seriesId).
		Order("episodes.title ASC").
		Find(&episodes)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return episodes, nil
}

func (r *EpisodeRepo) SetThumb(id uint, thumb db.Thumb) error {
	return r.db.Model(&db.Episode{}).
		Where("id = ?", id).
		Updates(db.Episode{Thumb: thumb}).Error
}

func (r *EpisodeRepo) DeleteById(id uint) (int64, error) {
	queryResult := r.db.Delete(&db.Episode{}, id)
	if queryResult.Error != nil {
		return 0, queryResult.Error
	}
	return queryResult.RowsAffected, nil
}

func (r *EpisodeRepo) Create(episode *db.Episode) (uint, error) {
	queryResult := r.db.Create(episode)
	if queryResult.Error != nil {
		return 0, queryResult.Error
	}
	return episode.ID, nil
}

var EpisodeExport = fx.Options(fx.Provide(NewEpisodeRepo))

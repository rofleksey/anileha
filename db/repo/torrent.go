package repo

import (
	"anileha/db"
	"anileha/rest"
	"anileha/util"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TorrentRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewTorrentRepo(db *gorm.DB, log *zap.Logger) *TorrentRepo {
	return &TorrentRepo{
		db:  db,
		log: log,
	}
}

func (r *TorrentRepo) ResetDownloadStatus() error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&db.Torrent{}).
			Where("status = ? or status = ?", db.TorrentDownloading, db.TorrentCreating).
			Updates(db.Torrent{Status: db.TorrentError}).Error; err != nil {
			return err
		}
		if err := tx.Model(&db.TorrentFile{}).
			Where("status = ?", db.TorrentFileDownloading).
			Updates(map[string]interface{}{"status": db.TorrentFileError, "selected": false}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *TorrentRepo) DeleteById(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&db.TorrentFile{}, "torrent_id = ?", id).Error; err != nil {
			return err
		}
		queryResult := tx.Delete(&db.Torrent{}, id)
		if queryResult.Error != nil {
			return queryResult.Error
		}
		if queryResult.RowsAffected == 0 {
			return rest.ErrNotFoundInst
		}
		return nil
	})
}

func (r *TorrentRepo) SetReady(torrent db.Torrent) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// update torrent status
		if err := tx.Model(&db.Torrent{}).Where("id = ?", torrent.ID).Updates(db.Torrent{Status: db.TorrentReady}).Error; err != nil {
			return err
		}

		// update downloaded files' statuses
		for _, file := range torrent.Files {
			if file.Selected && file.Status != db.TorrentFileReady {
				if err := tx.Model(&db.TorrentFile{}).
					Where("id = ?", file.ID).
					Updates(db.TorrentFile{Status: db.TorrentFileReady, ReadyPath: file.ReadyPath}).Error; err != nil {
					return err
				}
			}
		}

		// apply transaction
		return nil
	})
}

func (r *TorrentRepo) UpdateProgress(id uint, progress util.Progress, bytesRead uint) error {
	if err := r.db.Model(&db.Torrent{}).
		Where("id = ?", id).
		Updates(db.Torrent{Progress: progress, BytesRead: bytesRead}).Error; err != nil {
		return err
	}

	return nil
}

func (r *TorrentRepo) GetById(id uint) (*db.Torrent, error) {
	var torrent db.Torrent
	queryResult := r.db.Preload("Files", func(db *gorm.DB) *gorm.DB {
		return db.Order("torrent_files.client_index ASC")
	}).First(&torrent, "id = ?", id)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, nil
	}
	return &torrent, nil
}

func (r *TorrentRepo) GetBySeriesId(seriesId uint) ([]db.Torrent, error) {
	var torrentArr []db.Torrent
	queryResult := r.db.
		Where("series_id = ? and status != ?", seriesId, db.TorrentCreating).
		Order("torrents.created_at DESC").
		Find(&torrentArr)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return torrentArr, nil
}

func (r *TorrentRepo) GetAll() ([]db.Torrent, error) {
	var torrentArr []db.Torrent
	queryResult := r.db.Where("status != ?", db.TorrentCreating).
		Order("torrents.created_at DESC").
		Find(&torrentArr)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return torrentArr, nil
}

func (r *TorrentRepo) InitFiles(torrent db.Torrent, files []db.TorrentFile) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		queryResult := tx.Create(files)
		if queryResult.Error != nil {
			return queryResult.Error
		}
		if queryResult.RowsAffected == 0 {
			return util.ErrFileMapping
		}

		queryResult = tx.Save(&torrent)
		if queryResult.Error != nil {
			return queryResult.Error
		}

		return nil
	})
}

func (r *TorrentRepo) StartTorrent(id uint, unselectedIds []uint, selectedIds []uint, downloadLength uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&db.Torrent{}).
			Where("id = ?", id).
			Updates(db.Torrent{Status: db.TorrentDownloading, TotalDownloadLength: downloadLength}).Error
		if err != nil {
			return err
		}
		for _, id := range unselectedIds {
			err = tx.Model(&db.TorrentFile{}).
				Where("id = ?", id).
				Updates(map[string]interface{}{"status": db.TorrentFileIdle, "selected": false}).Error
			if err != nil {
				return err
			}
		}
		for _, id := range selectedIds {
			err = tx.Model(&db.TorrentFile{}).
				Where("id = ?", id).
				Updates(map[string]interface{}{"status": db.TorrentFileDownloading, "selected": true}).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *TorrentRepo) StopTorrent(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&db.Torrent{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{"status": db.TorrentIdle, "total_download_length": 0}).Error
		if err != nil {
			return rest.ErrInternal(err.Error())
		}
		err = tx.Model(&db.TorrentFile{}).
			Where("torrent_id = ?", id).
			Updates(map[string]interface{}{"status": db.TorrentFileIdle, "selected": false}).Error
		if err != nil {
			return rest.ErrInternal(err.Error())
		}
		return nil
	})
}

func (r *TorrentRepo) Create(torrent *db.Torrent) (uint, error) {
	queryResult := r.db.Create(torrent)
	if queryResult.Error != nil {
		return 0, queryResult.Error
	}
	return torrent.ID, nil
}

var TorrentRepoExport = fx.Options(fx.Provide(NewTorrentRepo))

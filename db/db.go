package db

import (
	"anileha/config"
	"fmt"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TODO: ordering, indices, cascading

func initDB(config *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		config.Db.Host, config.Db.Port, config.Db.DbName, config.Db.Username, config.Db.Password)), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&Series{}, &Thumb{}, &Torrent{}, &TorrentFile{}, &Conversion{}, &Episode{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

var ServiceExport = fx.Options(fx.Provide(initDB))

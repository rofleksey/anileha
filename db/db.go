package db

import (
	"anileha/config"
	"fmt"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

// TODO: ordering, indices

func initDB(config *config.Config) (*gorm.DB, error) {
	dbLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{})

	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		config.Db.Host, config.Db.Port, config.Db.DbName, config.Db.Username, config.Db.Password)), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&Series{}, &Torrent{}, &TorrentFile{}, &User{}, &Conversion{}, &Episode{}, &LastRSSUpdate{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

var ServiceExport = fx.Options(fx.Provide(initDB))

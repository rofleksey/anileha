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
	"time"
)

// TODO: ordering, indices

func initDB(config *config.Config) (*gorm.DB, error) {
	dbLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{})

	psqlDsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		config.Db.Host, config.Db.Port, config.Db.DbName, config.Db.Username, config.Db.Password)
	psql := postgres.New(postgres.Config{
		DSN: psqlDsn,
	})

	db, err := gorm.Open(psql, &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Series{}, &Torrent{}, &TorrentFile{}, &User{}, &Conversion{}, &Episode{}, &LastRSSUpdate{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(config.Db.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.Db.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.Db.ConnMaxLifetimeSecs) * time.Second)

	return db, nil
}

var ServiceExport = fx.Options(fx.Provide(initDB))

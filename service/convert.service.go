package service

import (
	"anileha/config"
	"anileha/util"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
	"path"
)

type ConversionService struct {
	db               *gorm.DB
	log              *zap.Logger
	conversionFolder string
}

func NewConversionService(lifecycle fx.Lifecycle, db *gorm.DB, log *zap.Logger, config *config.Config) (*ConversionService, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	conversionFolder := path.Join(workingDir, config.Data.Dir, util.ConversionSubDir)
	err = os.MkdirAll(conversionFolder, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return &ConversionService{
		db:               db,
		log:              log,
		conversionFolder: conversionFolder,
	}, nil
}

func (s *ConversionService) StartConversion() {

}

var ConversionServiceExport = fx.Options(fx.Provide(NewConversionService))

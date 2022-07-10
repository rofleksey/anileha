package service

import (
	"anileha/config"
	"fmt"
	"github.com/gofrs/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
	"path"
	"path/filepath"
)

type FileService struct {
	db         *gorm.DB
	log        *zap.SugaredLogger
	shutdowner fx.Shutdowner
	tempDir    string
}

func NewFileService(db *gorm.DB, log *zap.SugaredLogger, config *config.Config, shutdowner fx.Shutdowner) (*FileService, error) {
	thumbnailDir := path.Join(config.Data.Dir, "temp")
	err := os.MkdirAll(thumbnailDir, os.ModePerm)
	if err != nil {
		log.Error(err)
		err = shutdowner.Shutdown()
		if err != nil {
			log.Fatal(err)
		}
	}
	return &FileService{
		db, log, shutdowner, thumbnailDir,
	}, nil
}

func (s *FileService) GetTempFileDst(originalName string) (string, error) {
	fakeId, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	ext := filepath.Ext(originalName)
	dstPath := path.Join(s.tempDir, fmt.Sprintf("%s%s", fakeId, ext))
	return dstPath, nil
}

func (s *FileService) GetFileDst(folder string, originalName string) (string, error) {
	fakeId, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	ext := filepath.Ext(originalName)
	dstPath := path.Join(folder, fmt.Sprintf("%s%s", fakeId, ext))
	return dstPath, nil
}

func (s *FileService) DeleteFileAsync(tempDst string) {
	go func() {
		err := os.Remove(tempDst)
		if err != nil {
			s.log.Error("error deleting temp file", err)
		}
	}()
}

var FileServiceExport = fx.Options(fx.Provide(NewFileService))

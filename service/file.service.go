package service

import (
	"anileha/config"
	"anileha/util"
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
	db      *gorm.DB
	log     *zap.Logger
	tempDir string
}

func NewFileService(db *gorm.DB, log *zap.Logger, config *config.Config) (*FileService, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	thumbnailDir := path.Join(workingDir, config.Data.Dir, util.TempSubDir)
	err = os.MkdirAll(thumbnailDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return &FileService{
		db, log, thumbnailDir,
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

func (s *FileService) DeleteTempFileAsync(tempDst string) {
	go func() {
		_ = os.Remove(tempDst)
	}()
}

var FileServiceExport = fx.Options(fx.Provide(NewFileService))

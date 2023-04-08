package service

import (
	"anileha/config"
	"anileha/db"
	"anileha/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"os"
	"path"
	"path/filepath"
)

type ThumbService struct {
	log         *zap.Logger
	fileService *FileService
	thumbDir    string
}

func NewThumbService(config *config.Config, log *zap.Logger, fileService *FileService) (*ThumbService, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	thumbDir := path.Join(workingDir, config.Data.Dir, util.ThumbSubDir)
	err = os.MkdirAll(thumbDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return &ThumbService{
		log, fileService, thumbDir,
	}, nil
}

func (s *ThumbService) CreateFromTempFile(tempPath string) (db.Thumb, error) {
	newPath, err := s.fileService.GenFilePath(s.thumbDir, tempPath)
	if err != nil {
		return db.Thumb{}, err
	}
	err = os.Rename(tempPath, newPath)
	if err != nil {
		return db.Thumb{}, err
	}
	url := fmt.Sprintf("%s/%s", util.ThumbRoute, filepath.Base(newPath))
	thumb := db.Thumb{
		Path: newPath,
		Url:  url,
	}
	return thumb, nil
}

func registerStaticThumbs(engine *gin.Engine, config *config.Config) {
	engine.Static(util.ThumbRoute, path.Join(config.Data.Dir, util.ThumbSubDir))
}

var ThumbServiceExport = fx.Options(fx.Provide(NewThumbService), fx.Invoke(registerStaticThumbs))

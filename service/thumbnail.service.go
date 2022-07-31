package service

import (
	"anileha/config"
	"anileha/db"
	"anileha/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
	"path"
	"path/filepath"
)

type ThumbnailService struct {
	db           *gorm.DB
	log          *zap.Logger
	fileService  *FileService
	thumbnailDir string
}

func NewThumbnailService(db *gorm.DB, config *config.Config, log *zap.Logger, fileService *FileService) (*ThumbnailService, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	thumbnailDir := path.Join(workingDir, config.Data.Dir, util.ThumbSubDir)
	err = os.MkdirAll(thumbnailDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return &ThumbnailService{
		db, log, fileService, thumbnailDir,
	}, nil
}

func (s *ThumbnailService) GetThumbnailById(id uint) (*db.Thumbnail, error) {
	var thumb db.Thumbnail
	queryResult := s.db.First(&thumb, "id = ?", id)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, util.ErrNotFound
	}
	return &thumb, nil
}

func (s *ThumbnailService) GetAllThumbnails() ([]db.Thumbnail, error) {
	var thumbArr []db.Thumbnail
	queryResult := s.db.Find(&thumbArr)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return thumbArr, nil
}

func (s *ThumbnailService) DeleteThumbnailById(id uint) error {
	queryResult := s.db.Delete(&db.Thumbnail{}, id)
	if queryResult.Error != nil {
		return queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return util.ErrNotFound
	}
	s.log.Info("deleted thumbnail", zap.Uint("thumbId", id))
	return nil
}

func (s *ThumbnailService) AddThumbnail(name string, tempPath string) (uint, error) {
	newPath, err := s.fileService.GenFilePath(s.thumbnailDir, tempPath)
	if err != nil {
		return 0, err
	}
	err = os.Rename(tempPath, newPath)
	if err != nil {
		return 0, err
	}
	downloadUrl := fmt.Sprintf("%s/%s", util.ThumbRoute, filepath.Base(newPath))
	thumb := db.NewThumbnail(name, tempPath, downloadUrl)
	queryResult := s.db.Create(&thumb)
	if queryResult.Error != nil {
		deleteErr := os.Remove(newPath)
		if deleteErr != nil {
			s.log.Error("error deleting thumbnail on error", zap.Error(deleteErr))
		}
		return 0, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		deleteErr := os.Remove(newPath)
		if deleteErr != nil {
			s.log.Error("error deleting thumbnail on error", zap.Error(deleteErr))
		}
		return 0, util.ErrCreationFailed
	}
	s.log.Info("created thumbnail", zap.Uint("thumbId", thumb.ID), zap.String("thumbName", name))
	return thumb.ID, nil
}

func registerStaticThumbnails(engine *gin.Engine, config *config.Config) {
	engine.Static(util.ThumbRoute, path.Join(config.Data.Dir, util.ThumbSubDir))
}

var ThumbnailServiceExport = fx.Options(fx.Provide(NewThumbnailService), fx.Invoke(registerStaticThumbnails))

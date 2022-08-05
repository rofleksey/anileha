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

type ThumbService struct {
	db          *gorm.DB
	log         *zap.Logger
	fileService *FileService
	thumbDir    string
}

func NewThumbService(db *gorm.DB, config *config.Config, log *zap.Logger, fileService *FileService) (*ThumbService, error) {
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
		db, log, fileService, thumbDir,
	}, nil
}

func (s *ThumbService) cleanUpThumb(thumb db.Thumb) {
	if err := os.Remove(thumb.Path); err != nil {
		s.log.Error("failed to remove thumb file", zap.Uint("thumbId", thumb.ID), zap.String("file", thumb.Path), zap.Error(err))
	}
}

func (s *ThumbService) GetThumbById(id uint) (*db.Thumb, error) {
	var thumb db.Thumb
	queryResult := s.db.First(&thumb, "id = ?", id)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, util.ErrNotFound
	}
	return &thumb, nil
}

func (s *ThumbService) GetAllThumbs() ([]db.Thumb, error) {
	var thumbArr []db.Thumb
	queryResult := s.db.Find(&thumbArr, func(db *gorm.DB) *gorm.DB {
		return db.Order("thumbs.created_at ASC")
	})
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	return thumbArr, nil
}

func (s *ThumbService) DeleteThumbById(id uint) error {
	queryResult := s.db.Delete(&db.Thumb{}, id)
	if queryResult.Error != nil {
		return queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return util.ErrNotFound
	}
	s.log.Info("deleted thumb", zap.Uint("thumbId", id))
	return nil
}

func (s *ThumbService) AddThumb(name string, tempPath string) (uint, error) {
	newPath, err := s.fileService.GenFilePath(s.thumbDir, tempPath)
	if err != nil {
		return 0, err
	}
	err = os.Rename(tempPath, newPath)
	if err != nil {
		return 0, err
	}
	downloadUrl := fmt.Sprintf("%s/%s", util.ThumbRoute, filepath.Base(newPath))
	thumb := db.NewThumb(name, tempPath, downloadUrl)
	queryResult := s.db.Create(&thumb)
	if queryResult.Error != nil {
		deleteErr := os.Remove(newPath)
		if deleteErr != nil {
			s.log.Error("error deleting thumb on error", zap.Error(deleteErr))
		}
		return 0, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		deleteErr := os.Remove(newPath)
		if deleteErr != nil {
			s.log.Error("error deleting thumb on error", zap.Error(deleteErr))
		}
		return 0, util.ErrCreationFailed
	}
	s.log.Info("created thumb", zap.Uint("thumbId", thumb.ID), zap.String("thumbName", name))
	return thumb.ID, nil
}

func registerStaticThumbs(engine *gin.Engine, config *config.Config) {
	engine.Static(util.ThumbRoute, path.Join(config.Data.Dir, util.ThumbSubDir))
}

var ThumbServiceExport = fx.Options(fx.Provide(NewThumbService), fx.Invoke(registerStaticThumbs))

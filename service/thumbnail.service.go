package service

import (
	"anileha/config"
	"anileha/db"
	"anileha/util"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"os"
	"path"
	"path/filepath"
)

type ThumbnailService struct {
	db           *gorm.DB
	fileService  *FileService
	thumbnailDir string
}

func NewThumbnailService(db *gorm.DB, config *config.Config, fileService *FileService) (*ThumbnailService, error) {
	thumbnailDir := path.Join(config.Data.Dir, util.ThumbSubDir)
	err := os.MkdirAll(thumbnailDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return &ThumbnailService{
		db, fileService, thumbnailDir,
	}, nil
}

func (s *ThumbnailService) GetThumbnailById(id uint) (*db.Thumbnail, error) {
	var thumb db.Thumbnail
	queryResult := s.db.Where("id = ?", id).First(&thumb)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, errors.New("not found")
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

func (s *ThumbnailService) DeleteThumbnailsById(id uint) error {
	queryResult := s.db.Delete(&db.Thumbnail{}, id)
	if queryResult.Error != nil {
		return queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return errors.New("not found")
	}
	return nil
}

func (s *ThumbnailService) AddThumbnail(name string, tempPath string) (uint, error) {
	newPath, err := s.fileService.GetFileDst(util.ThumbSubDir, tempPath)
	if err != nil {
		return 0, err
	}
	err = os.Rename(tempPath, newPath)
	if err != nil {
		return 0, err
	}
	downloadUrl := fmt.Sprintf("/%s/%s", util.ThumbRoute, filepath.Base(newPath))
	thumb := db.NewThumbnail(name, tempPath, downloadUrl)
	queryResult := s.db.Create(&thumb)
	if queryResult.Error != nil {
		return 0, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return 0, errors.New("creation failed")
	}
	return thumb.ID, nil
}

func registerStaticThumbnails(engine *gin.Engine, config *config.Config) {
	engine.Static(util.ThumbRoute, path.Join(config.Data.Dir, util.ThumbSubDir))
}

var ThumbnailServiceExport = fx.Options(fx.Provide(NewThumbnailService), fx.Invoke(registerStaticThumbnails))

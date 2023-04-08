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

//func (s *ThumbService) getVideoFileThumbnail(inputFile string, timeSeconds int) ([]byte, error) {
//	s.log.Info("generating thumbnail", zap.String("inputFile", inputFile))
//	sizeCommand := ffmpeg.NewCommand(inputFile, 0, srtFileName)
//	sizeCommand.AddKeyValue("-ss", strconv.Itoa(timeSeconds), ffmpeg.OptionBase)
//	sizeCommand.AddKeyValue("-frames:v", "1", ffmpeg.OptionInput)
//	sizeCommand.AddKeyValue("-frames:v", "1", ffmpeg.OptionInput)
//	output, err := sizeCommand.ExecuteSync()
//	if err != nil {
//		p.log.Warn(fmt.Sprintf("failed to get sub text: %s", *output), zap.String("inputFile", inputFile), zap.Int("streamIndex", streamIndex), zap.Error(err))
//		return "", err
//	}
//	content, err := os.ReadFile(srtFileName)
//	if err != nil {
//		return "", err
//	}
//	return string(content), nil
//}

//func (s *ThumbService) CreateForVideo(videoFile string) (db.Thumb, error) {
//	newPath, err := s.fileService.GenFilePath(s.thumbDir, "thumb.jpg")
//	if err != nil {
//		return db.Thumb{}, err
//	}
//	err = os.Rename(tempPath, newPath)
//	if err != nil {
//		return db.Thumb{}, err
//	}
//	url := fmt.Sprintf("%s/%s", util.ThumbRoute, filepath.Base(newPath))
//	thumb := db.Thumb{
//		Path: newPath,
//		Url:  url,
//	}
//	return thumb, nil
//}

func registerStaticThumbs(engine *gin.Engine, config *config.Config) {
	engine.Static(util.ThumbRoute, path.Join(config.Data.Dir, util.ThumbSubDir))
}

var ThumbServiceExport = fx.Options(fx.Provide(NewThumbService), fx.Invoke(registerStaticThumbs))

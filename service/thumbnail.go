package service

import (
	"anileha/config"
	"anileha/db"
	"anileha/ffmpeg"
	"anileha/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

type ThumbService struct {
	log         *zap.Logger
	config      *config.Config
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
		log:         log,
		config:      config,
		fileService: fileService,
		thumbDir:    thumbDir,
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

type thumbnailResult struct {
	result []byte
	err    error
}

func (s *ThumbService) videoFileThumbnailWorker(inputFile string, timeSeconds int, outputChan chan thumbnailResult) {
	s.log.Info("generating thumbnail", zap.String("inputFile", inputFile))
	sizeCommand := ffmpeg.NewCommand(inputFile, 0, "-")
	sizeCommand.AddKeyValue("-ss", strconv.Itoa(timeSeconds), ffmpeg.OptionBase)
	sizeCommand.AddKeyValue("-c:v", "mjpeg", ffmpeg.OptionOutput)
	sizeCommand.AddKeyValue("-frames:v", "1", ffmpeg.OptionOutput)
	sizeCommand.AddKeyValue("-f", "image2", ffmpeg.OptionOutput)
	output, err := sizeCommand.ExecuteSync()
	if err != nil {
		outputChan <- thumbnailResult{
			err: fmt.Errorf("ffmpeg error: %w", err),
		}
		return
	}
	outputChan <- thumbnailResult{
		result: output,
	}
}

func (s *ThumbService) CreateForVideo(videoFile string, durationSec int) (db.Thumb, error) {
	var thumbBytes []byte

	var lastError error

	factors := s.config.Thumb.VideoFactors
	resultsChan := make(chan thumbnailResult, len(factors))

	for _, factor := range factors {
		go s.videoFileThumbnailWorker(videoFile, int(float32(durationSec)*factor), resultsChan)
	}

	for i := 0; i < len(factors); i++ {
		result := <-resultsChan

		if result.err != nil {
			s.log.Warn("failed to create thumbnail", zap.Error(result.err))
			lastError = result.err
			continue
		}

		if thumbBytes == nil || len(thumbBytes) < len(result.result) {
			thumbBytes = result.result
		}
	}

	if thumbBytes == nil {
		return db.Thumb{}, lastError
	}

	newPath, err := s.fileService.GenFilePath(s.thumbDir, "thumb.jpg")
	if err != nil {
		return db.Thumb{}, err
	}

	err = os.WriteFile(newPath, thumbBytes, 0644)
	if err != nil {
		return db.Thumb{}, err
	}
	url := fmt.Sprintf("%s/%s", util.ThumbRoute, filepath.Base(newPath))

	return db.Thumb{
		Path: newPath,
		Url:  url,
	}, nil
}

func registerStaticThumbs(engine *gin.Engine, config *config.Config) {
	engine.Static(util.ThumbRoute, path.Join(config.Data.Dir, util.ThumbSubDir))
}

var ThumbServiceExport = fx.Options(fx.Provide(NewThumbService), fx.Invoke(registerStaticThumbs))

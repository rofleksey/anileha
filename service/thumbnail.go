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
	"golang.org/x/exp/rand"
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
	tempPath, err := s.fileService.GenFilePath(s.thumbDir, "thumb.jpg")
	if err != nil {
		outputChan <- thumbnailResult{
			err: fmt.Errorf("can't create temp file: %w", err),
		}
		return
	}

	defer func() {
		_ = os.Remove(tempPath)
	}()

	thumbCmd := ffmpeg.NewCommand(inputFile, 0, tempPath)
	thumbCmd.AddKeyValue("-ss", strconv.Itoa(timeSeconds), ffmpeg.OptionBase)
	thumbCmd.AddKeyValue("-frames:v", "1", ffmpeg.OptionOutput)

	s.log.Info("generating thumbnail",
		zap.String("inputFile", inputFile),
		zap.String("command", thumbCmd.String()))

	_, err = thumbCmd.ExecuteSync()
	if err != nil {
		outputChan <- thumbnailResult{
			err: fmt.Errorf("ffmpeg error: %w", err),
		}
		return
	}

	imageBytes, err := os.ReadFile(tempPath)
	if err != nil {
		outputChan <- thumbnailResult{
			err: fmt.Errorf("failed to read result: %w", err),
		}
		return
	}

	outputChan <- thumbnailResult{
		result: imageBytes,
	}
}

func (s *ThumbService) CreateForVideo(videoFile string, durationSec int) (db.Thumb, error) {
	var thumbBytes []byte

	var lastError error

	factors := make([]float32, 0, s.config.Thumb.Attempts)
	for i := 0; i < s.config.Thumb.Attempts; i++ {
		factors = append(factors, rand.Float32())
	}

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

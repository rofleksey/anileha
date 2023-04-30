package service

import (
	"anileha/config"
	"anileha/db"
	"anileha/util"
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/mholt/archiver/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"io"
	"os"
	"path"
)

type FontService struct {
	log         *zap.Logger
	config      *config.Config
	fileService *FileService
	fontDir     string
}

func NewFontService(config *config.Config, log *zap.Logger, fileService *FileService) (*FontService, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	fontDir := path.Join(workingDir, config.Data.Dir, util.FontSubDir)
	err = os.MkdirAll(fontDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return &FontService{
		log:         log,
		config:      config,
		fileService: fileService,
		fontDir:     fontDir,
	}, nil
}

func (s *FontService) LoadFonts(ctx context.Context, files []db.TorrentFile) {
	newFontCounter := 0

	readyFontFiles := pie.Filter(files, func(file db.TorrentFile) bool {
		return file.Type == util.FileTypeFont && file.ReadyPath != nil
	})

	for _, file := range readyFontFiles {
		newPath := path.Join(s.fontDir, path.Base(file.TorrentPath))
		err := os.Rename(*file.ReadyPath, newPath)
		if err != nil {
			s.log.Warn("failed to load font",
				zap.String("path", file.TorrentPath),
				zap.Error(err))
		} else {
			newFontCounter++
		}
	}

	readyArchiveFiles := pie.Filter(files, func(file db.TorrentFile) bool {
		return file.Type == util.FileTypeArchive && file.ReadyPath != nil
	})

	for _, file := range readyArchiveFiles {
		count, err := s.handleArchive(ctx, *file.ReadyPath)
		if err != nil {
			s.log.Warn("failed to extract fonts from archive",
				zap.String("path", *file.ReadyPath),
				zap.Error(err))
		} else {
			newFontCounter += count
		}
	}

	s.log.Info("fonts loaded", zap.Int("count", newFontCounter))
}

func (s *FontService) handleArchive(ctx context.Context, filePath string) (int, error) {
	newFontCounter := 0

	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}

	defer file.Close()

	format, _, err := archiver.Identify(filePath, file)
	if err != nil {
		return 0, fmt.Errorf("failed to identify file format: %w", err)
	}

	extractor, extractorOk := format.(archiver.Extractor)
	if !extractorOk {
		return 0, fmt.Errorf("failed to obtain extractor: %w", err)
	}

	handler := func(ctx context.Context, f archiver.File) error {
		fileType := util.GetFileType(f.Name())
		if fileType != util.FileTypeFont {
			return nil
		}

		reader, err := f.Open()
		if err != nil {
			s.log.Warn("failed to open file for reading",
				zap.String("file", f.Name()),
				zap.Error(err))
			return nil
		}

		defer reader.Close()

		destPath := path.Join(s.fontDir, path.Base(f.Name()))

		destFile, err := os.Create(destPath)
		if err != nil {
			s.log.Warn("failed to open file for writing",
				zap.String("file", f.Name()),
				zap.Error(err))
			return nil
		}

		defer destFile.Close()

		_, err = io.Copy(destFile, reader)
		if err != nil {
			_ = os.Remove(destPath)

			s.log.Warn("failed to extract file",
				zap.String("file", f.Name()),
				zap.Error(err))
			return nil
		}

		newFontCounter++

		return nil
	}

	err = extractor.Extract(ctx, file, nil, handler)
	if err != nil {
		return 0, fmt.Errorf("failed to extract archive: %w", err)
	}

	return newFontCounter, nil
}

var FontExport = fx.Options(fx.Provide(NewFontService))

package service

import (
	"anileha/config"
	"anileha/db"
	"anileha/db/repo"
	"anileha/search"
	"anileha/search/nyaa"
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SearchService struct {
	seriesRepo     *repo.SeriesRepo
	lastRssRepo    *repo.LastRSSRepo
	fileService    *FileService
	torrentService *TorrentService
	nyaaService    *nyaa.Service
	log            *zap.Logger
	config         *config.Config

	pollCtx         context.Context
	pollCancel      context.CancelFunc
	pollWg          sync.WaitGroup
	pollTriggerChan chan struct{}
}

func NewSearchService(
	seriesRepo *repo.SeriesRepo,
	lastRssRepo *repo.LastRSSRepo,
	fileService *FileService,
	torrentService *TorrentService,
	nyaaService *nyaa.Service,
	log *zap.Logger,
	config *config.Config,
) *SearchService {
	pollCtx, pollCancel := context.WithCancel(context.Background())

	searchService := &SearchService{
		seriesRepo:     seriesRepo,
		lastRssRepo:    lastRssRepo,
		fileService:    fileService,
		torrentService: torrentService,
		nyaaService:    nyaaService,
		log:            log,
		config:         config,

		pollCtx:         pollCtx,
		pollCancel:      pollCancel,
		pollTriggerChan: make(chan struct{}, 1),
	}

	return searchService
}

func (s *SearchService) test(ctx context.Context, result *search.ResultRSS, query *db.SeriesQuery) bool {
	title := strings.ToLower(result.Title)

	if !pie.All(query.Include, func(value string) bool {
		return strings.Contains(title, value)
	}) {
		return false
	}

	if pie.Any(query.Exclude, func(value string) bool {
		return strings.Contains(title, value)
	}) {
		return false
	}

	if query.SingleFile {
		extra, err := s.nyaaService.GetById(ctx, result.ID)
		if err != nil {
			s.log.Error("failed to get extra by id",
				zap.String("id", result.ID),
				zap.String("title", result.Title),
				zap.Error(err))
			return false
		}

		if len(extra.Files) != 1 {
			s.log.Info("doesnt have single file, skipping",
				zap.String("id", result.ID),
				zap.String("title", result.Title),
				zap.Int("files", len(extra.Files)))
			return false
		}
	}

	return true
}

func (s *SearchService) onMatch(ctx context.Context, seriesId uint, auto db.AutoTorrent, rssId string, rssTitle string) bool {
	s.log.Info("found rss match",
		zap.String("id", rssId),
		zap.String("title", rssTitle))

	torrentBytes, err := s.nyaaService.DownloadById(ctx, rssId)
	if err != nil {
		s.log.Error("failed to download torrent by id",
			zap.String("id", rssId),
			zap.String("title", rssTitle),
			zap.Error(err))
		return false
	}

	tempDst, err := s.fileService.GenTempFilePath("new.torrent")
	if err != nil {
		s.log.Error("failed to create temp file",
			zap.String("id", rssId),
			zap.String("title", rssTitle),
			zap.Error(err))
		return false
	}
	defer s.fileService.DeleteTempFileAsync(tempDst)

	err = os.WriteFile(tempDst, torrentBytes, 0644)
	if err != nil {
		s.log.Error("failed to save torrent file",
			zap.String("id", rssId),
			zap.String("title", rssTitle),
			zap.Error(err))
		return false
	}

	err = s.torrentService.AddFromFile(seriesId, tempDst, &auto)
	if err != nil {
		s.log.Error("failed to add new torrent",
			zap.String("id", rssId),
			zap.String("title", rssTitle),
			zap.Error(err))
		return false
	}

	s.log.Info("successfully added new torrent", zap.String("title", rssTitle))
	return true
}

func (s *SearchService) TriggerRSSPoll() {
	select {
	case s.pollTriggerChan <- struct{}{}:
		s.log.Info("scheduled manual rss poll")
	default:
	}
}

func (s *SearchService) SearchOld(ctx context.Context, query db.SeriesQuery) ([]search.Result, error) {
	results := make([]search.Result, 0, 100)

	for page := 0; ; page++ {
		curResults, err := s.nyaaService.Search(ctx, search.Query{
			Query:    strings.Join(query.Include, " "),
			SortType: search.SortDate,
			Page:     page,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to search torrents: %w", err)
		}

		results = append(results, curResults...)

		if len(curResults) == 0 {
			break
		}
	}

	results = pie.Filter(results, func(result search.Result) bool {
		title := strings.ToLower(result.Title)

		return !pie.Any(query.Exclude, func(value string) bool {
			return strings.Contains(title, value)
		})
	})

	if query.SingleFile {
		results = pie.Filter(results, func(result search.Result) bool {
			extra, err := s.nyaaService.GetById(ctx, result.ID)
			if err != nil {
				s.log.Error("failed to get extra by id",
					zap.String("id", result.ID),
					zap.String("title", result.Title),
					zap.Error(err))
				return false
			}

			if len(extra.Files) != 1 {
				s.log.Info("doesnt have single file, skipping",
					zap.String("id", result.ID),
					zap.String("title", result.Title),
					zap.Int("files", len(extra.Files)))
				return false
			}

			return true
		})
	}

	return results, nil
}

func (s *SearchService) doPoll(ctx context.Context) error {
	lastRss, err := s.lastRssRepo.GetLast()
	if err != nil {
		return fmt.Errorf("failed to get last rss poll timestamp: %w", err)
	}
	lastRssId, _ := strconv.Atoi(lastRss.RssId)

	seriesWithQueries, err := s.seriesRepo.GetAllWithQuery()
	if err != nil {
		return fmt.Errorf("failed to get series with queries: %w", err)
	}

	if len(seriesWithQueries) == 0 {
		return fmt.Errorf("no series with queries found")
	}

	feed, err := s.nyaaService.GetRSS(ctx)
	if err != nil {
		return fmt.Errorf("failed to get RSS feed: %w", err)
	}

	if len(feed) == 0 {
		return fmt.Errorf("feed is empty")
	} else {
		s.log.Info("got rss feed", zap.Int("items", len(feed)))
	}

	var timestamp time.Time

	if feed[0].Timestamp != nil {
		timestamp = *feed[0].Timestamp
	}

	feedLenBefore := len(feed)
	feed = pie.Reverse(pie.Filter(feed, func(rss search.ResultRSS) bool {
		if rss.Timestamp != nil {
			return rss.Timestamp.After(lastRss.Timestamp)
		}

		curId, _ := strconv.Atoi(rss.ID)

		return curId >= lastRssId
	}))
	feedLenAfter := len(feed)

	if feedLenAfter < feedLenBefore {
		s.log.Info("removed old feed items", zap.Int("newCount", feedLenAfter))
	}

	if feedLenAfter == 0 {
		return fmt.Errorf("no new feed items since last time")
	}

	newCounter := 0

	for _, series := range seriesWithQueries {
		select {
		case <-ctx.Done():
			return fmt.Errorf("rss poll interrupted")
		default:
		}

		query := series.Query
		if query == nil {
			s.log.Error("query is nil",
				zap.Uint("seriesId", series.ID),
				zap.String("seriesTitle", series.Title))
			continue
		}

		queryValue := (*query).Data()

		for _, result := range feed {
			if !s.test(ctx, &result, &queryValue) {
				continue
			}

			if s.onMatch(ctx, series.ID, queryValue.Auto, result.ID, result.Title) {
				newCounter++
			}
		}
	}

	if newCounter > 0 {
		s.log.Info("finished adding new torrents", zap.Int("count", newCounter))
	} else {
		s.log.Info("no new torrents found", zap.Int("count", newCounter))
	}

	if err := s.lastRssRepo.SetLast(db.LastRSSUpdate{
		RssId:     feed[0].ID,
		Timestamp: timestamp,
	}); err != nil {
		return fmt.Errorf("failed to save last rss timestamp: %w", err)
	}

	return nil
}

func (s *SearchService) PollRSSRoutine(ctx context.Context) {
	s.pollWg.Add(1)
	defer s.pollWg.Done()

	ticker := time.NewTicker(time.Duration(s.config.Search.RssIntervalSec) * time.Second)
	for {
		select {
		case <-ctx.Done():
			s.log.Warn("exiting poll rss routine")
			ticker.Stop()
			return
		case <-s.pollTriggerChan:
			s.log.Info("starting poll manually")

			err := s.doPoll(ctx)
			if err != nil {
				s.log.Error("poll error", zap.Error(err))
			} else {
				s.log.Info("poll success")
			}
		case <-ticker.C:
			s.log.Info("starting poll")

			err := s.doPoll(ctx)
			if err != nil {
				s.log.Error("poll error", zap.Error(err))
			} else {
				s.log.Info("poll success")
			}
		}
	}
}

func startRSSPoll(lifecycle fx.Lifecycle, searchService *SearchService) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(_ context.Context) error {
				go searchService.PollRSSRoutine(searchService.pollCtx)
				return nil
			},
			OnStop: func(_ context.Context) error {
				searchService.pollCancel()
				searchService.pollWg.Wait()
				return nil
			},
		},
	)
}

var SearchExport = fx.Options(fx.Provide(NewSearchService), fx.Invoke(startRSSPoll))

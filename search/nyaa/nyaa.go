package nyaa

import (
	"anileha/config"
	"anileha/search"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
	goCache "github.com/patrickmn/go-cache"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	config      *config.Config
	log         *zap.Logger
	rateLimiter *rate.Limiter
	client      *http.Client
	cache       *goCache.Cache
}

var _ search.Provider = (*Service)(nil)

func NewService(
	config *config.Config,
	log *zap.Logger,
) (*Service, error) {
	rl, client, err := search.InitClientAndRateLimit(config)
	if err != nil {
		return nil, err
	}

	cache := goCache.New(5*time.Minute, 10*time.Minute)

	return &Service{
		log:         log,
		config:      config,
		rateLimiter: rl,
		client:      client,
		cache:       cache,
	}, nil
}

const baseUrl = "https://nyaa.si"

func (s *Service) Search(ctx context.Context, query search.Query) ([]search.Result, error) {
	const torrentsSelector = "body > div > div.table-responsive > table > tbody > tr"
	const viewLinkSelector = "td:nth-child(2) > a:last-child"
	const sizeSelector = "td:nth-child(4)"
	const dateSelector = "td:nth-child(5)"
	const seedersSelector = "td:nth-child(6)"

	req, err := s.genRequest(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error generating request: %w", err)
	}

	doc, err := search.LoadDocument(ctx, s.client, s.rateLimiter, s.log, req)
	if err != nil {
		return nil, fmt.Errorf("failed to laod document: %w", err)
	}

	results := make([]search.Result, 0, 75)

	doc.Find(torrentsSelector).Each(func(i int, sel *goquery.Selection) {
		linkHtml := sel.Find(viewLinkSelector)
		title := strings.TrimSpace(linkHtml.Text())

		viewLinkRelative, _ := linkHtml.Attr("href")
		viewLinkRelative = strings.TrimSpace(viewLinkRelative)
		viewLink := baseUrl + viewLinkRelative

		size := strings.TrimSpace(sel.Find(sizeSelector).Text())
		date := strings.TrimSpace(sel.Find(dateSelector).Text())
		seedersStr := strings.TrimSpace(sel.Find(seedersSelector).Text())
		seeders, _ := strconv.Atoi(seedersStr)

		results = append(results, search.Result{
			ID:      strings.TrimPrefix(viewLinkRelative, "/view/"),
			Title:   title,
			Seeders: seeders,
			Size:    size,
			Date:    date,
			Link:    viewLink,
		})
	})

	for _, res := range results {
		s.cache.Set(res.ID, res, 0)
	}

	return results, nil
}

func (s *Service) DownloadById(ctx context.Context, id string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/download/%s.torrent", baseUrl, id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	bytes, err := search.DownloadFile(ctx, s.client, s.rateLimiter, req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return bytes, nil
}

func (s *Service) GetById(ctx context.Context, id string) (search.ResultById, error) {
	const filesSelector = "div.torrent-file-list.panel-body > ul > li"
	const downloadLinkSelector = "body > div > div.panel.panel-success > div.panel-footer.clearfix > a:nth-child(1)"

	viewLink := baseUrl + "/view/" + id
	req, err := http.NewRequestWithContext(ctx, "GET", viewLink, nil)
	if err != nil {
		return search.ResultById{}, fmt.Errorf("failed to create extra request: %w", err)
	}

	doc, err := search.LoadDocument(ctx, s.client, s.rateLimiter, s.log, req)
	if err != nil {
		return search.ResultById{}, fmt.Errorf("failed to laod document: %w", err)
	}

	files := make([]string, 0, 32)

	doc.Find(filesSelector).Each(func(i int, sel *goquery.Selection) {
		files = append(files, strings.TrimSpace(sel.Text()))
	})

	downloadLink := doc.Find(downloadLinkSelector)
	relativeDownloadUrl, _ := downloadLink.Attr("href")
	relativeDownloadUrl = strings.TrimSpace(relativeDownloadUrl)

	downloadUrl := baseUrl + relativeDownloadUrl

	return search.ResultById{
		DownloadUrl: downloadUrl,
		Files:       files,
	}, nil
}

func (s *Service) GetRSS(ctx context.Context) ([]search.ResultRSS, error) {
	parser := gofeed.NewParser()
	parser.Client = s.client

	feed, err := parser.ParseURLWithContext(baseUrl+"/rss", ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load rss: %w", err)
	}

	results := make([]search.ResultRSS, 0, len(feed.Items))

	for _, item := range feed.Items {
		results = append(results, search.ResultRSS{
			ID:        strings.TrimPrefix(item.GUID, baseUrl+"/view/"),
			Title:     item.Title,
			Link:      item.GUID,
			Timestamp: item.PublishedParsed,
		})
	}

	return results, nil
}

func (s *Service) genRequest(ctx context.Context, query search.Query) (*http.Request, error) {
	urlQuery := make(url.Values)

	// anime
	urlQuery.Set("c", "1_0")
	urlQuery.Set("q", query.Query)
	// probably trust status
	urlQuery.Set("f", "0")
	urlQuery.Set("o", "desc")
	urlQuery.Set("p", strconv.Itoa(query.Page+1))

	if query.SortType == search.SortSeeders {
		urlQuery.Set("s", "seeders")
	} else {
		urlQuery.Set("s", "id")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseUrl, nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = urlQuery.Encode()

	return req, nil
}

var Export = fx.Options(fx.Provide(NewService))

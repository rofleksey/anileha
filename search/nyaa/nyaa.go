package nyaa

import (
	"anileha/config"
	"anileha/search/core"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Service struct {
	config      *config.Config
	log         *zap.Logger
	rateLimiter *rate.Limiter
	client      *http.Client
}

var _ core.Provider = (*Service)(nil)

func NewService(
	config *config.Config,
	log *zap.Logger,
) (*Service, error) {
	rl, client, err := core.InitClientAndRateLimit(config)
	if err != nil {
		return nil, err
	}
	return &Service{
		log:         log,
		config:      config,
		rateLimiter: rl,
		client:      client,
	}, nil
}

const baseUrl = "https://nyaa.si"
const torrentsSelector = "body > div > div.table-responsive > table > tbody > tr"
const viewLinkSelector = "td:nth-child(2) > a:last-child"
const sizeSelector = "td:nth-child(4)"
const dateSelector = "td:nth-child(5)"
const seedersSelector = "td:nth-child(6)"
const filesSelector = "div.torrent-file-list.panel-body > ul > li"
const descriptionSelector = "#torrent-description"
const downloadLinkSelector = "body > div > div.panel.panel-success > div.panel-footer.clearfix > a:nth-child(1)"

func (s *Service) Search(ctx context.Context, query core.Query) ([]core.Result, error) {
	req, err := s.genRequest(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error generating request: %w", err)
	}

	doc, err := core.LoadDocument(ctx, s.client, s.rateLimiter, s.log, req)
	if err != nil {
		return nil, fmt.Errorf("failed to laod document: %w", err)
	}

	results := make([]core.Result, 0, 75)

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

		results = append(results, core.Result{
			ID:      strings.TrimPrefix(viewLinkRelative, "/view/"),
			Title:   title,
			Seeders: seeders,
			Size:    size,
			Date:    date,
			Link:    viewLink,
			ExtraLoader: func(ctx context.Context) (core.ResultExtra, error) {
				return s.getExtra(ctx, viewLink)
			},
		})
	})

	return results, nil
}

func (s *Service) getExtra(ctx context.Context, viewLink string) (core.ResultExtra, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", viewLink, nil)
	if err != nil {
		return core.ResultExtra{}, fmt.Errorf("failed to create extra request: %w", err)
	}

	doc, err := core.LoadDocument(ctx, s.client, s.rateLimiter, s.log, req)
	if err != nil {
		return core.ResultExtra{}, fmt.Errorf("failed to laod document: %w", err)
	}

	files := make([]string, 0, 32)

	doc.Find(filesSelector).Each(func(i int, sel *goquery.Selection) {
		files = append(files, strings.TrimSpace(sel.Text()))
	})

	description := doc.Find(descriptionSelector).Text()
	description = strings.TrimSpace(description)

	downloadLink := doc.Find(downloadLinkSelector)
	relativeDownloadUrl, _ := downloadLink.Attr("href")
	relativeDownloadUrl = strings.TrimSpace(relativeDownloadUrl)

	downloadUrl := baseUrl + relativeDownloadUrl

	return core.ResultExtra{
		DownloadUrl: downloadUrl,
		Description: description,
		Files:       files,
		DownloadTorrent: func(ctx context.Context) ([]byte, error) {
			downloadReq, err := http.NewRequestWithContext(ctx, "GET", downloadUrl, nil)
			if err != nil {
				return nil, err
			}

			return core.DownloadFile(ctx, s.client, s.rateLimiter, downloadReq)
		},
	}, nil
}

func (s *Service) genRequest(ctx context.Context, query core.Query) (*http.Request, error) {
	urlQuery := make(url.Values)

	// anime
	urlQuery.Set("c", "1_0")
	urlQuery.Set("q", query.Query)
	// probably trust status
	urlQuery.Set("f", "0")

	if query.SortType == core.SortSeeders {
		urlQuery.Set("s", "seeders")
	} else {
		urlQuery.Set("s", "id")
	}

	urlQuery.Set("o", "desc")

	if query.Page > 0 {
		urlQuery.Set("p", strconv.Itoa(query.Page))
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseUrl, nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = urlQuery.Encode()

	return req, nil
}

var Export = fx.Options(fx.Provide(NewService))

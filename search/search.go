package search

import (
	"anileha/config"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Provider interface {
	GetRSS(ctx context.Context) ([]ResultRSS, error)
	Search(ctx context.Context, query Query) ([]Result, error)
	GetById(ctx context.Context, id string) (ResultById, error)
	DownloadById(ctx context.Context, id string) ([]byte, error)
}

type Sort int

const (
	SortDate    Sort = 0
	SortSeeders Sort = 1
)

type Query struct {
	Query    string
	SortType Sort
	Page     int
}

type ResultRSS struct {
	ID        string
	Title     string
	Link      string
	Timestamp *time.Time
}

type ResultById struct {
	DownloadUrl string
	Files       []string
}

type Result struct {
	ID      string
	Title   string
	Seeders int
	Size    string
	Date    string
	Link    string
}

func InitClientAndRateLimit(config *config.Config) (*rate.Limiter, *http.Client, error) {
	rlInterval := time.Duration(config.Search.RateLimit.IntervalMs) * time.Millisecond
	rlRequests := config.Search.RateLimit.Requests
	rl := rate.NewLimiter(rate.Every(rlInterval), rlRequests)
	var client *http.Client
	if config.Search.Proxy != "" {
		proxyUrl, err := url.Parse(config.Search.Proxy)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse proxy url: %w", err)
		}
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			},
		}
	} else {
		client = http.DefaultClient
	}
	return rl, client, nil
}

func DownloadFile(ctx context.Context, client *http.Client, rl *rate.Limiter, req *http.Request) ([]byte, error) {
	if err := rl.Wait(ctx); err != nil {
		return nil, fmt.Errorf("download cancelled: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}

	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return bodyBytes, nil
}

func LoadDocument(ctx context.Context, client *http.Client, rl *rate.Limiter, log *zap.Logger,
	req *http.Request) (*goquery.Document, error) {
	if err := rl.Wait(ctx); err != nil {
		return nil, fmt.Errorf("search cancelled: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}

	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	body := string(bodyBytes)

	if res.StatusCode != http.StatusOK {
		log.Error("invalid status code",
			zap.Int("code", res.StatusCode),
			zap.String("body", body))
		return nil, fmt.Errorf("got invalid status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}

	return doc, nil
}

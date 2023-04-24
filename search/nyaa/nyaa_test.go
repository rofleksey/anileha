package nyaa

import (
	"anileha/config"
	"anileha/search"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
)

func TestSearch(t *testing.T) {
	ctx := context.Background()
	cfg := config.GetDefaultConfig()

	service, err := NewService(&cfg, zap.NewNop())
	require.Nil(t, err)

	res, err := service.Search(ctx, search.Query{
		Query:    "blue lock erai 1080",
		SortType: search.SortDate,
	})
	require.Nil(t, err)

	first := res[0]

	assert.Equal(t, "1653158", first.ID)
	assert.Equal(t, "[Erai-raws] Blue Lock - 24 END [1080p][Multiple Subtitle] [ENG][POR-BR][SPA-LA][SPA][FRE][GER][ITA][RUS]", first.Title)
	assert.Equal(t, "2023-03-25 18:00", first.Date)
	assert.Equal(t, "https://nyaa.si/view/1653158", first.Link)
}

func TestGetById(t *testing.T) {
	ctx := context.Background()
	cfg := config.GetDefaultConfig()

	service, err := NewService(&cfg, zap.NewNop())
	require.Nil(t, err)

	extra, err := service.GetById(ctx, "1653158")
	require.Nil(t, err)

	assert.Equal(t, "https://nyaa.si/download/1653158.torrent", extra.DownloadUrl)
	require.Equal(t, 1, len(extra.Files))
	assert.Equal(t, "[Erai-raws] Blue Lock - 24 END [1080p][Multiple Subtitle][5CC77890].mkv (1.3 GiB)", extra.Files[0])
}

func TestDownloadById(t *testing.T) {
	ctx := context.Background()
	cfg := config.GetDefaultConfig()

	service, err := NewService(&cfg, zap.NewNop())
	require.Nil(t, err)

	torrentBytes, err := service.DownloadById(ctx, "1653158")
	require.Nil(t, err)

	assert.Equal(t, 27858, len(torrentBytes))
}

func TestGetRSS(t *testing.T) {
	ctx := context.Background()
	cfg := config.GetDefaultConfig()

	service, err := NewService(&cfg, zap.NewNop())
	require.Nil(t, err)

	feed, err := service.GetRSS(ctx)
	require.Nil(t, err)

	assert.Equal(t, 75, len(feed))
}

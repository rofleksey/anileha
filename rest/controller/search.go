package controller

import (
	"anileha/config"
	"anileha/db"
	"anileha/rest/dao"
	"anileha/rest/engine"
	"anileha/search"
	"anileha/search/nyaa"
	"anileha/service"
	"github.com/elliotchance/pie/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

func mapResultToResponse(r search.Result, provider string) dao.SearchResultDao {
	return dao.SearchResultDao{
		ID:       r.ID,
		Title:    r.Title,
		Provider: provider,
		Seeders:  r.Seeders,
		Size:     r.Size,
		Date:     r.Date,
		Link:     r.Link,
	}
}

func mapResultsToResponseSlice(results []search.Result, provider string) []dao.SearchResultDao {
	res := make([]dao.SearchResultDao, 0, len(results))
	for _, r := range results {
		res = append(res, mapResultToResponse(r, provider))
	}
	return res
}

func registerSearchController(
	ginEngine *gin.Engine,
	log *zap.Logger,
	config *config.Config,
	nyaaService *nyaa.Service,
	seriesService *service.SeriesService,
	searchService *service.SearchService,
) {
	searchGroup := ginEngine.Group("/admin/search")
	searchGroup.Use(engine.RoleMiddleware(log, []string{"admin"}))
	searchGroup.POST("/torrent", func(c *gin.Context) {
		var req dao.QueryRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(engine.ErrBadRequest(err.Error()))
			return
		}

		res, err := nyaaService.Search(c.Request.Context(), search.Query{
			Query:    req.Query,
			Page:     req.Page,
			SortType: search.SortSeeders,
		})
		if err != nil {
			c.Error(engine.ErrBadRequest(err.Error()))
			return
		}

		c.JSON(http.StatusOK, mapResultsToResponseSlice(res, "nyaa"))
	})

	searchGroup.POST("/series/setQuery", func(c *gin.Context) {
		var req dao.SeriesQueryRequestDao
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(engine.ErrBadRequest(err.Error()))
			return
		}

		var err error

		if req.Query != nil {
			include := pie.Map(pie.Filter(strings.Fields(strings.TrimSpace(req.Query.Include)), func(s string) bool {
				return len(s) > 0
			}), func(value string) string {
				return strings.ToLower(value)
			})

			exclude := pie.Map(pie.Filter(strings.Fields(strings.TrimSpace(req.Query.Exclude)), func(s string) bool {
				return len(s) > 0
			}), func(value string) string {
				return strings.ToLower(value)
			})

			err = seriesService.SetQuery(req.SeriesID, &db.SeriesQuery{
				Include:    include,
				Exclude:    exclude,
				Provider:   strings.TrimSpace(req.Query.Provider),
				SingleFile: req.Query.SingleFile,
				Auto:       req.Query.Auto,
			})
		} else {
			err = seriesService.SetQuery(req.SeriesID, nil)
		}
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, "OK")
	})
}

var SearchExport = fx.Options(fx.Invoke(registerSearchController))

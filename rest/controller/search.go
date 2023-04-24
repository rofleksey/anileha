package controller

import (
	"anileha/config"
	"anileha/rest/dao"
	"anileha/rest/engine"
	"anileha/search"
	"anileha/search/nyaa"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
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
}

var SearchExport = fx.Options(fx.Invoke(registerSearchController))

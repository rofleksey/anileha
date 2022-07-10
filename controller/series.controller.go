package controller

import (
	dao2 "anileha/dao"
	"anileha/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"net/http"
	"strconv"
)

func registerSeriesController(engine *gin.Engine, seriesService *service.SeriesService) {
	engine.GET("/series", func(c *gin.Context) {
		seriesArr, err := seriesService.GetAllSeries()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, seriesArr)
	})
	engine.GET("/series/:id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse id"})
			return
		}
		series, err := seriesService.GetSeriesById(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, series)
	})
	engine.DELETE("/series/:id", func(c *gin.Context) {
		idString := c.Param("id")
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse id"})
			return
		}
		err = seriesService.DeleteSeriesById(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "OK")
	})
	engine.POST("/series", func(c *gin.Context) {
		var dao dao2.SeriesDao
		if err := c.ShouldBindJSON(&dao); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id, err := seriesService.AddSeries(dao)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, strconv.FormatUint(uint64(id), 10))
	})
}

var SeriesControllerExport = fx.Options(fx.Invoke(registerSeriesController))

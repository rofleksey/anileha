package controller

import (
	"anileha/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"net/http"
)

func registerHealthController(ginEngine *gin.Engine, healthService *service.HealthService) {
	ginEngine.GET("/health", func(c *gin.Context) {
		health := healthService.GetHealth()
		if health {
			c.JSON(http.StatusOK, gin.H{"health": "green"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"health": "red"})
		}
	})
}

var HealthExport = fx.Options(fx.Invoke(registerHealthController))

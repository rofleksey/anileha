package controller

import (
	"anileha/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"net/http"
)

func registerHealthController(engine *gin.Engine, healthService *service.HealthService) {
	engine.GET("/health", func(c *gin.Context) {
		health := healthService.GetHealth()
		if health {
			c.JSON(http.StatusOK, gin.H{"health": "green"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"health": "red"})
		}
	})
}

var HealthControllerExport = fx.Options(fx.Invoke(registerHealthController))

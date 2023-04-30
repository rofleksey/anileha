package engine

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

func ErrorMiddleware(log *zap.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Next()

		ginError := c.Errors.Last()
		if ginError == nil {
			return
		}

		lastError := ginError.Err
		if lastError == nil {
			return
		}

		statusError, statusErrorOk := lastError.(*StatusError)
		if statusErrorOk {
			if statusError.StatusCode == http.StatusInternalServerError {
				c.JSON(statusError.StatusCode, gin.H{"error": "internal server error"})
				log.Error("internal server error", zap.Error(statusError))
				return
			}
			c.JSON(statusError.StatusCode, gin.H{"error": statusError.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError,
			gin.H{"error": fmt.Sprintf("unhandled error: %s", lastError.Error())})
	}
}

package middleware

import (
	"net/http"

	"github.com/baothaihcmut/Storage-app/internal/common/exception"
	"github.com/baothaihcmut/Storage-app/internal/common/response"
	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			status := exception.ErrorStatusMapper(c.Errors[0])
			if status != http.StatusInternalServerError {
				c.JSON(status, response.InitResponse(false, c.Errors[0].Error(), nil))
				return
			}
			c.JSON(http.StatusInternalServerError, response.InitResponse(false, "Internal error", nil))
		}
	}
}

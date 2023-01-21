package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Errors is a gin compatible middleware function responsible for
// handling unexpected errors coming out of the call chain.
func Errors() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Next()

		if lastErr := c.Errors.Last(); lastErr != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, struct {
				Status int    `json:"status"`
				Error  string `json:"error"`
			}{
				Status: http.StatusInternalServerError,
				Error:  lastErr.Error(),
			})
		}
	}
}

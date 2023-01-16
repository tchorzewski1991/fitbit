package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/tchorzewski1991/fitbit/core/web"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// Logger is a gin compatible middleware function responsible for
// handling HTTP logs.
func Logger(log *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Web related ctx values should be always added before
		// any other middleware. Having traceID is mandatory for
		// every request either successful or failed one.
		v, err := web.GetCtxValues(c.Request.Context())
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		c.Next()

		fields := []interface{}{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"trace_id", v.TraceID,
			"status", c.Writer.Status(),
			"size", c.Writer.Size(),
			"took", time.Since(v.StartedAt).String(),
		}

		if lastErr := c.Errors.Last(); lastErr != nil {
			log.Errorw("request error", append([]any{"error", lastErr}, fields...)...)
		} else {
			log.Infow("request finished", fields...)
		}
	}
}

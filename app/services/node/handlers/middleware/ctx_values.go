package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tchorzewski1991/fitbit/core/web"
	"time"
)

// CtxValues is a gin compatible middleware function responsible for
// extending HTTP request context with web related context values like
// TraceID or time when request has started.
func CtxValues() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := web.SetCtxValues(c.Request.Context(), &web.Values{
			TraceID:   uuid.New().String(),
			StartedAt: time.Now(),
		})
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

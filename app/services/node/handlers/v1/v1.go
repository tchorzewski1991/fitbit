package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/tchorzewski1991/fitbit/app/services/node/handlers/v1/private"
	"github.com/tchorzewski1991/fitbit/app/services/node/handlers/v1/public"
	"go.uber.org/zap"
)

const version = "v1"

// Config collects all dependencies required by either public or private v1 handlers.
type Config struct {
	Log *zap.SugaredLogger
}

// PublicHandlers registers all v1 public routes.
func PublicHandlers(mux *gin.Engine, cfg Config) {
	h := public.Handlers{
		Log: cfg.Log,
	}
	v1 := mux.Group("/" + version)
	v1.GET("/health", h.Health)
}

// PrivateHandlers registers all v1 private routes.
func PrivateHandlers(mux *gin.Engine, cfg Config) {
	h := private.Handlers{
		Log: cfg.Log,
	}
	v1 := mux.Group("/" + version)
	v1.GET("/health", h.Health)
}

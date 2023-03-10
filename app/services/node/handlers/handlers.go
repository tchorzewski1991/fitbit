package handlers

import (
	"github.com/gin-contrib/cors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tchorzewski1991/fitbit/app/services/node/handlers/middleware"
	v1 "github.com/tchorzewski1991/fitbit/app/services/node/handlers/v1"
	"github.com/tchorzewski1991/fitbit/core/blockchain/state"
	"github.com/tchorzewski1991/fitbit/core/nameservice"
	"go.uber.org/zap"
)

// Config collects all dependencies required by either public or private handlers.
type Config struct {
	Log         *zap.SugaredLogger
	State       *state.State
	NameService *nameservice.NameService
}

// PublicMux constructs a new http.Handler and registers all public routes.
// Every new version of public routes should be registered here.
func PublicMux(cfg Config) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	mux := gin.New()

	mux.Use(
		middleware.CtxValues(),
		middleware.Logger(cfg.Log),
		middleware.Errors(),
		cors.Default(),
	)

	v1.PublicHandlers(mux, v1.Config{
		Log:         cfg.Log,
		State:       cfg.State,
		NameService: cfg.NameService,
	})

	return mux
}

// PrivateMux constructs a new http.Handler and registers all private routes.
// Every new version of private routes should be registered here.
func PrivateMux(cfg Config) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	mux := gin.New()

	mux.Use(
		middleware.CtxValues(),
		middleware.Logger(cfg.Log),
		middleware.Errors(),
	)

	v1.PrivateHandlers(mux, v1.Config{
		Log:   cfg.Log,
		State: cfg.State,
	})

	return mux
}

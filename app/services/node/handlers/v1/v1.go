package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/tchorzewski1991/fitbit/app/services/node/handlers/v1/private"
	"github.com/tchorzewski1991/fitbit/app/services/node/handlers/v1/public"
	"github.com/tchorzewski1991/fitbit/core/blockchain/state"
	"github.com/tchorzewski1991/fitbit/core/nameservice"
	"go.uber.org/zap"
)

const version = "v1"

// Config collects all dependencies required by either public or private v1 handlers.
type Config struct {
	Log         *zap.SugaredLogger
	State       *state.State
	NameService *nameservice.NameService
}

// PublicHandlers registers all v1 public routes.
func PublicHandlers(mux *gin.Engine, cfg Config) {
	h := public.Handlers{
		Log:         cfg.Log,
		State:       cfg.State,
		NameService: cfg.NameService,
	}
	v1 := mux.Group("/" + version)
	v1.GET("/health", h.Health)
	v1.GET("/genesis", h.Genesis)
	v1.GET("/accounts", h.Accounts)
	v1.GET("/accounts/:address", h.Account)
	v1.GET("/tx/uncommitted", h.UncommittedWalletTx)
	v1.GET("/tx/uncommitted/:address", h.UncommittedWalletAccountTx)
	v1.POST("/tx/submit", h.SubmitWalletTx)
}

// PrivateHandlers registers all v1 private routes.
func PrivateHandlers(mux *gin.Engine, cfg Config) {
	h := private.Handlers{
		Log: cfg.Log,
	}
	v1 := mux.Group("/" + version)
	v1.GET("/health", h.Health)
}

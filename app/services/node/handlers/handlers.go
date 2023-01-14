package handlers

import (
	"go.uber.org/zap"
	"net/http"
)

type Config struct {
	Log *zap.SugaredLogger
}

// PublicMux takes a Config and constructs http.Handler with all public routes.
func PublicMux(c Config) http.Handler {
	return nil
}

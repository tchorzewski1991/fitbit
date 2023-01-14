package private

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handlers struct {
	Log *zap.SugaredLogger
}

func (h Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	})
}

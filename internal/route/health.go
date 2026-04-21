package route

import (
	"blog-post-gin/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupHealthRouter(rg *gin.RouterGroup, h *handler.HealthHandler) {
	health := rg.Group("/health")

	health.GET("", h.CheckHealth)
}

package route

import (
	"blog-post-gin/internal/handler"

	"github.com/gin-gonic/gin"
)

type Routes struct {
	HealthHandler *handler.HealthHandler
	AuthHandler   *handler.AuthHandler
}

func SetupRouter(e *gin.Engine, r *Routes) {
	api := e.Group("/api")

	SetupHealthRouter(api, r.HealthHandler)
	SetupAuthRouter(api, r.AuthHandler)
}

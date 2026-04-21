package route

import (
	"blog-post-gin/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupAuthRouter(rg *gin.RouterGroup, h *handler.AuthHandler) {
	auth := rg.Group("/auth")

	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/logout", h.Logout)
}

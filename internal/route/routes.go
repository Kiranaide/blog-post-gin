package route

import (
	"blog-post-gin/internal/handler"
	"blog-post-gin/internal/middleware"
	"blog-post-gin/internal/service"

	"github.com/gin-gonic/gin"
)

type Routes struct {
	HealthHandler *handler.HealthHandler
	AuthHandler   *handler.AuthHandler
	AuthService   *service.AuthService
}

func ProtectedGroup(parent *gin.RouterGroup, relativePath string, auth *service.AuthService, requiredRoles ...string) *gin.RouterGroup {
	group := parent.Group(relativePath)
	group.Use(middleware.Authenticate(auth))

	if len(requiredRoles) > 0 {
		group.Use(middleware.RequireRole(auth, requiredRoles...))
	}

	return group
}

func SetupRouter(e *gin.Engine, r *Routes) {
	api := e.Group("/api")

	SetupHealthRouter(api, r.HealthHandler)
	SetupAuthRouter(api, r.AuthHandler)
}

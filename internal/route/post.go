package route

import (
	"blog-post-gin/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupPostRouter(rg *gin.RouterGroup, h *handler.PostHandler) {
	post := rg.Group("/posts")

	post.POST("/", h.CreatePost)
	post.GET("/", h.GetAllPosts)
	post.GET("/:id", h.GetPostByID)
	post.GET("/:slug", h.GetPostBySlug)
	post.GET("/:tagid", h.GetPostsByTagID)
	post.PATCH("/:id", h.UpdatePost)
	post.DELETE("/:id", h.DeletePost)
	post.GET("/deleted", h.GetDeletedPosts)
	post.PATCH("/restore/:id", h.RestoreDeletedPost)
}

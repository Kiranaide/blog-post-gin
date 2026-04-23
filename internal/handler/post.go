package handler

import (
	"blog-post-gin/internal/dto"
	"blog-post-gin/internal/helper"
	"blog-post-gin/internal/middleware"
	"blog-post-gin/internal/response"
	"blog-post-gin/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	s *service.PostService
}

func NewPostHandler(s *service.PostService) *PostHandler {
	return &PostHandler{s: s}
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	var req dto.CreatePost

	if err := response.BindAndValidate(c, &req); err != nil {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || strings.TrimSpace(userID) == "" {
		response.SendError(c, helper.ErrTokenInvalid.WithMessage("Authentication is required"))
		return
	}

	post, message, err := h.s.CreatePost(req, userID)
	if err != nil {
		response.SendError(c, err)
		return
	}

	response.Created(c, message, post)
}

func (h *PostHandler) GetAllPosts(c *gin.Context) {

	posts, message, err := h.s.GetAllPosts()
	if err != nil {
		response.SendError(c, err)
		return
	}

	response.OK(c, message, posts)
}

func (h *PostHandler) GetPostByID(c *gin.Context) {
	postID := c.Param("id")
	if strings.TrimSpace(postID) == "" {
		response.SendError(c, helper.ErrResourceConflict.WithMessage("Post ID is required"))
		return
	}

	post, message, err := h.s.GetPostByID(postID)
	if err != nil {
		response.SendError(c, err)
		return
	}

	response.OK(c, message, post)
}

func (h *PostHandler) GetPostBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if strings.TrimSpace(slug) == "" {
		response.SendError(c, helper.ErrResourceConflict.WithMessage("Post slug is required"))
		return
	}

	post, message, err := h.s.GetPostBySlug(slug)
	if err != nil {
		response.SendError(c, err)
		return
	}

	response.OK(c, message, post)
}

func (h *PostHandler) GetPostsByTagID(c *gin.Context) {
	tag := c.Param("tag")
	if strings.TrimSpace(tag) == "" {
		response.SendError(c, helper.ErrResourceConflict.WithMessage("Tag is required"))
		return
	}

	posts, message, err := h.s.GetPostsByTagID(tag)
	if err != nil {
		response.SendError(c, err)
		return
	}

	response.OK(c, message, posts)
}

func (h *PostHandler) UpdatePost(c *gin.Context) {
	postID := c.Param("id")
	if strings.TrimSpace(postID) == "" {
		response.SendError(c, helper.ErrResourceConflict.WithMessage("Post ID is required"))
		return
	}

	var req dto.UpdatePost
	if err := response.BindAndValidate(c, &req); err != nil {
		return
	}

	post, message, err := h.s.UpdatePost(postID, req)
	if err != nil {
		response.SendError(c, err)
		return
	}

	response.OK(c, message, post)
}

func (h *PostHandler) DeletePost(c *gin.Context) {
	postID := c.Param("id")
	if strings.TrimSpace(postID) == "" {
		response.SendError(c, helper.ErrResourceConflict.WithMessage("Post ID is required"))
		return
	}

	message, err := h.s.DeletePost(postID)
	if err != nil {
		response.SendError(c, err)
		return
	}

	response.OK(c, message, "")
}

func (h *PostHandler) GetDeletedPosts(c *gin.Context) {

	posts, message, err := h.s.GetDeletedPosts()
	if err != nil {
		response.SendError(c, err)
		return
	}

	response.OK(c, message, posts)
}

func (h *PostHandler) RestoreDeletedPost(c *gin.Context) {
	postID := c.Param("id")
	if strings.TrimSpace(postID) == "" {
		response.SendError(c, helper.ErrResourceConflict.WithMessage("Post ID is required"))
		return
	}

	post, message, err := h.s.RestoreDeletedPost(postID)
	if err != nil {
		response.SendError(c, err)
		return
	}

	response.OK(c, message, post)
}

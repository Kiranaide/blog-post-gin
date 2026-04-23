package dto

import (
	"github.com/google/uuid"
)

type CreatePost struct {
	Title   string      `json:"title" binding:"required,min=10"`
	TagID   []uuid.UUID `json:"tag_id" binding:"required"`
	Content string      `json:"content" binding:"required,min=10"`
}

type UpdatePost struct {
	Title   *string     `json:"title"`
	TagID   []uuid.UUID `json:"tag_id" binding:"required"`
	Content *string     `json:"content"`
}

package response

import (
	"blog-post-gin/internal/entity"
	"time"
)

type CommentResponse struct {
	ID        string       `json:"id"`
	PostID    string       `json:"post_id"`
	Comment   string       `json:"comment"`
	User      UserResponse `json:"user"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

func ToCommentResponse(c entity.Comment) CommentResponse {
	return CommentResponse{
		ID:        c.ID.String(),
		PostID:    c.PostID.String(),
		Comment:   c.Comment,
		User:      ToUserResponse(c.User),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

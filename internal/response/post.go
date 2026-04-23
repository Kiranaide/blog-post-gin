package response

import (
	"blog-post-gin/internal/entity"
	"time"
)

type PostResponse struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	Slug      string            `json:"slug"`
	Content   string            `json:"content"`
	Author    UserResponse      `json:"author"`
	Tags      []TagResponse     `json:"tags"`
	Comments  []CommentResponse `json:"comments"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

func ToPostResponse(p entity.Post) PostResponse {
	return PostResponse{
		ID:      p.ID.String(),
		Title:   p.Title,
		Slug:    p.Slug,
		Content: p.Content,
		Author:  ToUserResponse(p.Author),
		Tags: func(tags []*entity.Tag) []TagResponse {
			res := make([]TagResponse, len(tags))
			for i, tag := range tags {
				res[i] = ToTagResponse(*tag)
			}
			return res
		}(p.Tags),
		Comments: func(comments []entity.Comment) []CommentResponse {
			res := make([]CommentResponse, len(comments))
			for i, comment := range comments {
				res[i] = ToCommentResponse(comment)
			}
			return res
		}(p.Comments),
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

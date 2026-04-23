package response

import "blog-post-gin/internal/entity"

type TagResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func ToTagResponse(t entity.Tag) TagResponse {
	return TagResponse{
		ID:   t.ID.String(),
		Name: t.Name,
	}
}

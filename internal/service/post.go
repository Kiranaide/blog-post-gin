package service

import (
	"blog-post-gin/internal/dto"
	"blog-post-gin/internal/entity"
	"blog-post-gin/internal/helper"
	"blog-post-gin/internal/response"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostService struct {
	db *gorm.DB
}

func NewPostService(db *gorm.DB) *PostService {
	return &PostService{db: db}
}

func (s *PostService) CreatePost(req dto.CreatePost, userID string) (*response.PostResponse, string, error) {
	var tags []*entity.Tag
	if len(req.TagID) > 0 {
		if err := s.db.Where("id IN ?", req.TagID).Find(&tags).Error; err != nil {
			return nil, "", helper.ErrInternal.WithCause(err)
		}

		if len(tags) != len(req.TagID) {
			return nil, "", helper.ErrResourceNotFound.WithMessage("One or more tags not found")
		}
	}

	authorID, err := uuid.Parse(strings.TrimSpace(userID))
	if err != nil {
		return nil, "", helper.ErrTokenInvalid.WithMessage("Authentication is required")
	}

	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	slug := helper.GenerateUniqueSlug(req.Title)

	payload := entity.Post{
		Title:    req.Title,
		Tags:     tags,
		Slug:     slug,
		Content:  req.Content,
		AuthorID: authorID,
	}

	if err := trx.Create(&payload).Error; err != nil {
		trx.Rollback()
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	if err := trx.Commit().Error; err != nil {
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	if err := s.db.Preload("Author").Preload("Tags").Preload("Comments").First(&payload, "id = ?", payload.ID).Error; err != nil {
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	data := response.ToPostResponse(payload)

	return &data, "Post created successfully", nil
}

func (s *PostService) GetAllPosts() (*[]response.PostResponse, string, error) {
	var posts []entity.Post

	if err := s.db.Preload("Author").Preload("Tags").Find(&posts).Error; err != nil {
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	if len(posts) == 0 {
		emptyData := []response.PostResponse{}
		return &emptyData, "No posts found", nil
	}

	data := make([]response.PostResponse, len(posts))
	for i, post := range posts {
		data[i] = response.ToPostResponse(post)
	}

	return &data, "Posts retrieved successfully", nil
}

func (s *PostService) GetPostByID(id string) (*response.PostResponse, string, error) {
	var post entity.Post

	postID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, "", helper.ErrResourceNotFound.WithMessage("Post ID is invalid")
	}

	if err := s.db.Preload("Author").Preload("Tags").Preload("Comments").First(&post, "id = ?", postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", helper.ErrResourceNotFound.WithMessage("Post not found")
		}
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	data := response.ToPostResponse(post)

	return &data, "Post retrieved successfully", nil
}

func (s *PostService) GetPostBySlug(slug string) (*response.PostResponse, string, error) {
	var post entity.Post

	if err := s.db.Preload("Author").Preload("Tags").Preload("Comments").First(&post, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", helper.ErrResourceNotFound.WithMessage("Post not found")
		}
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	data := response.ToPostResponse(post)

	return &data, "Post retrieved successfully", nil
}

func (s *PostService) GetPostsByTagID(tagID string) (*[]response.PostResponse, string, error) {
	var posts []entity.Post

	tagUUID, err := uuid.Parse(strings.TrimSpace(tagID))
	if err != nil {
		return nil, "", helper.ErrResourceNotFound.WithMessage("Tag ID is invalid")
	}

	if err := s.db.Joins("JOIN post_tags ON post_tags.post_id = posts.id").
		Where("post_tags.tag_id = ?", tagUUID).
		Preload("Author").
		Preload("Tags").
		Find(&posts).Error; err != nil {
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	if len(posts) == 0 {
		return &[]response.PostResponse{}, "No posts found for this tag", nil
	}

	data := make([]response.PostResponse, len(posts))
	for i, post := range posts {
		data[i] = response.ToPostResponse(post)
	}

	return &data, "Posts retrieved successfully", nil

}

func (s *PostService) UpdatePost(id string, req dto.UpdatePost) (*response.PostResponse, string, error) {
	var post entity.Post

	postID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, "", helper.ErrResourceNotFound.WithMessage("Post ID is invalid")
	}

	if err := s.db.Preload("Tags").First(&post, "id = ?", postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", helper.ErrResourceNotFound.WithMessage("Post not found")
		}
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	if req.Title != nil && *req.Title != post.Title {
		post.Title = *req.Title
		slug := helper.GenerateUniqueSlug(post.Title)
		if err := trx.Model(&entity.Post{}).Where("slug = ? AND id <> ?", slug, post.ID).Error; err != nil {
			trx.Rollback()
			return nil, "", helper.ErrInternal.WithCause(err)
		}
	}

	if req.Content != nil {
		post.Content = *req.Content
	}

	if req.TagID != nil {
		var tags []*entity.Tag
		if len(req.TagID) > 0 {
			if err := trx.Where("id IN ?", req.TagID).Find(&tags).Error; err != nil {
				trx.Rollback()
				return nil, "", helper.ErrInternal.WithCause(err)
			}
			if len(tags) != len(req.TagID) {
				trx.Rollback()
				return nil, "", helper.ErrResourceNotFound.WithMessage("One or more tags not found")
			}
		}
		if err := trx.Model(&post).Association("Tags").Replace(tags); err != nil {
			trx.Rollback()
			return nil, "", helper.ErrInternal.WithCause(err)
		}
	}

	if err := trx.Save(&post).Error; err != nil {
		trx.Rollback()
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	if err := trx.Commit().Error; err != nil {
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	if err := s.db.Preload("Author").Preload("Tags").Preload("Comments").First(&post, "id = ?", post.ID).Error; err != nil {
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	data := response.ToPostResponse(post)

	return &data, "Post updated successfully", nil
}

func (s *PostService) DeletePost(id string) (string, error) {
	postID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return "", helper.ErrResourceNotFound.WithMessage("Post ID is invalid")
	}

	deleted := s.db.Delete(&entity.Post{}, "id = ?", postID)
	if deleted.Error != nil {
		return "", helper.ErrInternal.WithCause(deleted.Error)
	}
	if deleted.RowsAffected == 0 {
		return "", helper.ErrResourceNotFound.WithMessage("Post not found")
	}

	return "Post deleted successfully", nil
}

func (s *PostService) GetDeletedPosts() (*[]response.PostResponse, string, error) {
	var posts []entity.Post

	if err := s.db.Unscoped().
		Where("deleted_at IS NOT NULL").
		Preload("Author").Preload("Tags").
		Find(&posts).Error; err != nil {
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	if len(posts) == 0 {
		emptyData := []response.PostResponse{}
		return &emptyData, "No deleted posts found", nil
	}

	data := make([]response.PostResponse, len(posts))
	for i, post := range posts {
		data[i] = response.ToPostResponse(post)
	}

	return &data, "Deleted posts retrieved successfully", nil
}

func (s *PostService) RestoreDeletedPost(id string) (*response.PostResponse, string, error) {
	var post entity.Post

	postID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, "", helper.ErrResourceNotFound.WithMessage("Post ID is invalid")
	}

	if err := s.db.Unscoped().Preload("Author").Preload("Tags").Preload("Comments").First(&post, "id = ?", postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", helper.ErrResourceNotFound.WithMessage("Post not found")
		}
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	if !post.DeletedAt.Valid {
		return nil, "", helper.ErrResourceConflict.WithMessage("Post is not deleted")
	}

	if err := s.db.Unscoped().Model(&post).Update("deleted_at", nil).Error; err != nil {
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	if err := s.db.Preload("Author").Preload("Tags").Preload("Comments").First(&post, "id = ?", post.ID).Error; err != nil {
		return nil, "", helper.ErrInternal.WithCause(err)
	}

	data := response.ToPostResponse(post)

	return &data, "Post restored successfully", nil
}

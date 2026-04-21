package response

import (
	"blog-post-gin/internal/entity"

	"github.com/google/uuid"
)

type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Username string    `json:"username"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	User        UserResponse `json:"user"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type RegisterResponse struct {
	User UserResponse `json:"user"`
}

func ToUserResponse(u entity.User) UserResponse {
	return UserResponse{
		ID:       u.ID,
		Name:     u.Name,
		Username: u.Username,
	}
}

func ToLoginResponse(accessToken string, user entity.User) LoginResponse {
	return LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		User:        ToUserResponse(user),
	}
}

func ToRefreshResponse(accessToken string) RefreshResponse {
	return RefreshResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}
}

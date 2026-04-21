package middleware

import (
	"blog-post-gin/internal/entity"
	"blog-post-gin/internal/helper"
	"blog-post-gin/internal/response"
	"blog-post-gin/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ContextClaimsKey   = "auth.claims"
	ContextUserIDKey   = "auth.user_id"
	ContextUsernameKey = "auth.username"
)

func Authenticate(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			response.SendError(c, helper.ErrTokenInvalid.WithMessage("Authorization header is required"))
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			response.SendError(c, helper.ErrTokenInvalid.WithMessage("Authorization header must start with 'Bearer '"))
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
		if tokenString == "" {
			response.SendError(c, helper.ErrTokenInvalid.WithMessage("Token is required"))
			return
		}

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			response.SendError(c, err)
			return
		}

		c.Set(ContextClaimsKey, claims)
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUsernameKey, claims.Username)
		c.Next()
	}
}

func GetClaims(c *gin.Context) (*entity.JWTClaims, bool) {
	claims, exists := c.Get(ContextClaimsKey)
	if !exists {
		return nil, false
	}
	castedClaims, ok := claims.(*entity.JWTClaims)
	return castedClaims, ok
}

func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(ContextUserIDKey)
	if !exists {
		return "", false
	}
	castedUserID, ok := userID.(string)
	return castedUserID, ok
}

func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get(ContextUsernameKey)
	if !exists {
		return "", false
	}
	castedUsername, ok := username.(string)
	return castedUsername, ok
}

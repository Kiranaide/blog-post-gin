package middleware

import (
	"blog-post-gin/internal/helper"
	"blog-post-gin/internal/response"
	"blog-post-gin/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

func RequireRole(auth *service.AuthService, requiredRoles ...string) gin.HandlerFunc {
	allowedRoles := normalizeRoleList(requiredRoles)

	return func(c *gin.Context) {
		if auth == nil {
			response.SendError(c, helper.ErrInternal.WithMessage("Auth service is not configured"))
			return
		}
		if len(allowedRoles) == 0 {
			response.SendError(c, helper.ErrInternal.WithMessage("No roles specified for this endpoint"))
			return
		}

		userID, ok := GetUserID(c)
		if !ok || strings.TrimSpace(userID) == "" {
			response.SendError(c, helper.ErrTokenInvalid.WithMessage("Authentication is required"))
			return
		}

		tokenRole, ok := GetRole(c)
		if !ok || strings.TrimSpace(tokenRole) == "" {
			response.SendError(c, helper.ErrTokenInvalid.WithMessage("Invalid token role"))
			return
		}

		dbRole, err := auth.GetUserRoleByID(userID)
		if err != nil {
			if appErr, isAppErr := helper.AsError(err); isAppErr && appErr.HTTPStatus >= 500 {
				response.SendError(c, err)
				return
			}
			response.SendError(c, helper.ErrTokenInvalid.WithMessage("Failed to retrieve user role"))
			return
		}

		normalizedTokenRole := strings.ToLower(strings.TrimSpace(tokenRole))
		normalizedDBRole := strings.ToLower(strings.TrimSpace(dbRole))

		if normalizedTokenRole != normalizedDBRole {
			response.SendError(c, helper.ErrTokenInvalid.WithMessage("Token role does not match database role"))
			return
		}

		if !allowedRoles[normalizedDBRole] {
			response.SendError(c, helper.ErrInsufficientRole.WithMessage("You do not have permission to access this resource"))
			return
		}

		c.Next()
	}
}

func normalizeRoleList(roles []string) map[string]bool {
	out := make(map[string]bool, len(roles))
	for _, role := range roles {
		n := strings.ToLower(strings.TrimSpace(role))
		if n != "" {
			out[n] = true
		}
	}
	return out
}

package handler

import (
	"blog-post-gin/internal/dto"
	"blog-post-gin/internal/entity"
	"blog-post-gin/internal/helper"
	"blog-post-gin/internal/response"
	"blog-post-gin/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	s  *service.AuthService
	sc entity.SessionCookieConfig
}

func NewAuthHandler(s *service.AuthService, sc entity.SessionCookieConfig) *AuthHandler {
	return &AuthHandler{s: s, sc: sc}
}

// Register godoc
// @Summary Register user
// @Description Create new account with username and password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register request"
// @Success 201 {object} response.BaseResponse{data=response.RegisterResponse}
// @Failure 400 {object} response.ErrorResponseDTO
// @Failure 401 {object} response.ErrorResponseDTO
// @Failure 500 {object} response.ErrorResponseDTO
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := response.BindAndValidate(c, &req); err != nil {
		return
	}

	user, message, err := h.s.Register(req)
	if err != nil {
		response.SendError(c, err)
		return
	}

	response.Created(c, message, user)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and issue access token plus refresh cookie.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login request"
// @Success 200 {object} response.BaseResponse{data=response.LoginResponse}
// @Failure 400 {object} response.ErrorResponseDTO
// @Failure 401 {object} response.ErrorResponseDTO
// @Failure 500 {object} response.ErrorResponseDTO
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := response.BindAndValidate(c, &req); err != nil {
		return
	}

	data, refreshToken, message, err := h.s.Login(req, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		response.SendError(c, err)
		return
	}

	h.setSessionCookie(c, refreshToken)
	response.OK(c, message, data)
}

// Refresh godoc
// @Summary Refresh access token
// @Description Use refresh token cookie to get new access token and rotate refresh token.
// @Tags Auth
// @Produce json
// @Success 200 {object} response.BaseResponse{data=response.RefreshResponse}
// @Failure 400 {object} response.ErrorResponseDTO
// @Failure 401 {object} response.ErrorResponseDTO
// @Failure 500 {object} response.ErrorResponseDTO
// @Router /api/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	cookie := h.s.SessionCookieConfig()

	rawToken, err := c.Cookie(cookie.Name)
	if err != nil || strings.TrimSpace(rawToken) == "" {
		response.SendError(c, err)
		return
	}

	data, refreshToken, message, err := h.s.Refresh(rawToken, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		response.SendError(c, helper.ErrRefreshTokenInvalid.WithMessage("Refresh token is required"))
		return
	}

	h.setSessionCookie(c, refreshToken)
	response.OK(c, message, data)
}

// Logout godoc
// @Summary Logout user
// @Description Revoke refresh token when present and clear session cookie.
// @Tags Auth
// @Produce json
// @Success 200 {object} response.BaseResponse{data=object}
// @Failure 500 {object} response.ErrorResponseDTO
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	cookie := h.s.SessionCookieConfig()
	rawToken, _ := c.Cookie(cookie.Name)

	token := strings.TrimSpace(rawToken)
	if token != "" {
		if err := h.s.Logout(token); err != nil {
			response.SendError(c, err)
			return
		}
	}

	h.clearSessionCookie(c)
	response.OK(c, "Logged out successfully", gin.H{})
}

func (h *AuthHandler) setSessionCookie(c *gin.Context, value string) {
	cfg := h.s.SessionCookieConfig()
	c.SetSameSite(parseSameSite(cfg.SameSite))
	c.SetCookie(cfg.Name, value, h.s.SessionCookieMaxAgeSeconds(), cfg.Path, cfg.Domain, cfg.Secure, cfg.HTTPOnly)
}

func (h *AuthHandler) clearSessionCookie(c *gin.Context) {
	cfg := h.s.SessionCookieConfig()
	c.SetSameSite(parseSameSite(cfg.SameSite))
	c.SetCookie(cfg.Name, "", -1, cfg.Path, cfg.Domain, cfg.Secure, cfg.HTTPOnly)
}

func parseSameSite(sameSite string) http.SameSite {
	switch strings.ToLower(sameSite) {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}

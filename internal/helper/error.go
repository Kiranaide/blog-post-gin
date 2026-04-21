package helper

import (
	"errors"
	"fmt"
	"net/http"
)

const requestIDKey = "request_id"

type AppError struct {
	HTTPStatus int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    any    `json:"details,omitempty"`
	Cause      error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Cause }

func (e *AppError) WithCause(err error) *AppError {
	clone := *e
	clone.Cause = err
	return &clone
}

func (e *AppError) WithDetails(d any) *AppError {
	clone := *e
	clone.Details = d
	return &clone
}

func (e *AppError) WithMessage(msg string) *AppError {
	clone := *e
	clone.Message = msg
	return &clone
}

func NewAppError(httpStatus int, code, message string) *AppError {
	return &AppError{HTTPStatus: httpStatus, Code: code, Message: message}
}

func BadRequest(code, message string) *AppError {
	return NewAppError(http.StatusBadRequest, code, message)
}

func Unauthorized(code, message string) *AppError {
	return NewAppError(http.StatusUnauthorized, code, message)
}

func Forbidden(code, message string) *AppError {
	return NewAppError(http.StatusForbidden, code, message)
}

func NotFound(code, message string) *AppError {
	return NewAppError(http.StatusNotFound, code, message)
}

func Conflict(code, message string) *AppError {
	return NewAppError(http.StatusConflict, code, message)
}

func UnprocessableEntity(code, message string) *AppError {
	return NewAppError(http.StatusUnprocessableEntity, code, message)
}

func TooManyRequests(code, message string) *AppError {
	return NewAppError(http.StatusTooManyRequests, code, message)
}

func Internal(code, message string) *AppError {
	return NewAppError(http.StatusInternalServerError, code, message)
}

func ServiceUnavailable(code, message string) *AppError {
	return NewAppError(http.StatusServiceUnavailable, code, message)
}

func AsError(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

// ── Auth ───────────────────────────────────────────────────────────────────────
var (
	ErrInvalidCredentials  = Unauthorized("auth.invalid_credentials", "Invalid email or password")
	ErrTokenExpired        = Unauthorized("auth.token_expired", "Access token has expired")
	ErrTokenInvalid        = Unauthorized("auth.token_invalid", "Access token is invalid or malformed")
	ErrRefreshTokenExpired = Unauthorized("auth.refresh_token_expired", "Refresh token has expired")
	ErrRefreshTokenInvalid = Unauthorized("auth.refresh_token_invalid", "Refresh token is invalid or malformed")
	ErrInsufficientRole    = Forbidden("auth.insufficient_role", "You do not have permission to perform this action")
)

// ── User ───────────────────────────────────────────────────────────────────────
var (
	ErrUserNotFound         = NotFound("user.not_found", "User not found")
	ErrUsernameAlreadyTaken = Conflict("user.username_taken", "Username is already in use")
	ErrUserDeactivated      = Forbidden("user.deactivated", "This account has been deactivated")
)

// ── Validation ────────────────────────────────────────────────────────────────
var (
	ErrValidationFailed = NewAppError(http.StatusUnprocessableEntity, "validation.failed", "One or more fields are invalid")
	ErrInvalidUUID      = BadRequest("validation.invalid_uuid", "The provided ID is not a valid UUID")
	ErrInvalidPageParam = BadRequest("validation.invalid_page", "Page must be a positive integer")
)

// ── Resource ──────────────────────────────────────────────────────────────────
var (
	ErrResourceNotFound = NotFound("resource.not_found", "The requested resource was not found")
	ErrResourceConflict = Conflict("resource.conflict", "This resource already exists")
)

// ── System ────────────────────────────────────────────────────────────────────
var (
	ErrDatabaseUnavailable = ServiceUnavailable("system.db_unavailable", "Database is temporarily unavailable")
	ErrCacheUnavailable    = ServiceUnavailable("system.cache_unavailable", "Cache is temporarily unavailable")
	ErrInternal            = Internal("system.internal", "An unexpected error occurred")
)

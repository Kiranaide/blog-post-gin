package response

import (
	"blog-post-gin/internal/helper"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	responseDataKey    = "response_data"
	responseMessageKey = "response_message"
	responseStatusKey  = "response_status"
	requestIDKey       = "request_id"
)

func OK[T any](c *gin.Context, message string, data T) {
	send(c, http.StatusOK, message, data)
}

func Created[T any](c *gin.Context, message string, data T) {
	send(c, http.StatusCreated, message, data)
}

func Accepted[T any](c *gin.Context, message string, data T) {
	send(c, http.StatusAccepted, message, data)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func send[T any](c *gin.Context, status int, message string, data T) {
	resp := BaseResponse[T]{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: requestID(c),
		Timestamp: time.Now().UTC(),
	}
	c.JSON(status, resp)
}

func SendError(c *gin.Context, err error) {
	ae, ok := helper.AsError(err)
	if !ok {
		ae = helper.ErrInternal.WithCause(err)
	}

	message := ae.Message

	if ae.HTTPStatus >= http.StatusInternalServerError {
		message = "An unexpected error occurred. Please try again later."
	}

	c.AbortWithStatusJSON(ae.HTTPStatus, ErrorResponseDTO{
		Success: false,
		Error: ErrorBody{
			Code:    ae.Code,
			Message: message,
			Details: ae.Details,
		},
		RequestID: requestID(c),
		Timestamp: time.Now().UTC(),
	})
}

func SendValidationError(c *gin.Context, details any) {
	ae := helper.ErrValidationFailed.WithDetails(details)
	SendError(c, ae)
}

func BindAndValidate(c *gin.Context, obj any) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			details := buildValidationDetails(ve)
			SendValidationError(c, details)
			return err
		}
		ae := helper.BadRequest("validation.parse_error", "Request body is malformed or contains invalid types")
		SendError(c, ae)
		return err
	}
	return nil
}

func buildValidationDetails(ve validator.ValidationErrors) map[string]string {
	out := make(map[string]string, len(ve))
	for _, fe := range ve {
		field := toSnakeCase(fe.Field())
		out[field] = validationMessage(fe)
	}
	return out
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Must be at least " + fe.Param() + " characters"
	case "max":
		return "Must be at most " + fe.Param() + " characters"
	case "len":
		return "Must be exactly " + fe.Param() + " characters"
	case "gte":
		return "Must be greater than or equal to " + fe.Param()
	case "lte":
		return "Must be less than or equal to " + fe.Param()
	case "oneof":
		return "Must be one of: " + strings.ReplaceAll(fe.Param(), " ", ", ")
	case "uuid":
		return "Must be a valid UUID"
	case "url":
		return "Must be a valid URL"
	default:
		return "Failed validation: " + fe.Tag()
	}
}

func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + 32)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func requestID(c *gin.Context) string {
	if id, exists := c.Get(requestIDKey); exists {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return c.GetHeader("X-Request-ID")
}

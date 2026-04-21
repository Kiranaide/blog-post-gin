package response

import "time"

type T = any

type BaseResponse[T any] struct {
	Success   bool      `json:"success"`
	Data      T         `json:"data,omitempty"`
	Message   string    `json:"message,omitempty"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

func ToSuccessResponse[T any](data T, message string) BaseResponse[T] {
	return BaseResponse[T]{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
}

type ErrorResponseDTO struct {
	Success   bool      `json:"success"`
	Error     ErrorBody `json:"error"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func ToErrorResponse(code, message string, details any) ErrorResponseDTO {
	return ErrorResponseDTO{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}
}

type EmptyResponse *struct{}

func ToEmptyResponse() EmptyResponse {
	return nil
}

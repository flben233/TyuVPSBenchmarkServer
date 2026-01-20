package common

const (
	SuccessCode       = 0
	InternalErrorCode = -1
	BadRequestCode    = -2
	ForbiddenCode     = -3
)

// APIResponse is the unified response structure for all API endpoints
type APIResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

// PaginatedResponse is the response structure for paginated data
type PaginatedResponse[T any] struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	Data     T      `json:"data,omitempty"`
	Total    int64  `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// Success creates a successful response with data
func Success[T any](data T) APIResponse[T] {
	return APIResponse[T]{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// SuccessWithMessage creates a successful response with a custom message
func SuccessWithMessage[T any](message string, data T) APIResponse[T] {
	return APIResponse[T]{
		Code:    0,
		Message: message,
		Data:    data,
	}
}

// Error creates an error response
func Error(code int, message string) APIResponse[any] {
	return APIResponse[any]{
		Code:    code,
		Message: message,
	}
}

// SuccessPaginated creates a successful paginated response
func SuccessPaginated[T any](data T, total int64, page, pageSize int) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Code:     0,
		Message:  "success",
		Data:     data,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

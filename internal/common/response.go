package common

const (
	SuccessCode       = 0
	InternalErrorCode = -1
	BadRequestCode    = -2
	ForbiddenCode     = -3
)

// APIResponse is the unified response structure for all API endpoints
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse is the response structure for paginated data
type PaginatedResponse struct {
	Code     int         `json:"code"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// Success creates a successful response with data
func Success(data interface{}) APIResponse {
	return APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// SuccessWithMessage creates a successful response with a custom message
func SuccessWithMessage(message string, data interface{}) APIResponse {
	return APIResponse{
		Code:    0,
		Message: message,
		Data:    data,
	}
}

// Error creates an error response
func Error(code int, message string) APIResponse {
	return APIResponse{
		Code:    code,
		Message: message,
	}
}

// SuccessPaginated creates a successful paginated response
func SuccessPaginated(data interface{}, total int64, page, pageSize int) PaginatedResponse {
	return PaginatedResponse{
		Code:     0,
		Message:  "success",
		Data:     data,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

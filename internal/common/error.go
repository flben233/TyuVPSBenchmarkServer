package common

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type LimitExceededError struct {
	Message string
}

func (e *LimitExceededError) Error() string {
	return e.Message
}

type InvalidParamError struct {
	Message string
}

func (e *InvalidParamError) Error() string {
	return e.Message
}

type RateLimitExceededError struct {
	Message string
}

func (e *RateLimitExceededError) Error() string {
	return e.Message
}

func DefaultErrorHandler(ctx *gin.Context, err error) {
	var limitErr *LimitExceededError
	var invalidParamErr *InvalidParamError
	var rateLimitErr *RateLimitExceededError

	switch {
	case errors.As(err, &limitErr):
		ctx.JSON(http.StatusOK, Error(LimitExceededCode, limitErr.Error()))
	case errors.As(err, &invalidParamErr):
		ctx.JSON(http.StatusOK, Error(InvalidParamCode, invalidParamErr.Error()))
	case errors.As(err, &rateLimitErr):
		ctx.JSON(http.StatusOK, Error(RateLimitExceededCode, rateLimitErr.Error()))
	default:
		ctx.JSON(http.StatusInternalServerError, Error(InternalErrorCode, err.Error()))
	}
}

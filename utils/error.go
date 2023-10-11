package utils

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AppError struct {
	StatusCode int    `json:"-"`
	Reason     string `json:"error"`
	Message    any    `json:"message,omitempty"`
}

func ReturnError(ctx *gin.Context, err error) {
	var appError *AppError
	switch {
	case errors.As(err, &appError):
		ctx.JSON(appError.StatusCode, appError)
	default:
		ctx.JSON(http.StatusInternalServerError, err)
	}
}

func (e *AppError) Error() string {
	return e.Reason
}

func NewAppError(statusCode int, err string, message any) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Reason:     err,
		Message:    message,
	}
}

func BadRequest(message any) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Reason:     "BadRequestError",
		Message:    message,
	}
}

func PermissionDenied(message any) *AppError {
	return &AppError{
		StatusCode: http.StatusForbidden,
		Reason:     "PermissionDenied",
		Message:    message,
	}
}

func NotFound(message any) *AppError {
	return &AppError{
		StatusCode: http.StatusNotFound,
		Reason:     "NotFound",
		Message:    message,
	}
}

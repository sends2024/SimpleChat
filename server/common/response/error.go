package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AppError struct {
	Code int
	Msg  string
}

func (e *AppError) Error() string {
	return e.Msg
}

func NewAppError(code int, msg string) *AppError {
	return &AppError{Code: code, Msg: msg}
}

func HandleServiceError(c *gin.Context, err error) {
	if appErr, ok := err.(*AppError); ok {
		Fail(c, appErr.Code, appErr.Msg)
	} else {
		Fail(c, http.StatusInternalServerError, "Internal server error")
	}
}

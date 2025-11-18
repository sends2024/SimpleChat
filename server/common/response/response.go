package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Msg:  "ok",
		Data: data,
	})
}

func Fail(c *gin.Context, status int, msg string) {
	c.JSON(status, Response{
		Msg:  msg,
		Data: nil,
	})
}

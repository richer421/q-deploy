package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

func Fail(c *gin.Context, err error) {
	c.JSON(http.StatusOK, Response{Code: -1, Message: err.Error()})
}

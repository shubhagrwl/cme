package controller

import (
	"cme/internal/app/service/dto/response"

	"github.com/gin-gonic/gin"
)

func RespondWithError(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, response.ResponseV2{Success: false, Message: message})
}

func RespondWithSuccess(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, response.ResponseV2{Success: true, Message: message, Data: data})
}

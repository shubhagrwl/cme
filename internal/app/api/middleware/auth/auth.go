package auth

import (
	"cme/internal/app/api/middleware/jwt"
	"cme/internal/app/constants"
	"cme/internal/app/controller"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// authentication is a middleware that verify JWT token headers
func Authentication(jwt jwt.IJwtService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := getHeaderToken(ctx)
		if err != nil {
			controller.RespondWithError(ctx, http.StatusUnauthorized, err.Error())
			return
		}

		claims, valid := jwt.VerifyToken(ctx, token)
		if !valid {
			controller.RespondWithError(ctx, http.StatusUnauthorized, "unauthorized")
			return
		}
		ctx.Set(constants.CTK_CLAIM_KEY.String(), claims)
		ctx.Next()
	}
}

func getHeaderToken(ctx *gin.Context) (string, error) {
	header := string(ctx.GetHeader(constants.AUTHORIZATION))
	return extractToken(header)
}

func extractToken(header string) (string, error) {
	if strings.HasPrefix(header, constants.BEARER) {
		return header[len(constants.BEARER):], nil
	}
	return "", errors.New("token not found")
}

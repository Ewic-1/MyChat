package controller

import (
	"net/http"
	"strings"

	"mychat_server/internal/dto/request"
	"mychat_server/internal/dto/respond"
	"mychat_server/internal/middleware"
	"mychat_server/pkg/utils/jwtutil"

	"github.com/gin-gonic/gin"
)

// RefreshToken 使用 refresh token 换取新的 token 对。
func RefreshToken(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request",
		})
		return
	}

	pair, err := jwtutil.RotateTokenPair(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "invalid or expired refresh token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "refresh success",
		"data": respond.TokenRespond{
			Token:            pair.AccessToken,
			RefreshToken:     pair.RefreshToken,
			AccessExpiresAt:  pair.AccessExpiresAt.Unix(),
			RefreshExpiresAt: pair.RefreshExpiresAt.Unix(),
		},
	})
}

// Logout 撤销当前用户在服务端记录的全部 refresh token。
func Logout(c *gin.Context) {
	uuidValue, ok := c.Get(middleware.ContextUserUUIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "unauthorized",
		})
		return
	}

	uuid, ok := uuidValue.(string)
	if !ok || strings.TrimSpace(uuid) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "unauthorized",
		})
		return
	}

	jwtutil.RevokeRefreshTokensByUUID(uuid)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "logout success",
	})
}

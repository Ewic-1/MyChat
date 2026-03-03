package middleware

import (
	"net/http"
	"strings"

	"mychat_server/pkg/utils/jwtutil"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserUUIDKey = "current_uuid"
	ContextJWTClaims   = "jwt_claims"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "missing authorization header",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		claims, err := jwtutil.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		c.Set(ContextUserUUIDKey, claims.Uuid)
		c.Set(ContextJWTClaims, claims)
		c.Next()
	}
}

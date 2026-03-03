package ssl

import "github.com/gin-gonic/gin"

// TlsHandler currently acts as a pass-through middleware.
// Keep the signature stable so future TLS redirect logic can be added safely.
func TlsHandler(host string, port int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

package token

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gonzaloccnc/marketplace-go/pkg/httpx"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
	// contextClaimsKey is the gin context key under which the authenticated
	// user's claims are stored for downstream handlers.
	contextClaimsKey = "userClaims"
)

// Middleware validates the Bearer JWT on incoming requests. On success it stores
// the parsed claims in the gin context; otherwise it aborts with 401.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(authorizationHeader)
		if header == "" || !strings.HasPrefix(header, bearerPrefix) {
			httpx.WriteError(c, http.StatusUnauthorized, "missing or malformed authorization header")
			c.Abort()
			return
		}

		raw := strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))

		claims, err := Validate(raw)
		if err != nil {
			httpx.WriteError(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(contextClaimsKey, claims)
		c.Next()
	}
}

// ClaimsFromContext returns the authenticated user's claims previously stored by
// Middleware. The boolean is false when the request was not authenticated.
func ClaimsFromContext(c *gin.Context) (*Claims, bool) {
	value, exists := c.Get(contextClaimsKey)
	if !exists {
		return nil, false
	}

	claims, ok := value.(*Claims)
	return claims, ok
}

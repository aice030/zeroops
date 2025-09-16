package middleware

import (
	"github.com/fox-gonic/fox"
)

// Authentication is a placeholder global middleware. It currently allows all requests.
// Per-alerting webhook uses its own path-scoped auth.
func Authentication(c *fox.Context) {
	c.Next()
}

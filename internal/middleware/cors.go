package middleware

import (
	"net/http"

	"github.com/fox-gonic/fox"
)

// CORS 跨域中间件
func CORS() fox.HandlerFunc {
	return func(c *fox.Context) {
		origin := c.Request.Header.Get("Origin")

		// 设置CORS头
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 设置实际的Origin（如果有的话）
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Next()
	}
}

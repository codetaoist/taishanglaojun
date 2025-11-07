package middleware

import (
	"net/http"
	"strings"

	"github.com/codetaoist/taishanglaojun/api/internal/config"
	"github.com/gin-gonic/gin"
)

func CORS(cfg config.Config) gin.HandlerFunc {
	allowed := cfg.AllowedOrigins
	allowAll := false
	for _, o := range allowed {
		if o == "*" {
			allowAll = true
			break
		}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			// Non-CORS request
			c.Next()
			return
		}
		allowedOrigin := ""
		if allowAll {
			allowedOrigin = origin
		} else {
			for _, o := range allowed {
				if strings.EqualFold(o, origin) {
					allowedOrigin = origin
					break
				}
			}
		}
		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", "GET,POST,DELETE,PUT,PATCH,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,Accept,Origin,User-Agent,Cache-Control,X-Requested-With,Referer,"+cfg.TraceHeader)
			c.Header("Access-Control-Expose-Headers", cfg.TraceHeader)
			c.Header("Access-Control-Allow-Credentials", "false")
			if c.Request.Method == http.MethodOptions {
				c.Status(http.StatusNoContent)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
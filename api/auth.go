package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bugfan/ang-admin/service"
	"github.com/gin-gonic/gin"
)

func extractToken(c *gin.Context) string {
	// 1. Prioritize Authorization header (Bearer <token>) sent by frontend axios interceptor
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token != "" {
			return token
		}
	} else if authHeader != "" {
		return authHeader
	}

	if xToken := c.GetHeader("X-Token"); xToken != "" {
		return xToken
	}

	// 2. Fall back to Cookie (check JSON format or raw string)
	for _, cookieName := range []string{"authorized-token", "_ty", "token"} {
		if cookieVal, err := c.Cookie(cookieName); err == nil && cookieVal != "" {
			if strings.HasPrefix(cookieVal, "{") {
				var cookieObj struct {
					AccessToken string `json:"accessToken"`
				}
				if err := json.Unmarshal([]byte(cookieVal), &cookieObj); err == nil && cookieObj.AccessToken != "" {
					return cookieObj.AccessToken
				}
			}
			return cookieVal
		}
	}

	return ""
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未登录或凭证过期，请先登录",
			})
			return
		}

		claims, err := service.ParseToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的登录凭证或登录已过期",
			})
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}

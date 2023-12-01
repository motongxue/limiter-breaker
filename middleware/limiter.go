package middleware

import (
	"github.com/gin-gonic/gin"
	"limiter-breaker/limiter"
	"net/http"
)

func Limiter(l *limiter.Limiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !l.Allow() {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": "当前可用令牌数为0，请稍后再试",
			})
			ctx.Abort()
		}
		ctx.Next()
	}
}

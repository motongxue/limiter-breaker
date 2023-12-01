package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"limiter-breaker/breaker"
	"limiter-breaker/limiter"
	"limiter-breaker/middleware"
	"net/http"
	"time"
)

func main() {
	r := gin.Default()
	// 限流器
	r.GET("/limiter", middleware.Limiter(limiter.NewLimiter(time.Second, 4)), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "limiter",
		})
	})

	//断路器
	b := breaker.NewBreaker(4, 4, 2, time.Second*15)
	r.GET("/breaker", func(c *gin.Context) {
		err := b.Exec(func() error {
			value, _ := c.GetQuery("value")
			if value == "a" {
				return errors.New("value 为 a 返回错误")
			}
			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "breaker",
		})
	})
	r.Run(":8080")
}

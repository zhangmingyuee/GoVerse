package middleware

import (
	"bluebell/controllers"
	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
	"time"
)

// RateLimitMiddleware 创建一个基于令牌桶的限流中间件
// @param fillInterval time.Duration - 令牌生成的间隔时间（如 `time.Second` 表示每秒生成一个令牌）
// @param cap int64 - 令牌桶的容量（最大存储多少个令牌）
// @return func(c *gin.Context) - 返回一个 Gin 中间件函数
func RateLimitMiddleware(fillInterval time.Duration, cap int64) func(c *gin.Context) {
	// 创建一个令牌桶，每 `fillInterval` 生成一个令牌，最多存 `cap` 个
	bucket := ratelimit.NewBucket(fillInterval, cap)

	// 返回 Gin 中间件函数
	return func(c *gin.Context) {
		// 从令牌桶中尝试取出 1 个令牌
		if bucket.TakeAvailable(1) < 1 {
			// 如果没有令牌可用，返回限流响应
			controllers.ResponseError(c, controllers.CodeRateLimit)
			//c.String(http.StatusOK, "rate limit...") // 可改为 http.StatusTooManyRequests (429)
			c.Abort() // 终止后续处理
			return
		}

		// 继续处理下一个中间件或最终的请求处理函数
		c.Next()
	}
}

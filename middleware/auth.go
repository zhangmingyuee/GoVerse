package middleware

import (
	"bluebell/controllers"
	"bluebell/dao/mysql"
	"bluebell/pkg/jwt"
	"github.com/gin-gonic/gin"
	"strings"
)

// JWTAuthMiddleware 基于JWT的认证中间件
func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 中间件首先从 HTTP 请求头中获取 Authorization 字段的值，它包含了客户端提交的 JWT
		// 客户端携带Token有三种方式 1.放在请求头 2.放在请求体 3.放在URI
		// 这里假设Token放在Header的Authorization中，并使用Bearer开头
		// 这里的具体实现方式要依据你的实际业务情况决定
		// Authorization: Bearer xxxxxxx.xxx.xxx

		// 检查获取到的 Authorization 头是否为空，以及格式是否正确。
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			controllers.ResponseError(c, controllers.CodeNeedLogin)
			c.Abort()
			return
		}
		// 按空格分割
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			controllers.ResponseError(c, controllers.CodeInvalidToken)
			c.Abort()
			return
		}
		// 🔴 新增：检查 Token 是否已被拉黑
		token := parts[1]
		c.Set(controllers.CtxTokenKey, token)
		// 检查 Token 是否在黑名单中
		if mysql.IsTokenBlacklisted(token) {
			controllers.ResponseError(c, controllers.CodeInvalidToken)
			c.Abort()
			return
		}
		// token是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
		mc, err := jwt.ParseToken(token)
		if err != nil {
			controllers.ResponseError(c, controllers.CodeInvalidToken)
			c.Abort()
			return
		}
		// 将当前请求的userid信息保存到请求的上下文c上
		c.Set(controllers.CtxUserIDKey, mc.Userid)
		c.Next() // 后续的处理函数可以用过c.Get(controllers.CtxUserIDKey)来获取当前请求的用户信息
	}
}

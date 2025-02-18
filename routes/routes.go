package routes

import (
	"bluebell/controllers"
	_ "bluebell/docs" // 千万不要忘了导入把你上一步生成的docs
	"bluebell/logger"
	"bluebell/middleware"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
	"time"
)

func Setup(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode) // gin设置成发布模式
	}

	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true), middleware.RateLimitMiddleware(time.Second*2, 5))

	// 注册路由组
	v1 := r.Group("/api/v1")

	// 登录业务路由
	v1.GET("/login", controllers.LogInHandler)

	// 注册业务路由
	v1.POST("/signup", controllers.SignUpHandler)

	// 生成验证码 API
	v1.POST("/gen-otp", controllers.GenerateOTPHandler)

	// 注册登录认证中间件
	v1.Use(middleware.JWTAuthMiddleware()) // JWTAuthMiddleware() 应用登录认证的中间件

	{
		// 注销
		v1.POST("/logout", controllers.LogOutHandler)

		// 查询个人信息
		v1.GET("/user", controllers.GetUserInfoHandler)

		// 修改个人信息
		v1.PATCH("/user", controllers.ModifyUserInfoHandler)

		// 修改密码 (还有问题)
		v1.PATCH("/password", controllers.ModifyPasswordHandler)

		// 根据用户名获取单个用户信息
		v1.GET("/user/:name", controllers.GetUserByNameHandler)

		// 根据用户名获取相关用户群
		v1.GET("/users", controllers.GetUsersByNameHandler)

		// 获取用户的全部帖子
		v1.GET("/user/posts/:id", controllers.GetUserPostHandler)

		// 获取全部社区
		v1.GET("/community", controllers.GetCommunityHandler)

		// 创建社区
		v1.POST("/community", controllers.CreateCommunityHandler)

		// 根据社区id/name获取社区详情
		v1.GET("/community/detail", controllers.CommunityDetailHandler)

		// 创建帖子
		v1.POST("/post", controllers.CreatePostHandler)

		// 根据帖子id获取帖子
		v1.GET("/post/:id", controllers.GetPostDetailHandler)

		// 根据时间或分数或获取帖子列表(可以按照社区分区)
		v1.GET("/post", controllers.GetPostListHandler)

		// 投票
		v1.POST("/vote", controllers.PostVoteController)

		// 创建评论
		v1.POST("/comment", controllers.CreateCommentConntroller)

		// 查看顶级评论
		v1.GET("/comment/:post_id", controllers.GetCommentConntroller)

		// 查看评论的子评论
		v1.GET("/comment/child/:parent_id", controllers.GetChildCommentsController)

		// 删除评论
		v1.DELETE("/comment", controllers.DeleteCommentController)

		// 置顶评论路由，例如 POST /comment/pin/:comment_id
		v1.POST("/comment/pin/:comment_id", controllers.PinCommentController)

		// 取消置顶评论路由，例如 POST /comment/pin/:comment_id
		v1.DELETE("/comment/pin/:comment_id", controllers.UnpinCommentController)

	}

	r.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))

	pprof.Register(r) // 注册pprof相关路由

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	return r
}

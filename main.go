package main

import (
	"bluebell/controllers"
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/logger"
	"bluebell/logic"
	"bluebell/pkg/snowflake"
	"bluebell/routes"
	"bluebell/settings"
	"context"
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title 第一个Go语言项目--Bluebell
// @version 1.0
// @description 这是一个仿rabbit的新闻分享的小项目，目前完成了后端的开发
// @termsOfService http://swagger.io/terms/

// @contact.name 沧叶
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host 这里写接口服务的host
// @BasePath 这里写base path
func main() {
	// 1.加载配置
	if err := settings.Init(); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}

	// 2.初始化日志
	if err := logger.Init(viper.GetString("app.mode")); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	zap.L().Debug("Logger init success...")
	defer zap.L().Sync()

	// 3.初始化Mysql
	defer mysql.Close()
	if err := mysql.Init(); err != nil {
		zap.L().Error("init mysql failed, err:%v\n", zap.Error(err))
		return
	}

	// 4.初始化Redis
	defer redis.Close()
	if err := redis.Init(); err != nil {
		zap.L().Error("init redis failed, err:%v\n", zap.Error(err))
		return
	}

	//// 5. 初始化postgresql
	//defer postgresql.Close()
	//if err := postgresql.Init(); err != nil {
	//	zap.L().Error("init postgresql failed, err:%v\n", zap.Error(err))
	//	return
	//}

	// 初始化注册id
	if err := snowflake.Init(viper.GetString("app.start_time"), viper.GetInt64("app.machine_id")); err != nil {
		zap.L().Error("init snowflake failed, err:%v\n", zap.Error(err))
		return
	}

	// 初始化gin框架内置的校验器使用的翻译器
	if err := controllers.InitTrans("zh"); err != nil {
		zap.L().Error("init translator failed, err:%v\n", zap.Error(err))
		return
	}

	logic.StartCronJob()

	// 5.注册路由
	r := routes.Setup(viper.GetString("app.mode"))
	//r.Run("0.0.0.0:8080") // 允许任何人访问

	// 6.启动服务（优雅关机）
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.GetInt("app.port")),
		Handler: r,
	}

	go func() {
		// 开启一个goroutine启动服务
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("listen: %s\n", zap.Error(err))
		}
	}()

	// 等待中断信号来优雅地关闭服务器，为关闭服务器操作设置一个5秒的超时
	quit := make(chan os.Signal, 1) // 创建一个接收信号的通道
	// kill 默认会发送 syscall.SIGTERM 信号
	// kill -2 发送 syscall.SIGINT 信号，我们常用的Ctrl+C就是触发系统SIGINT信号
	// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它
	// signal.Notify把收到的 syscall.SIGINT或syscall.SIGTERM 信号转发给quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 此处不会阻塞
	<-quit                                               // 阻塞在此，当接收到上述两种信号时才会往下执行
	zap.L().Info("Shutdown Server ...")
	// 创建一个5秒超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 5秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过5秒就超时退出
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server Shutdown", zap.Error(err))
	}

	zap.L().Info("Server exiting")

}

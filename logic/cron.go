package logic

import (
	"bluebell/dao/redis"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// 启动定时任务
func StartCronJob() {
	c := cron.New()
	_, err := c.AddFunc("@every 10m", syncHotScoresToMySQL) // 每 10 分钟同步增量数据
	if err != nil {
		zap.L().Error("定时任务创建失败", zap.Error(err))
	}

	_, err = c.AddFunc("@daily", redis.CleanOldRedisData) // 每天清理 Redis 过期数据
	if err != nil {
		zap.L().Error("定时任务创建失败", zap.Error(err))
	}

	c.Start()
}

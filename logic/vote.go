package logic

import (
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/models"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"math"
	"strconv"
	"time"
)

// 投票功能

// 本项目使用简化版的投票分数实战
// 投一票加432分 86400/200 --> 200张赞成票可以给帖子续一天 --> <<redis实战>>

/*
	投票的几种情况：
	direction=1时，
		之前没有投过票，现在投赞成票 --> 更新分数和投票纪录
		之前投反对票，现在改投反对票 --> 更新分数和投票纪录
	direction=0时，
		之前投过赞成票，现在要取消投票 --> 更新分数和投票纪录
		之前透过反对票，现在取消投票 --> 更新分数和投票纪录
	direction=-1时，
		之前没有投过票，现在投反对票 --> 更新分数和投票纪录
		之前投赞成票，现在改投反对票 --> 更新分数和投票纪录

	投票的限制：
	每个帖子自发表之日起，仅限一个星期之内允许用户投票。
		1. 到期之后redis中保存的赞成票或反对票数存储到mysql中。
		2. 到期之后删除那个 KeyPostVotedZSetPreix
*/

// 计算 Reddit 热度
func computeRedditHotScore(ups, downs, postTime int64) float64 {
	score := math.Max(float64(ups-downs), 1)
	order := math.Log10(score)
	//seconds := float64(postTime - 1134028003) // Reddit 基准时间
	seconds := float64(postTime - 1735689600) // bluebell发布的基准时间（2025-01-01）
	return order + seconds/45000.0            // 约12.5小时热度加一
}

// VoteForPost 为帖子投票的函数
func VoteForPost(c *gin.Context, userID int64, p *models.ParamVoteData) error {
	zap.L().Debug("VoteForPost", zap.Int64("userID", userID), zap.Int64("postid", p.PostID), zap.Int8("direction", p.Direction))
	uidStr := strconv.FormatInt(userID, 10)
	pidStr := strconv.FormatInt(p.PostID, 10)
	// 更新用户为该帖子投票结果
	if err := redis.VoteForPost(c, uidStr, pidStr, float64(p.Direction)); err != nil {
		zap.L().Error("VoteForPost", zap.Error(err))
		return err
	}
	// 更新热度 --> 得到赞成票、反对票以及发帖时间
	approve, err := redis.GetPostVoteByID(c, strconv.FormatInt(p.PostID, 10))
	if err != nil {
		zap.L().Error("GetPostVoteByID", zap.Error(err))
		return err
	}
	against, err := redis.GetPostVoteAgainstByID(c, strconv.FormatInt(p.PostID, 10))
	if err != nil {
		zap.L().Error("GetPostVoteAgainstData", zap.Error(err))
		return err
	}
	createTimeStamp, err := GetPostCreateTimeCached(c, p.PostID)
	//fmt.Println("createTimeStamp:", createTimeStamp)
	if err != nil {
		zap.L().Error("GetPostCreateTimeCached", zap.Error(err))
		return err
	}
	// 更新热度
	score := computeRedditHotScore(approve, against, createTimeStamp)
	fmt.Println("score:", score)
	return redis.UpdateScore(c, strconv.FormatInt(p.PostID, 10), score)

}

// 获取上次同步的时间戳
func getLastSyncTime(c context.Context) (int64, error) {
	lastSyncTimeStr, err := redis.GetLastSyncTime(c)
	if err != nil {
		zap.L().Error("GetLastSyncTime failed", zap.Error(err))
		return time.Now().Unix() - 600, err // 默认同步 10 分钟内的数据
	}
	lastSyncTime, err := strconv.ParseInt(lastSyncTimeStr, 10, 64)
	if err != nil {
		zap.L().Error("strconv.ParseInt(lastSyncTimeStr) failed", zap.Error(err))
		return time.Now().Unix() - 600, err
	}
	return lastSyncTime, err

}

// 设置新的 last_sync_time
func setLastSyncTime(c context.Context, timestamp int64) error {
	if err := redis.SetLastSyncTime(c, timestamp); err != nil {
		zap.L().Error("redis.SetLastSyncTime failed", zap.Error(err))
		return err
	}
	return nil
}

// 增量同步 Redis 热度数据到 MySQL
func syncHotScoresToMySQL() {
	fmt.Println("增量同步开始:", time.Now())
	c := context.Background()
	// 获取上次同步的时间
	lastSyncTime, err := getLastSyncTime(c)
	currentTime := time.Now().Unix()
	postUpdateScores, err := redis.GetScoreLimitTime(c, lastSyncTime, currentTime)
	if err != nil {
		zap.L().Error("redis.GetScoreLimitTime failed", zap.Error(err))
		return
	}

	if len(postUpdateScores) == 0 {
		zap.L().Info("没有需要同步的数据")
		setLastSyncTime(c, currentTime)
		return
	}

	// 开始MySQL事务，并将数据存入mysql数据库
	if err := mysql.BatchInsertPostHotScores(postUpdateScores); err != nil {
		zap.L().Error("BatchInsertPostHotScores failed", zap.Error(err))
	}

	// 更新 last_sync_time
	err = setLastSyncTime(c, currentTime)
	if err != nil {
		zap.L().Error("setLastSyncTime failed", zap.Error(err))
		return
	}
	zap.L().Info("同步数据完成", zap.Int("updateNum", len(postUpdateScores)))
	return
}

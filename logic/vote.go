package logic

import (
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/models"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	rd "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"math"
	"strconv"
	"time"
)

// 投票功能
/*
	1. 用户点赞时，直接更新 Redis 热度（不实时写 MySQL）
	2. 用户查询时，直接从 Redis 读取，无需计算，速度快
	3. 定期任务同步 Redis → MySQL，确保最终一致性
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
	// 更新用户和帖子行为表
	// 更新帖子用户行为信息
	//var behavior *models.UserPostBehavior
	behavior, err := mysql.CheckBehavior(userID, p.PostID)
	if err == gorm.ErrRecordNotFound {
		// 没有找到记录, 创建记录
		zap.L().Info("不存在该用户-帖子行为", zap.Int64("postID", p.PostID), zap.Int64("userID", userID))
		if err := mysql.CreateBehavior(userID, p.PostID, mysql.BehaviorLike); err != nil {
			zap.L().Error("mysql.CreateBehavior failed", zap.Error(err))
			return err
		}
		return nil // 返回 nil 表示没有记录

	} else if err != nil {
		zap.L().Error("mysql.CheckBehavior failed", zap.Error(err))
		return err
	}

	// 行为存在，查看是否需要更新行为
	if behavior.Like == 0 {
		if err := mysql.UpdateBehavior(userID, p.PostID, mysql.BehaviorLike); err != nil {
			zap.L().Error("mysql.UpdateBehavior failed", zap.Error(err))
			return err
		}
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
	// 计算热度
	score := computeRedditHotScore(approve, against, createTimeStamp)
	//fmt.Println("score:", score)
	err = redis.UpdateScore(c, strconv.FormatInt(p.PostID, 10), score)
	if err != nil {
		zap.L().Error("UpdateScore", zap.Error(err))
	}
	// 在redis中更新投票的的最新时间
	return redis.UpdateVoteTime(c, strconv.FormatInt(p.PostID, 10), float64(time.Now().Unix()))
}

// 获取上次同步的时间戳
func getLastSyncTime(c context.Context, syncKey string) (int64, error) {
	lastSyncTimeStr, err := redis.GetLastSyncTime(c, syncKey)
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
func setLastSyncTime(c context.Context, syncKey string, timestamp int64) error {
	if err := redis.SetLastSyncTime(c, syncKey, timestamp); err != nil {
		zap.L().Error("redis.SetLastSyncTime failed", zap.Error(err))
		return err
	}
	return nil
}

// 增量同步 Redis 热度数据到 MySQL
func SyncHotsDLikesToMySQL() {
	fmt.Println("热度增量同步开始:", time.Now())
	c := context.Background()
	// 获取上次同步的时间
	lastSyncTime, err := getLastSyncTime(c, redis.LastSyncTimeHotDLikesKey)
	currentTime := time.Now().Unix()
	postUpdateScores, err := redis.GetTimeAndScore(c, lastSyncTime, currentTime)
	if err != nil {
		zap.L().Error("redis.GetScoreLimitTime failed", zap.Error(err))
		return
	}
	// 没有新增点赞点踩数据
	if len(postUpdateScores) == 0 {
		zap.L().Info("没有需要同步的热度数据", zap.Int64("lastSyncTime", lastSyncTime), zap.Int64("currentTime", currentTime))
		setLastSyncTime(c, redis.LastSyncTimeHotDLikesKey, currentTime)
		return
	}
	// 增量更新热度到mysql
	if err := syncHots(c, postUpdateScores); err != nil {
		zap.L().Error("syncHots failed", zap.Error(err))
		return
	}
	// 增量更新热度到redis
	if err := syncLikes(c, postUpdateScores); err != nil {
		zap.L().Error("syncLikes failed", zap.Error(err))
		return
	}
	// 增量更新点踩到redis
	if err := syncDislikes(c, postUpdateScores); err != nil {
		zap.L().Error("syncDislikes failed", zap.Error(err))
		return
	}

	// 更新 last_hot_sync_time
	err = setLastSyncTime(c, redis.LastSyncTimeHotDLikesKey, currentTime)
	if err != nil {
		zap.L().Error("setLastSyncTime failed", zap.Error(err))
		return
	}
	zap.L().Info("同步数据完成", zap.Int("updateNum", len(postUpdateScores)))
	return
}

// 增量更新热度
func syncHots(c context.Context, postUpdateScores []rd.Z) error {
	// 根据postUpdateScores中的post_id从redis中获取热度
	idhot := redis.GetHotsByIds(c, postUpdateScores)

	// 开始MySQL事务，并将数据存入mysql数据库
	if err := mysql.BatchInsertPostHotScores(idhot); err != nil {
		zap.L().Error("BatchInsertPostHotScores failed", zap.Error(err))
		return err
	}
	return nil
}

// 增量更新点赞数
func syncLikes(c context.Context, postUpdateScores []rd.Z) error {
	// 根据postUpdateScores中的post_id从redis中获取点赞数
	idlikes := redis.GetLikesByIds(c, postUpdateScores)

	// 开始MySQL事务，并将数据存入mysql数据库
	if err := mysql.BatchInsertLikes(idlikes); err != nil {
		zap.L().Error("BatchInsertPostHotScores failed", zap.Error(err))
		return err
	}
	return nil
}

// 增量更新点踩数
func syncDislikes(c context.Context, postUpdateScores []rd.Z) error {
	// 根据postUpdateScores中的post_id从redis中获取点赞数
	iddislikes := redis.GetDisLikesByIds(c, postUpdateScores)

	// 开始MySQL事务，并将数据存入mysql数据库
	if err := mysql.BatchInsertDisLikes(iddislikes); err != nil {
		zap.L().Error("BatchInsertPostHotScores failed", zap.Error(err))
		return err
	}
	return nil
}

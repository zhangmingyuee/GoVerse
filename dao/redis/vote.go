package redis

import (
	"bluebell/models"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strconv"
	"time"
)

//const (
//	oneWeekInSecond = 7 * 24 * 3600
//	scorePerVote    = 432 // 每一票贡献的得分
//)
//
//var ErrVoteTimeExpire = errors.New("投票时间已过")
//var ErrVoteRepeat = errors.New("不允许重复投票")

// VoteForPost 用户为帖子投票，更新用户投票redis
func VoteForPost(c *gin.Context, userID, postID string, v float64) error {
	// 更新当前用户对当前帖子的投票
	if v == 0 {
		rdb.ZRem(c, getRedisKey(KeyPostVotedZSetPreix+postID), userID)
		return nil
	} else {
		// 查询当前用户对当前帖子之前的投票纪录
		pipline := rdb.TxPipeline()
		ov, err := rdb.ZScore(c, getRedisKey(KeyPostVotedZSetPreix+postID), userID).Result()
		if err == redis.Nil {
			ov = 0
		} else if err != nil {
			return err
		}
		// 不允许重复投票
		if ov == v {
			return nil
		}
		// 更新投票数
		pipline.ZAdd(c, getRedisKey(KeyPostVotedZSetPreix+postID), redis.Z{
			Score:  v,
			Member: userID,
		})

		_, err = pipline.Exec(c)
		return err
	}
}

// 获取赞同投票数
func GetPostVoteByID(c *gin.Context, postid string) (int64, error) {
	key := getRedisKey(KeyPostVotedZSetPreix + postid)
	return rdb.ZCount(c, key, "1", "1").Result()
}

// 获取反对投票数
func GetPostVoteAgainstByID(c *gin.Context, postid string) (int64, error) {
	key := getRedisKey(KeyPostVotedZSetPreix + postid)
	return rdb.ZCount(c, key, "-1", "-1").Result()
}

// GetPostVoteData 从redis获取全部帖子的投赞成票数量
func GetPostVoteData(c *gin.Context, ps []*models.Post) ([]int64, error) {
	if len(ps) == 0 {
		return nil, nil
	}
	// 使用pipeline 减少 Redis 请求的 RTT
	pipe := rdb.Pipeline()
	cmds := make([]*redis.IntCmd, len(ps))

	for i, p := range ps {
		id := strconv.FormatInt(p.ID, 10)
		key := getRedisKey(KeyPostVotedZSetPreix + id)
		cmds[i] = pipe.ZCount(c, key, "1", "1")
	}

	_, err := pipe.Exec(c)
	if err != nil {
		return nil, err
	}

	// 统计每个帖子的票数
	data := make([]int64, len(ps))
	for i, cmd := range cmds {
		vote, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		data[i] = vote
	}

	return data, nil
}

// GetPostVoteAgainstData 从redis获取全部帖子的投反对票数量
func GetPostVoteAgainstData(c *gin.Context, ps []*models.Post) ([]int64, error) {
	if len(ps) == 0 {
		return nil, nil
	}
	// 使用pipeline 减少 Redis 请求的 RTT
	pipe := rdb.Pipeline()
	cmds := make([]*redis.IntCmd, len(ps))

	for i, p := range ps {
		id := strconv.FormatInt(p.ID, 10)
		key := getRedisKey(KeyPostVotedZSetPreix + id)
		cmds[i] = pipe.ZCount(c, key, "-1", "-1")
	}

	_, err := pipe.Exec(c)
	if err != nil {
		return nil, err
	}

	// 统计每个帖子的票数
	data := make([]int64, len(ps))
	for i, cmd := range cmds {
		vote, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		data[i] = vote
	}

	return data, nil
}

// UpdateScore 更新帖子得分
func UpdateScore(c *gin.Context, postid string, score float64) error {
	return rdb.ZAdd(c, getRedisKey(KeyPostScoreZSet), redis.Z{
		Score: score, Member: postid}).Err()
}

// UpdateVoteTime 更新帖子投票的时间
func UpdateVoteTime(c *gin.Context, postid string, time float64) error {
	return rdb.ZAdd(c, getRedisKey(KeyPostUpdateTimeZSet), redis.Z{
		Score: time, Member: postid}).Err()
}

// 获取上次同步的时间戳
func GetLastSyncTime(c context.Context, syncKey string) (string, error) {
	return rdb.Get(c, syncKey).Result()
}

// 设置上次同步的时间戳
func SetLastSyncTime(c context.Context, syncKey string, lastSyncTime int64) error {
	return rdb.Set(c, syncKey, strconv.FormatInt(lastSyncTime, 10), 0).Err()
}

// 获取最近 10 分钟点赞更新过的帖子
func GetTimeAndScore(c context.Context, lastSyncTime, currentTime int64) ([]redis.Z, error) {
	postUpdateScores, err := rdb.ZRangeByScoreWithScores(c, getRedisKey(KeyPostUpdateTimeZSet), &redis.ZRangeBy{
		Min:    fmt.Sprintf("%f", float64(lastSyncTime)),
		Max:    fmt.Sprintf("%f", float64(currentTime)),
		Offset: 0,
		Count:  10000, // 限制每次最多同步 10000 条数据，防止负载过高
	}).Result()
	return postUpdateScores, err
}

// Redis 过期策略：定期清理 30 天前的热度数据
func CleanOldRedisData() {
	c := context.Background()
	thresholdTime := time.Now().Unix() - 30*24*3600 // 30 天前的时间戳
	_, err := rdb.ZRemRangeByScore(c, getRedisKey(KeyPostUpdateTimeZSet), "-inf", strconv.FormatInt(thresholdTime, 10)).Result()
	if err != nil {
		zap.L().Error("清理 Redis 过期数据失败", zap.Error(err))
	} else {
		zap.L().Error("清理 Redis 过期数据完成")
	}
}

// 得到传入的各个帖子的点赞数
func GetLikesByIds(c context.Context, postUpdateScores []redis.Z) []redis.Z {
	result := []redis.Z{}

	for _, item := range postUpdateScores {
		post_id, ok := item.Member.(string)
		if !ok {
			zap.L().Warn("更新不完整", zap.Any("item", item))
			continue
		}
		key := getRedisKey(KeyPostVotedZSetPreix + post_id)
		// 统计 score=1 的总数量
		countOne, err := rdb.ZCount(c, key, "1", "1").Result()
		if err != nil {
			zap.L().Warn("更新不完整", zap.Any("item", item))
			continue
		}
		// 将该 ID 加入结果，score 设为 score=1 的总数量
		result = append(result, redis.Z{
			Member: post_id,
			Score:  float64(countOne),
		})
	}

	return result
}

// 得到传入的各个帖子的热度
func GetHotsByIds(c context.Context, postUpdateScores []redis.Z) []redis.Z {
	result := []redis.Z{}

	for _, item := range postUpdateScores {
		post_id, ok := item.Member.(string)
		if !ok {
			zap.L().Warn("更新不完整", zap.Any("item", item))
			continue
		}
		key := getRedisKey(KeyPostScoreZSet)
		// 统计post_id的热度
		hot, err := rdb.ZScore(c, key, post_id).Result()
		if err != nil {
			zap.L().Warn("更新不完整", zap.Any("item", item))
			continue
		}
		// 将该 ID 加入结果，score 设为 hot
		result = append(result, redis.Z{
			Member: post_id,
			Score:  hot,
		})
	}

	return result
}

// 得到传入的各个帖子的点赞数
func GetDisLikesByIds(c context.Context, postUpdateScores []redis.Z) []redis.Z {
	result := []redis.Z{}

	for _, item := range postUpdateScores {
		post_id, ok := item.Member.(string)
		if !ok {
			zap.L().Warn("更新不完整", zap.Any("item", item))
			continue
		}
		key := getRedisKey(KeyPostVotedZSetPreix + post_id)
		// 统计 score=1 的总数量
		countOne, err := rdb.ZCount(c, key, "-1", "-1").Result()
		if err != nil {
			zap.L().Warn("更新不完整", zap.Any("item", item))
			continue
		}
		// 将该 ID 加入结果，score 设为 score=1 的总数量
		result = append(result, redis.Z{
			Member: post_id,
			Score:  float64(countOne),
		})
	}

	return result
}

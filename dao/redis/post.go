package redis

import (
	"bluebell/models"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

func getIDsFromKey(c *gin.Context, key string, offset int64, limit int64) ([]string, error) {
	start := (offset - 1) * limit
	end := start + limit - 1
	// ZREVRANGE 按照分数从大到小查询指定数量的元素
	return rdb.ZRevRange(c, key, start, end).Result()
}

// GetPostIdsInOrder 从redis获取帖子id
func GetPostIdsInOrder(c *gin.Context, p *models.ParamPostList) ([]string, error) {
	// 查询key
	key := getRedisKey(KeyPostTimeZSet)
	if p.Order == models.OrderScore {
		key = getRedisKey(KeyPostScoreZSet)
	}
	// 确定查询起始点并查询
	return getIDsFromKey(c, key, p.Offset, p.Limit)

}

// GetPostVoteData 从redis获取全部帖子的投赞成票数量
func GetPostVoteData(c *gin.Context, ids []string) ([]int64, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	// 使用pipeline 减少 Redis 请求的 RTT
	pipe := rdb.Pipeline()
	cmds := make([]*redis.IntCmd, len(ids))

	for i, id := range ids {
		key := getRedisKey(KeyPostVotedZSetPreix + id)
		cmds[i] = pipe.ZCount(c, key, "1", "1")
	}

	_, err := pipe.Exec(c)
	if err != nil {
		return nil, err
	}

	data := make([]int64, len(ids))
	for i, cmd := range cmds {
		vote, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		data[i] = vote
	}

	return data, nil
}

// GetCommPostIdsInOrder 按社区从redis获取帖子id
func GetCommPostIdsInOrder(c *gin.Context, p *models.ParamPostList) ([]string, error) {
	// 获取社区的 `ZSET` Key
	cKey := getRedisKey(KeyCommunitySetPreix + strconv.FormatInt(p.Community_id, 10))

	// 定义排序 `ZSET` Key（时间 or 热度）
	orderKey := getRedisKey(KeyPostTimeZSet)
	if p.Order == models.OrderScore {
		orderKey = getRedisKey(KeyPostScoreZSet)
	}

	// 计算 `ZInterStore` 结果的 Key，避免重复计算(排序类型:社区id)
	newKey := getRedisKey(p.Order + ":" + strconv.FormatInt(p.Community_id, 10))

	// 如果 `newKey` 不存在，则计算 `ZInterStore`
	if rdb.Exists(c, newKey).Val() < 1 {
		// 开启 Redis pipeline
		pipeline := rdb.Pipeline()

		// 计算 `ZInterStore`
		pipeline.ZInterStore(c, newKey, &redis.ZStore{
			Keys:      []string{cKey, orderKey},
			Aggregate: "MAX"})

		// 设置缓存过期时间（1 分钟）
		pipeline.Expire(c, newKey, time.Minute)

		// 执行 pipeline
		_, err := pipeline.Exec(c)
		if err != nil {
			return nil, err
		}
	}

	// 从 `newKey` 获取排序后的帖子 ID
	return getIDsFromKey(c, newKey, p.Offset, p.Limit)
}

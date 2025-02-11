package redis

import (
	"bluebell/models"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strconv"
)

func getIDsFromKey(c *gin.Context, key string, offset int64, limit int64) ([]string, error) {
	start := (offset - 1) * limit
	end := start + limit - 1
	// ZREVRANGE 按照分数从大到小查询指定数量的元素
	return rdb.ZRevRange(c, key, start, end).Result()
}

// GetPostIdsInOrder 从redis获取帖子ids并以[]string返回
func GetPostIdsInOrder(c *gin.Context, p *models.ParamPostList) ([]string, error) {
	// 查询key
	key := getRedisKey(KeyPostTimeZSet)
	// 确定查询起始点并查询
	return getIDsFromKey(c, key, p.Offset, p.Limit)
}

// 从redis获取create_time
func GetPostCreateTime(c *gin.Context, postid int64) (float64, error) {
	key := getRedisKey(KeyPostTimeZSet)
	return rdb.ZScore(c, key, strconv.FormatInt(postid, 10)).Result()
}

// 插入帖子的create_time
func UpadtePostCreateTime(c *gin.Context, postid int64, ctimestamp int64) error {
	key := getRedisKey(KeyPostTimeZSet)
	ctimefloat := float64(ctimestamp)
	rdb.ZAdd(c, key, redis.Z{Member: postid, Score: ctimefloat})
	return nil
}

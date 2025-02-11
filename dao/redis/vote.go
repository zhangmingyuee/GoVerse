package redis

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

const (
	oneWeekInSecond = 7 * 24 * 3600
	scorePerVote    = 432 // 每一票贡献的得分
)

var ErrVoteTimeExpire = errors.New("投票时间已过")
var ErrVoteRepeat = errors.New("不允许重复投票")

// CreatePost  创建帖子时redis数据库的变动
func CreatePost(c *gin.Context, postid, commid int64) error {
	pipeline := rdb.Pipeline()
	// 将创建时间添加到redis数据库：帖子时间分数zset中
	pipeline.ZAdd(c, getRedisKey(KeyPostTimeZSet), redis.Z{Score: float64(time.Now().Unix()), Member: postid})

	// 将创建id添加到对应的社区set中
	pipeline.SAdd(c, getRedisKey(KeyCommunitySetPreix+strconv.FormatInt(commid, 10)), postid)

	_, err := pipeline.Exec(c)
	return err
}

func VoteForPost(c *gin.Context, userID, postID string, v float64) error {
	// 1. 判断投票的限制
	postTime, err := rdb.ZScore(c, getRedisKey(KeyPostTimeZSet), postID).Result()
	if err != nil {
		return err
	}

	if float64(time.Now().Unix())-postTime > oneWeekInSecond {
		return ErrVoteTimeExpire
	}
	// 2. 更新分数和投票记录, 放到一个事务中

	// 更新当前用户对当前帖子的投票
	if v == 0 {
		rdb.ZRem(c, getRedisKey(KeyPostVotedZSetPreix+postID), userID)
		return nil
	} else {
		// 查询当前用户对当前帖子之前的投票纪录
		ov := rdb.ZScore(c, getRedisKey(KeyPostVotedZSetPreix+postID), userID).Val()
		// 不允许重复投票
		if ov == v {
			return ErrVoteRepeat
		}

		pipline := rdb.TxPipeline()
		pipline.ZAdd(c, getRedisKey(KeyPostVotedZSetPreix+postID), redis.Z{
			Score:  v,
			Member: userID,
		})
		// 计算并更新当前帖子的得分
		addScore := (v - ov) * scorePerVote
		pipline.ZIncrBy(c, getRedisKey(KeyPostScoreZSet), addScore, postID)
		_, err = pipline.Exec(c)
		return err
	}
}

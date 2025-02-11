package redis

import (
	"bluebell/models"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strconv"
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

package redis

/*
redis key 尽量使用命名空间的方式，方便查询和拆分
*/
const (
	keyPrefix             = "bluebell:"
	KeyPostTimeZSet       = "post:time"   // zset: 帖子及发帖时间
	KeyPostScoreZSet      = "post:score"  // zset: 帖子及投票的分数
	KeyPostVotedZSetPreix = "post:voted:" // zset: 记录用户及投票类型, 参数帖子post_id
)

// 给redis key加上前缀
func getRedisKey(key string) string {
	return keyPrefix + key
}

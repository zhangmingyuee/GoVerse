package redis

/*
redis key 尽量使用命名空间的方式，方便查询和拆分
*/
const (
	keyPrefix             = "bluebell:"
	KeyPostTimeZSet       = "post:time"        // zset: 帖子及发帖时间
	KeyPostScoreZSet      = "post:score"       // zset: 帖子及投票的分数
	KeyPostVotedZSetPreix = "post:voted:"      // zset: 记录用户及投票类型, 参数帖子post_id
	KeyPostUpdateTimeZSet = "post:update_time" // zset: 帖子及给帖子上次投票时间

	LastSyncTimeHotDLikesKey = "last_hot_sync_time" // string: 记录上次点赞数同步时间的 Redis Key

)

// 给redis key加上前缀
func getRedisKey(key string) string {
	return keyPrefix + key
}

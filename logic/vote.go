package logic

import (
	"bluebell/dao/redis"
	"bluebell/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
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

// VoteForPost 为帖子投票的函数
func VoteForPost(c *gin.Context, userID int64, p *models.ParamVoteData) error {
	zap.L().Debug("VoteForPost", zap.Int64("userID", userID), zap.Int64("postid", p.PostID), zap.Int8("direction", p.Direction))
	uidStr := strconv.FormatInt(userID, 10)
	pidStr := strconv.FormatInt(p.PostID, 10)
	return redis.VoteForPost(c, uidStr, pidStr, float64(p.Direction))
}

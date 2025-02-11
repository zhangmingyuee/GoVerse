package logic

import (
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/models"
	"bluebell/pkg/snowflake"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreatePost 创建一个帖子
func CreatePost(c *gin.Context, p *models.Post) error {
	// 1. 生成post id
	p.ID = snowflake.GenID()
	// 3. 帖子信息保存mysql到数据库
	if err := mysql.CreatePost(p); err != nil {
		zap.L().Error("mysql.CreatePost failed", zap.Error(err))
		return err
	}
	// 2. 创建时间保存到redis数据库并返回
	return redis.CreatePost(c, p.ID, p.CommunityID)

}

// GetPostDetail 根据id获得帖子信息（ApiPostDetail）
func GetPostDetail(pid int64) (p *models.ApiPostDetail, err error) {
	// 查询并组合需要的数据
	// 查询帖子信息
	post := new(models.Post)
	if post, err = mysql.GetPostByID(pid); err != nil {
		zap.L().Error("mysql.GetPostByID falied", zap.Error(err))
		return
	}
	// 查询作者信息
	user := new(models.UserSafe)
	if user, err = mysql.GetUserByID(post.AuthorID); err != nil {
		zap.L().Error("mysql.GetUserByID falied", zap.Error(err))
		return
	}

	// 查询社区信息
	commDetail := new(models.CommunityDetail)
	if commDetail, err = mysql.GetCommunityById(post.CommunityID); err != nil {
		zap.L().Error("mysql.GetCommunityByID falied", zap.Error(err))
		return nil, err
	}

	// 填充信息
	p = &models.ApiPostDetail{
		AuthorName:      user.Username,
		Post:            post,
		CommunityDetail: commDetail,
	}
	return
}

// GetPostList 获得全部的帖子信息（ApiPostDetail）
//func GetPostList(offset, limit int64) (apips []*models.ApiPostDetail, err error) {
//	var ps []*models.Post
//	ps, err = mysql.GetPostList(offset, limit)
//	if err != nil {
//		zap.L().Error("mysql.GetPostList falied", zap.Error(err))
//		return
//	}
//
//	for _, post := range ps {
//		// 查询作者信息
//		user := new(models.UserSafe)
//		if user, err = mysql.GetUserByID(post.AuthorID); err != nil {
//			zap.L().Error("mysql.GetUserByID falied", zap.Error(err))
//			return
//		}
//		// 查询社区信息
//		commDetail := new(models.CommunityDetail)
//		if commDetail, err = mysql.GetCommunityById(post.CommunityID); err != nil {
//			zap.L().Error("mysql.GetCommunityByID falied", zap.Error(err))
//			return nil, err
//		}
//		// 填充信息
//		p := &models.ApiPostDetail{
//			AuthorName:      user.Username,
//			Post:            post,
//			CommunityDetail: commDetail,
//		}
//		apips = append(apips, p)
//	}
//	return
//}

// GetPostList2 获取帖子列表
func GetPostList2(c *gin.Context, p *models.ParamPostList) (apips []*models.ApiPostDetail, err error) {
	// 1. 去redis查询id列表
	var ids []string
	ids, err = redis.GetPostIdsInOrder(c, p)
	if err != nil {
		zap.L().Error("redis.GetPostIdsInOrder return 0 data", zap.Error(err))
		return
	}
	if len(ids) == 0 {
		return
	}

	// 2. 根据id去mysql数据库查询帖子信息(按照给定顺序)
	var ps []*models.Post
	ps, err = mysql.GetPostsListByIds(ids)
	if err != nil {
		return
	}

	// 查询帖子赞成票的数量
	var vs []int64
	vs, err = redis.GetPostVoteData(c, ids)

	// 3. 填充帖子的作者和分区信息
	for idx, post := range ps {
		// 查询作者信息
		user := new(models.UserSafe)
		if user, err = mysql.GetUserByID(post.AuthorID); err != nil {
			zap.L().Error("mysql.GetUserByID falied", zap.Error(err))
			return
		}
		// 查询社区信息
		commDetail := new(models.CommunityDetail)
		if commDetail, err = mysql.GetCommunityById(post.CommunityID); err != nil {
			zap.L().Error("mysql.GetCommunityByID falied", zap.Error(err))
			return nil, err
		}
		// 填充信息
		p := &models.ApiPostDetail{
			AuthorName:      user.Username,
			VoteNum:         vs[idx],
			Post:            post,
			CommunityDetail: commDetail,
		}
		apips = append(apips, p)
	}
	return
}

// GetCommPostList 获取社区帖子列表
func GetCommPostList(c *gin.Context, p *models.ParamPostList) (apips []*models.ApiPostDetail, err error) {
	// 1. 按照社区去redis查询id列表
	var ids []string
	ids, err = redis.GetCommPostIdsInOrder(c, p)
	if err != nil {
		zap.L().Error("redis.GetPostIdsInOrder return 0 data", zap.Error(err))
		return
	}
	if len(ids) == 0 {
		return
	}

	// 2. 根据id去mysql数据库查询帖子信息(按照给定顺序)
	var ps []*models.Post
	ps, err = mysql.GetPostsListByIds(ids)
	if err != nil {
		return
	}

	// 查询帖子赞成票的数量
	var vs []int64
	vs, err = redis.GetPostVoteData(c, ids)

	// 3. 填充帖子的作者和分区信息
	for idx, post := range ps {
		// 查询作者信息
		user := new(models.UserSafe)
		if user, err = mysql.GetUserByID(post.AuthorID); err != nil {
			zap.L().Error("mysql.GetUserByID falied", zap.Error(err))
			return
		}
		// 查询社区信息
		commDetail := new(models.CommunityDetail)
		if commDetail, err = mysql.GetCommunityById(post.CommunityID); err != nil {
			zap.L().Error("mysql.GetCommunityByID falied", zap.Error(err))
			return nil, err
		}
		// 填充信息
		p := &models.ApiPostDetail{
			AuthorName:      user.Username,
			VoteNum:         vs[idx],
			Post:            post,
			CommunityDetail: commDetail,
		}
		apips = append(apips, p)
	}
	return
}

// 查询帖子列表（按照score/time/commid查询）
func GetPostListNew(c *gin.Context, p *models.ParamPostList) (apips []*models.ApiPostDetail, err error) {
	if p.Community_id == 0 {
		// 不是按照社区id查询，而是查询全部社区的帖子
		apips, err = GetPostList2(c, p)
		if err != nil {
			zap.L().Error("GetPostList2 falied", zap.Error(err))
			return nil, err
		}
	} else {
		// 按照社区id查询
		apips, err = GetCommPostList(c, p)
		if err != nil {
			zap.L().Error("GetCommPostList falied", zap.Error(err))
			return nil, err
		}
	}
	return apips, nil
}

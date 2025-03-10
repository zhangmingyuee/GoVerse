package logic

import (
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/models"
	"bluebell/pkg/snowflake"
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var ErrNotPostAuthor = errors.New("not the post author")

// CreatePost 创建一个帖子
func CreatePost(c *gin.Context, p *models.Post) error {
	// 1. 生成post id
	p.ID = snowflake.GenID()
	// 3. 帖子信息保存mysql到数据库
	if err := mysql.CreatePost(p); err != nil {
		zap.L().Error("mysql.CreatePost failed", zap.Error(err))
		return err
	}
	return nil

}

// GetPostDetail 根据id获得帖子信息（ApiPostDetail）
func GetPostDetail(uid, pid int64) (p *models.ApiPostDetail, err error) {

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

	// 更新用户行为
	// 更新帖子用户行为信息
	// 更新帖子用户行为信息
	var behavior *models.UserPostBehavior
	behavior, err = mysql.CheckBehavior(uid, pid)
	if err == gorm.ErrRecordNotFound {
		// 没有找到记录, 创建记录
		zap.L().Info("不存在该用户-帖子行为", zap.Int64("postID", pid), zap.Int64("userID", uid))
		if err := mysql.CreateBehavior(uid, pid, mysql.BehaviorBrowser); err != nil {
			zap.L().Error("mysql.CreateBehavior failed", zap.Error(err))
			return nil, err
		}
		return nil, nil // 返回 nil 表示没有记录

	} else if err != nil {
		zap.L().Error("mysql.CheckBehavior failed", zap.Error(err))
		return nil, err
	}

	// 行为存在，查看是否需要更新行为
	if behavior.Browse == 0 {
		if err := mysql.UpdateBehavior(uid, pid, mysql.BehaviorComment); err != nil {
			zap.L().Error("mysql.UpdateBehavior failed", zap.Error(err))
			return nil, err
		}
	}

	return
}

// 查询帖子列表（按照score/time/commid查询）
func GetPostListByScore(c *gin.Context, p *models.ParamPostList) (apips []*models.ApiPostDetail, err error) {
	var ps []*models.Post

	// 1. 如果是根据score排序，则去redis中获取帖子id列表
	if p.Order == "score" {
		var ids []string
		ids, err = redis.GetPostIdsInOrder(c, p)
		//fmt.Println(len(ids))
		if err != nil {
			zap.L().Error("redis.GetPostIdsInOrder failed data", zap.Error(err))
			return
		}
		if len(ids) == 0 {
			return
		}
		// 2. 根据id去mysql数据库查询帖子信息(按照给定顺序)
		if p.Community_id == 0 {
			ps, err = mysql.GetPostsListByIds(ids)
			if err != nil {
				return
			}
		} else {
			ps, err = mysql.GetPostsListByIdsAndComm(p.Community_id, ids)
		}
	} else if p.Order == "time" {
		if p.Community_id == 0 {
			ps, err = mysql.GetPostIdsInTime(p)
			if err != nil {
				zap.L().Error("mysql.GetPostIdsInTime failed data", zap.Error(err))
				return
			}
		} else {
			ps, err = mysql.GetPostIdsInCommTime(p)
			if err != nil {
				zap.L().Error("mysql.GetPostIdsInCommTime failed data", zap.Error(err))
				return
			}

		}
	}

	// 查询帖子赞成票的数量
	var vs []int64
	vs, err = redis.GetPostVoteData(c, ps)

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

// 获取帖子创建时间（优先从 Redis 读取）
func GetPostCreateTimeCached(c *gin.Context, pid int64) (int64, error) {
	// 先尝试从 Redis 获取 `create_time`
	createtimeFloat, err := redis.GetPostCreateTime(c, pid)
	var ctimestamp int64
	if err != nil {
		zap.L().Warn("redis.GetPostCreateTimeCache falied", zap.Error(err))

		// Redis 没有缓存，从 MySQL 查询
		ctimestamp, err = mysql.GetPostCreateTime(pid)
		if err != nil {
			return 0, err
		}
		// 存入 Redis，过期时间设置 24 小时
		err = redis.UpadtePostCreateTime(c, pid, ctimestamp)
		if err != nil {
			zap.L().Error("redis.UpadtePostCreateTimeCache falied", zap.Error(err))
			return 0, err
		}
	} else {
		ctimestamp = int64(createtimeFloat)
		if err != nil {
			zap.L().Error("strconv.ParseInt failed", zap.Error(err))
			return 0, err
		}
	}
	return ctimestamp, nil
}

// UploadImage 上传图片
func UploadImage(userID int64, image *models.ParamImage) error {
	// 判断是否是当前用户的帖子
	selectID, err := mysql.GetPostAuthor(image.PostID)
	if err != nil {
		zap.L().Error("mysql.GetPostAuthor falied", zap.Error(err))
		return err
	}
	if selectID != userID {
		zap.L().Error("非帖子作者", zap.Error(err))
		return ErrNotPostAuthor
	}
	if err := mysql.SaveImageToDB(image); err != nil {
		zap.L().Error("mysql.UploadImage falied", zap.Error(err))
		return err
	}
	return nil
}

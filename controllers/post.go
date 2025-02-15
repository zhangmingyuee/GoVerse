package controllers

import (
	"bluebell/logic"
	"bluebell/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"strconv"
)

// CreatePostHandler 创建一个帖子
// @Summary 创建一个帖子
// @Description 根据传来的json格式的帖子信息，创建一个新的帖子
// @Tags 帖子
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param object query models.Post false "查询参数"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponsePostList
// @Router /post [post]
func CreatePostHandler(c *gin.Context) {
	// CreatePostHandler 创建一个帖子
	// 1. 获取参数及参数校验
	p := new(models.Post)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("Create post with invalid param", zap.Error(err))
		if errs, ok := err.(validator.ValidationErrors); !ok {
			ResponseError(c, CodeInvalidParam)
			return
		} else {
			ResponseErrorWithMcg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
			return
		}
	}
	// 从上下文取到当前用户的id
	userid, err := getCurrentUserID(c)
	if err != nil {
		zap.L().Error("Get user id failed", zap.Error(err))
		ResponseError(c, CodeNeedLogin)
		return
	}
	p.AuthorID = userid

	// 2. 创建帖子
	if err := logic.CreatePost(c, p); err != nil {
		zap.L().Error("Create post failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// 3. 返回响应
	ResponseSuccess(c, nil)
}

// GetPostDetailHandler 获取帖子详情的处理函数
func GetPostDetailHandler(c *gin.Context) {
	// 1. 获取参数（url获取帖子的id）
	pidStr := c.Param("id")
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		zap.L().Error("Get post detail with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 2. 根据id取得帖子的数据
	var p *models.ApiPostDetail
	if p, err = logic.GetPostDetail(pid); err != nil {
		zap.L().Error("logic.GetPostDetail(pid) failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// 3. 返回相应
	ResponseSuccess(c, p)
}

// GetPostListHandler 获取全部帖子详情的处理函数的升级版
// 根据前端传来的参数（创建时间/分数）动态地获取帖子列表
// 1. 获取参数
// 2. 去redis查询id列表
// 3. 根据id去数据库查询帖子信息
func GetPostListHandler(c *gin.Context) {
	// GET请求参数（query string）: /api/v1/post2?page=1&size=10&order=time
	// 初始化结构体时指定初始参数
	p := &models.ParamPostList{
		Offset:       1,
		Limit:        10,
		Order:        models.OrderTime,
		Community_id: 0,
	}
	if err := c.ShouldBindQuery(p); err != nil {
		zap.L().Error("Get post list 2 with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 获取全部帖子
	ps, err := logic.GetPostListByScore(c, p)
	if err != nil {
		zap.L().Error("logic.GetPostList failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, ps)
}

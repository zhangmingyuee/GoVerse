package controllers

import (
	"bluebell/logic"
	"bluebell/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"strconv"
)

// --------- 全部与社区相关的 -----------

func GetCommunityHandler(c *gin.Context) {
	// 查询到所有的社区(community_id, community_name)并以列表的形式返回
	data, err := logic.GetCommunity()
	if err != nil {
		zap.L().Error("logic.GetCommunity() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}

// CreateCommunityHandler 创建一个社区
func CreateCommunityHandler(c *gin.Context) {
	// 1. 获取参数及参数校验
	comm := new(models.CommunityDetail)
	if err := c.ShouldBindJSON(comm); err != nil {
		zap.L().Error("Create community with invalid param", zap.Error(err))
		if errs, ok := err.(validator.ValidationErrors); !ok {
			ResponseError(c, CodeInvalidParam)
			return
		} else {
			ResponseErrorWithMcg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
			return
		}
	}
	// 2. 创建社区
	if err := logic.CreateCommunity(comm); err != nil {
		zap.L().Error("Create community failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 3. 返回响应
	ResponseSuccess(c, nil)
}

// CommunityDetailHandler 社区分类详情
func CommunityDetailHandler(c *gin.Context) {
	// 1. 获取社区id
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 2. 根据id查询社区详细信息
	commdetail, err := logic.GetCommunityDetail(id)
	if err != nil {
		zap.L().Error("logic.GetCommunityDetail() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, commdetail)
}

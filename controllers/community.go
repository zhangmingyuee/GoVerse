package controllers

import (
	"bluebell/logic"
	"bluebell/models"
	"database/sql"
	"errors"
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
	// 1. 获取社区id/name
	idStr := c.Query("uid")
	nameStr := c.Query("uname")
	// 没有传入参数，显示全部社区
	if idStr == "" && nameStr == "" {
		GetCommunityHandler(c)
		return
	}
	// 传入社区uid
	if idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			ResponseError(c, CodeInvalidParam)
			return
		}
		// 2. 根据id查询社区详细信息
		commdetail, err := logic.GetCommunityDetail(id)
		if errors.Is(err, sql.ErrNoRows) { // 使用 errors.Is 进行错误比较，兼容性更好
			zap.L().Warn("there is no community in db", zap.Int64("community_id", id))
			ResponseError(c, CodeCommNotExist)
			return
		}
		if err != nil {
			zap.L().Error("logic.GetCommunityDetail() failed", zap.Error(err))
			ResponseError(c, CodeServerBusy)
			return
		}
		ResponseSuccess(c, commdetail)
		return
	}
	// 传入社区uname
	if nameStr != "" {
		// 根据uname查询社区详细信息
		commdetail, err := logic.GetCommunityByNameDetail(nameStr)
		if errors.Is(err, sql.ErrNoRows) { // 使用 errors.Is 进行错误比较，兼容性更好
			zap.L().Warn("there is no community in db", zap.String("community_name", nameStr))
			ResponseError(c, CodeCommNotExist)
			return
		}
		if err != nil {
			zap.L().Error("logic.GetCommunityByNameDetail() failed", zap.Error(err))
			ResponseError(c, CodeServerBusy)
			return
		}
		ResponseSuccess(c, commdetail)
		return
	}

}

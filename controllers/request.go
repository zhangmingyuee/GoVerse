package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
)

const (
	CtxTokenKey  = "token"
	CtxUserIDKey = "userid"
)

var ErrorUserNotLogin = errors.New("用户未登录")

// getCurrentUserID 获取当前登录的用户id
func getCurrentUserID(c *gin.Context) (userid int64, err error) {
	uid, ok := c.Get(CtxUserIDKey)
	if !ok {
		err = ErrorUserNotLogin
		return
	}
	userid, ok = uid.(int64)
	if !ok {
		err = ErrorUserNotLogin
		return
	}
	return
}

func getPageInfo(c *gin.Context) (offset int64, limit int64) {
	// 获取分页参数并转换
	offsetStr := c.Query("offset")
	limitStr := c.Query("limit")

	// 解析 offset
	var err error
	fmt.Println("offsetStr, limitStr", offsetStr, limitStr)
	offset, err = strconv.ParseInt(offsetStr, 10, 64)
	if err != nil || offset < 0 { // 额外处理负数情况
		offset, limit = 1, 10
		zap.L().Error("Get post list with invalid offset param", zap.String("offset", offsetStr), zap.Error(err))
		//ResponseError(c, CodeInvalidParam)
		return
	}
	// 解析 limit
	limit, err = strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limit <= 0 { // 确保 limit > 0，避免 limit = 0 的无效查询
		offset, limit = 1, 10
		zap.L().Error("Get post list with invalid limit param", zap.String("limit", limitStr), zap.Error(err))
		//ResponseError(c, CodeInvalidParam)
		return
	}
	return
}

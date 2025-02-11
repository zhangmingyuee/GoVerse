package controllers

import (
	"bluebell/dao/mysql"
	"bluebell/logic"
	"bluebell/models"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"strconv"
)

// SignUpHandler 用户注册请求函数
func SignUpHandler(c *gin.Context) {
	// 1. 获取参数和参数校验
	//var p models.ParamSignUp
	p := new(models.ParamSignUp)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("SignUp with invalid param", zap.Error(err))
		// 判断err是不是validator.ValidationErrors类型(存储多个字段的校验错误信息的错误类型)
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			// 说明 JSON 解析失败
			ResponseError(c, CodeInvalidParam)
			return
		}
		// 说明 JSON 解析 成功，但是 字段数据校验失败
		ResponseErrorWithMcg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
		return
	}

	// 2. 业务处理
	if err := logic.Signup(p); err != nil {
		zap.L().Error("SignUp logic failed", zap.Error(err))
		if errors.Is(err, mysql.ErrorUserExist) {
			ResponseError(c, CodeUserExist)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}

	// 3. 返回响应
	ResponseSuccess(c, nil)

}

// LogInHandler 用户登录请求函数
func LogInHandler(c *gin.Context) {
	// 1. 获取登录参数
	p := new(models.ParamLogIn)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("LogIn with invalid param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMcg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
		return
	}

	// 2. 业务处理
	user, err := logic.LogIn(c, p)
	if err != nil {
		zap.L().Error("LogIn logic failed", zap.String("username", p.Username), zap.Error(err))
		if errors.Is(err, mysql.ErrorUserNotExist) {
			ResponseError(c, CodeUserNotExist)
			return
		}
		if errors.Is(err, mysql.ErrorUserPassword) {
			ResponseError(c, CodeInvalidPassword)
			return
		}
		if errors.Is(err, logic.ErrorOPTExpired) {
			ResponseError(c, CodeOTPExpired)
			return
		}
		if errors.Is(err, logic.ErrorOPTInvalid) {
			ResponseError(c, CodeOTPInvalid)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, gin.H{
		"userid":   fmt.Sprintf("%d", user.UserID), // id值大于2**53-1时，json超出范围
		"username": user.Username,
		"token":    user.Token,
	})
}

// LogOutHandler 用户注销请求函数
func LogOutHandler(c *gin.Context) {
	token, ok := c.Get(CtxTokenKey)
	if !ok {
		zap.L().Error("c.Get(CtxTokenKey) failed", zap.Any(CtxTokenKey, token))
		ResponseError(c, CodeServerBusy)
		return
	}

	//fmt.Println("token:", token)
	if err := logic.LogOut(token.(string)); err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
	//c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// GenerateOTPHandler 用户登录生成验证码请求函数
func GenerateOTPHandler(c *gin.Context) {
	// 验证用户是否存在
	req := new(models.ParamUsernameRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	userID, flag := logic.GetExistUser(req)
	if !flag {
		ResponseError(c, CodeUserNotExist)
		return
	}

	// 生成 6 位验证码
	err, otp := logic.GenerateOTP(userID, c)
	if err != nil {
		zap.L().Error("GenerateOTP logic failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, otp)
}

// GetUserInfoHandler 获取用户信息请求函数
func GetUserInfoHandler(c *gin.Context) {
	// 根据jwt中的用户id获取用户信息
	userIDAny, ok := c.Get(CtxUserIDKey)
	if !ok {
		ResponseError(c, CodeServerBusy)
		return
	}
	userID := userIDAny.(int64)
	user, err := logic.GetUserInfo(userID)
	if err != nil {
		zap.L().Error("GetUserInfo logic failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, user)

}

// ModifyUserInfoHandler 修改用户信息
func ModifyUserInfoHandler(c *gin.Context) {
	// map 动态字段更新，用户可以选择更新不同字段
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		zap.L().Error("ModifyUserInfo with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 根据jwt中的用户id获取用户信息
	userIDAny, ok := c.Get(CtxUserIDKey)
	if !ok {
		ResponseError(c, CodeServerBusy)
		return
	}
	userID := userIDAny.(int64)

	if err := logic.ModifyUserInfo(userID, updates); err != nil {
		zap.L().Error("ModifyUserInfo logic failed", zap.Error(err))
		if errors.Is(err, logic.ErrorModifyNil) {
			ResponseError(c, CodeModifyNil)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}
	GetUserInfoHandler(c)
	//ResponseSuccess(c, nil)
}

// ModifyPasswordHandler 修改密码
func ModifyPasswordHandler(c *gin.Context) {
	p := new(models.ParamPassword)
	userIDAny, ok := c.Get(CtxUserIDKey)
	userID := userIDAny.(int64)
	if !ok {
		zap.L().Error("c.Get(CtxUserIDKey) failed", zap.Any(CtxUserIDKey, userID))
		ResponseError(c, CodeServerBusy)
		return
	}
	if err := logic.ModifyPassword(userID, p); err != nil {
		zap.L().Error("ModifyPassword logic failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

// GetUserByNameHandler 根据用户名查询他的信息
func GetUserByNameHandler(c *gin.Context) {
	username := c.Param("name")
	if len(username) == 0 {
		ResponseError(c, CodeInvalidParam)
		return
	}

	user, err := logic.GetUserByName(username)
	if err != nil {
		zap.L().Error("GetUserByName logic failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, user)

}

// GetUsersByNameHandler 根据用户名匹配用户们
func GetUsersByNameHandler(c *gin.Context) {
	p := new(models.ParamUsername)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("GetUsersByName with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	fmt.Println(p, len(p.Username))
	if len(p.Username) == 0 {
		// 显示全部的用户
		users, err := logic.GetUsers()
		if err != nil {
			zap.L().Error("logic.GetUsers failed", zap.Error(err))
			ResponseError(c, CodeServerBusy)
			return
		}
		ResponseSuccess(c, users)
		return
	}
	// 根据传入的用户名查询用户
	users, err := logic.GetUsersByName(p.Username)
	if err != nil {
		zap.L().Error("GetUsersByName logic failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, users)
}

// GetUserPostHandler 得到用户对应的帖子
func GetUserPostHandler(c *gin.Context) {
	userIDStr := c.Param("id")
	if len(userIDStr) == 0 {
		zap.L().Error("GetUserPost with empty info", zap.String("id", userIDStr))
		ResponseError(c, CodeInvalidParam)
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		zap.L().Error("strconv.ParseInt failed", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 处理获取用户帖子的逻辑
	var userpost *models.UserPost
	userpost, err = logic.GetUserPosts(userID)
	if err != nil {
		zap.L().Error("logic.GetUserPosts failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, userpost)
}

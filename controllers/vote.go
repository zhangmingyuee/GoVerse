package controllers

import (
	"bluebell/logic"
	"bluebell/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// 也可以直接也到帖子里，这里为了看的清晰，单独加了一个vote

// PostVoteController 用户投票控制
func PostVoteController(c *gin.Context) {
	p := new(models.ParamVoteData)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("PostVoteController ShouldBindJSON error", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors) // 类型断言
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMcg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
		return
	}

	// 获取用户id
	userid, err := getCurrentUserID(c)
	if err != nil {
		zap.L().Error("getCurrentUserID error", zap.Error(err))
		ResponseError(c, CodeNeedLogin)
		return
	}
	// 投票
	if err := logic.VoteForPost(c, userid, p); err != nil {
		zap.L().Error("logic.VoteForPost error", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

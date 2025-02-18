package controllers

import (
	"bluebell/logic"
	"bluebell/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

// CreateCommentConntroller 创建评论的处理函数
func CreateCommentConntroller(c *gin.Context) {
	req := new(models.ParamComment)

	// 绑定请求数据到结构体，并进行校验
	if err := c.ShouldBind(req); err != nil {
		zap.L().Error("CreateCommentConntroller: ShouldBind(&req)", zap.Error(err))
		// 校验失败
		if validationErrors, ok := err.(validator.ValidationErrors); !ok {
			ResponseError(c, CodeInvalidParam)
			return
		} else {
			ResponseErrorWithMcg(c, CodeInvalidParam, removeTopStruct(validationErrors.Translate(trans)))
			return
		}
	}

	userID, err := getCurrentUserID(c)
	if err != nil {
		zap.L().Error("获取用户ID失败", zap.Error(err))
		return
	}
	req.UserID = userID

	// 调用业务逻辑层，创建评论
	commentID, err := logic.CreateComment(req.PostID, req.ParentID, req.UserID, req.Content)
	if err != nil {
		zap.L().Error("CreateCommentConntroller: logic.CreateComment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}
	req.Comment_id = commentID

	// 成功返回
	ResponseSuccess(c, req)
}

// GetCommentConntroller 获取顶级评论的处理函数
func GetCommentConntroller(c *gin.Context) {
	// 从 URL 路径获取 post_id 参数
	postIDStr := c.Param("post_id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		zap.L().Error("invalid post_id", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	if comments, err := logic.GetCommentByPostID(postID); err != nil {
		zap.L().Error("GetCommentConntroller: logic.GetCommentByPostID", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	} else {
		ResponseSuccess(c, comments)
	}
}

// GetChildCommentsController 展开某个父评论下的子评论
func GetChildCommentsController(c *gin.Context) {
	// 从 URL 路径获取 parent_id 参数，例如：/comment/child/:parent_id
	parentIDStr := c.Param("parent_id")
	parentID, err := strconv.ParseInt(parentIDStr, 10, 64)
	if err != nil {
		zap.L().Error("invalid parent_id", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	comments, err := logic.GetChildComments(parentID)
	if err != nil {
		zap.L().Error("logic.GetChildComments error", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, comments)
}

// DeleteCommentController 删除评论处理函数
func DeleteCommentController(c *gin.Context) {
	commentIDParam := new(models.ParamDeleteComment)
	if err := c.ShouldBindJSON(commentIDParam); err != nil {
		zap.L().Error("DeleteCommentController: ShouldBindJSON(&req)", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 调用业务逻辑层删除评论
	if err := logic.DeleteComment(commentIDParam.CommentID); err != nil {
		zap.L().Error("DeleteCommentController: logic.DeleteComment", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// 成功删除评论，返回成功响应
	ResponseSuccess(c, nil)
}

// PinCommentController 处理置顶评论的请求
// 例如，路由为：DELETE 或 POST /comment/pin/:comment_id
func PinCommentController(c *gin.Context) {
	commentIDStr := c.Param("comment_id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		zap.L().Error("Invalid comment_id", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 获取用户id
	userID, err := getCurrentUserID(c)
	if err != nil {
		zap.L().Error("获取用户id失败", zap.Error(err))
		return
	}

	// 调用逻辑层置顶评论
	if err := logic.PinComment(commentID, userID); err != nil {
		zap.L().Error("Failed to pin comment", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, nil)
}

// UnpinCommentController 处理取消置顶评论的请求
func UnpinCommentController(c *gin.Context) {
	commentIDStr := c.Param("comment_id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		zap.L().Error("Invalid comment_id", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}

	userID, err := getCurrentUserID(c)
	if err != nil {
		zap.L().Error("获取用户id失败", zap.Error(err))
		return
	}

	// 调用逻辑层取消置顶评论
	if err := logic.UnpinComment(commentID, userID); err != nil {
		zap.L().Error("Failed to unpin comment", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

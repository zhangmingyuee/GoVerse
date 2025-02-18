package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"bluebell/pkg/snowflake"
	"go.uber.org/zap"
)

// CreateComment 创建评论逻辑
func CreateComment(postID, parentID, userID int64, content string) (int64, error) {
	// 生成评论id
	commentID := snowflake.GenID()

	// 调用 DAO 层，保存评论到数据库
	commentID, err := mysql.CreateComment(commentID, postID, parentID, userID, content)
	if err != nil {
		zap.L().Error("mysql.CreateComment failed", zap.Error(err))
		return 0, err
	}

	// 可以在此添加额外的业务逻辑，比如计算评论热度等
	return commentID, nil
}

// GetCommentByPostID 查看某个帖子的顶级评论
func GetCommentByPostID(postID int64) ([]*models.Comment, error) {
	comments, err := mysql.GetCommentByPostID(postID)
	if err != nil {
		zap.L().Error("mysql.getCommentByPostID failed", zap.Error(err))
		return nil, err
	}
	return comments, nil
}

// GetChildComments 获取指定父评论下的所有子评论
func GetChildComments(parentID int64) ([]*models.Comment, error) {
	if comments, err := mysql.GetChildCommentsByParentID(parentID); err != nil {
		zap.L().Error("mysql.GetChildCommentsByParentID failed", zap.Error(err))
		return nil, err
	} else {
		return comments, nil
	}
}

// DeleteComment 删除评论
func DeleteComment(commentID int64) error {
	// 删除 comment_id 对应的评论
	if err := mysql.DeleteCommentByID(commentID); err != nil {
		zap.L().Error("mysql.DeleteCommentByID failed", zap.Error(err))
		return err
	}

	// 删除评论该评论下的子评论
	if err := mysql.DeleteCommentByParentID(commentID); err != nil {
		zap.L().Error("mysql.DeleteCommentByParentID failed", zap.Error(err))
		return err
	}

	return nil
}

// PinComment 置顶帖子的处理逻辑
func PinComment(commentID, userID int64) error {
	// 判断commentID评论是否是userID的帖子
	if err := mysql.CheckCoometAndPost(commentID, userID); err != nil {
		zap.L().Error("mysql.CheckCoometAndPost failed", zap.Error(err))
		return err
	}
	// 是本人的帖子下的评论，执行置顶操作
	return mysql.PinCommentByID(commentID)
}

// UnpinComment 取消置顶帖子的处理逻辑
func UnpinComment(commentID, userID int64) error {
	// 判断commentID评论是否是userID的帖子
	if err := mysql.CheckCoometAndPost(commentID, userID); err != nil {
		zap.L().Error("mysql.CheckCoometAndPost failed", zap.Error(err))
		return err
	}
	// 是本人的帖子下的评论，执行取消置顶操作
	return mysql.UnpinCommentByID(commentID, userID)
}

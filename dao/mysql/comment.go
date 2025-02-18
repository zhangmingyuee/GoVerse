package mysql

import (
	"bluebell/models"
	"time"
)

// CreateComment 发表评论
func CreateComment(commentID, postID, parentID, userID int64, content string) (int64, error) {
	strSql := `
        INSERT INTO comments (comment_id, post_id, parent_id, user_id, content)
        VALUES (?, ?, ?, ?, ?)
    `
	result, err := db.Exec(strSql, commentID, postID, parentID, userID, content)
	if err != nil {
		return 0, err
	}

	// 获取插入的评论ID
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastInsertID, nil
}

// GetCommentByPostID 查看某一个帖子的全部顶级评论
func GetCommentByPostID(postID int64) ([]*models.Comment, error) {
	strSql := `select post_id, comment_id, parent_id, user_id, content, likes, 
            dislikes, status, create_time, update_time FROM comments 
        WHERE post_id = ? and parent_id = 0
        ORDER BY create_time DESC;`

	// 执行查询，获取结果
	var comments []*models.Comment
	err := db.Select(&comments, strSql, postID)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// GetChildCommentsByParentID 查询指定父评论下的子评论
func GetChildCommentsByParentID(parentID int64) ([]*models.Comment, error) {
	strSql := `
		SELECT 
			 comment_id, post_id, parent_id, user_id, content, likes, dislikes, status, create_time, update_time
		FROM comments 
		WHERE parent_id = ? 
		ORDER BY create_time DESC;
	`
	var comments []*models.Comment
	err := db.Select(&comments, strSql, parentID)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// DeleteCommentByID 删除一个评论
func DeleteCommentByID(commentID int64) error {
	strSql := `delete from comments where comment_id = ?;`
	_, err := db.Exec(strSql, commentID)
	if err != nil {
		return err
	}
	return nil
}

// DeleteCommentByParentID 删除以某评论为父评论的子评论
func DeleteCommentByParentID(parentID int64) error {
	strSql := `delete from comments where parent_id = ?;`
	_, err := db.Exec(strSql, parentID)
	if err != nil {
		return err
	}
	return nil
}

// PinCommentByID 更新评论为置顶状态
func PinCommentByID(commentID int64) error {
	strSql := `
		UPDATE comments 
		SET is_top = 1, top_time = ? 
		WHERE comment_id = ?;
	`
	_, err := db.Exec(strSql, time.Now(), commentID)
	if err != nil {
		return err
	}
	return nil
}

// UnpinCommentByID 取消置顶
func UnpinCommentByID(commentID, userID int64) error {
	strSql := `
		UPDATE comments 
		SET is_top = 0, top_time = NULL 
		WHERE comment_id = ?;
	`
	_, err := db.Exec(strSql, commentID)
	if err != nil {
		return err
	}
	return nil
}

// CheckCoometAndPost 判断comment_id对应的帖子的作者是否是userID
func CheckCoometAndPost(commentID, userID int64) error {
	strSql := `
		SELECT p.author_id
		FROM comments c
		JOIN post p ON c.post_id = p.post_id
		WHERE c.comment_id = ? AND p.author_id = ?;
	`

	var postAuthorID int64
	return db.Get(&postAuthorID, strSql, commentID, userID)
}

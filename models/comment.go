package models

// Comment 用于绑定评论相关的请求数据
type Comment struct {
	ID         int64  `json:"-" db:"id" form:"id" binding:"omitempty"`                         // 评论 ID，数据库自增主键
	CommentID  int64  `json:"comment_id" db:"comment_id" form:"comment_id" binding:"required"` // 评论的唯一标识
	PostID     int64  `json:"post_id" db:"post_id" form:"post_id" binding:"required"`          // 帖子 ID，必填
	ParentID   *int64 `json:"parent_id" db:"parent_id" form:"parent_id"`                       // 父评论 ID，可空
	UserID     int64  `json:"user_id" db:"user_id" form:"user_id" binding:"required"`          // 用户 ID，必填
	Content    string `json:"content" db:"content" form:"content" binding:"required"`          // 评论内容，必填
	Likes      int    `json:"likes" db:"likes" form:"likes" binding:"required"`                // 点赞数，默认为 0
	Dislikes   int    `json:"dislikes" db:"dislikes" form:"dislikes" binding:"required"`       // 点踩数，默认为 0
	Status     int8   `json:"status" db:"status" form:"status" binding:"required"`             // 评论状态，默认为 1
	CreateTime string `json:"create_time" db:"create_time" form:"create_time"`                 // 创建时间
	UpdateTime string `json:"update_time" db:"update_time" form:"update_time"`                 // 更新时间
}

type ParamDeleteComment struct {
	CommentID int64 `json:"comment_id,string" db:"comment_id" form:"comment_id" binding:"required"`
}

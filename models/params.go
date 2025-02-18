package models

import "time"

// 帖子排序方式
const (
	OrderTime  = "time"
	OrderScore = "score"
)

// ParamSignUp 注册请求参数
type ParamSignUp struct {
	Username   string    `json:"uname" binding:"required"`
	Password   string    `json:"upwd" binding:"required"`
	RePassword string    `json:"urepwd" binding:"required,eqfield=Password"`
	CreatedAt  time.Time `db:"create_time" json:"created_at"` // 账户创建时间
}

// ParamLogIn 登录请求参数
type ParamLogIn struct {
	Username     string `json:"uname" binding:"required"`
	Password     string `json:"upwd" binding:"required"`
	IdentifyCode string `json:"identify_code" binding:"required"`
}

// ParamPostList 查询帖子请求参数（按照某一顺序）
type ParamPostList struct {
	Offset       int64  `json:"offset" form:"offset"`
	Limit        int64  `json:"limit" form:"limit"`
	Order        string `json:"order" form:"order"`
	Community_id int64  `json:"community_id" form:"community_id"`
}

// ParamUsernameRequest 用户登录获取验证码请求结构
type ParamUsernameRequest struct {
	Username string `json:"uname" binding:"required"`
}

// ParamPassword 修改密码请求结构
type ParamPassword struct {
	OrPassword string `json:"oupwd" binding:"required"`
	Password   string `json:"upwd" binding:"required"`
	RePassword string `json:"urepwd" binding:"required,eqfield=Password"`
}

// ParamUsername
type ParamUsername struct {
	Username string `json:"uname"`
}

// ParamVoteData 投票数据
type ParamVoteData struct {
	// UserID 从请求中换取当前用户
	PostID    int64 `json:"post_id,string" binding:"required"`
	Direction int8  `json:"direction,string" binding:"oneof=1 0 -1"`
}

// 定义用于创建评论请求的结构体
type ParamComment struct {
	Comment_id int64  `json:"comment_id,string" db:"comment_id" binding:"omitempty"`
	PostID     int64  `json:"post_id,string" db:"post_id" form:"post_id" binding:"required"`        // 帖子ID，必填
	ParentID   int64  `json:"parent_id,string" db:"parent_id" form:"parent_id" binding:"omitempty"` // 父评论ID，非必填
	UserID     int64  `json:"user_id,string" db:"user_id" form:"user_id"`                           // 用户ID，必填
	Content    string `json:"content" db:"content" form:"content" binding:"required"`               // 评论内容，必填
}

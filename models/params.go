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

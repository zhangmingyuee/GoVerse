package models

// UserPostBehavior 表示用户对文章的行为，如点赞、评论等
type UserPostBehavior struct {
	UserID  int64 `json:"uid" db:"user_id"`
	PostID  int64 `json:"post_id" db:"post_id"`
	Rate    int8  `json:"rate" db:"rate"` // 用户对文章的评分或行为分数
	Browse  int8  `json:"browse" db:"browse"`
	Like    int8  `json:"like" db:"like"`
	Comment int8  `json:"comment" db:"comment"`
}

func (UserPostBehavior) TableName() string {
	return "user_post_behavior"
}

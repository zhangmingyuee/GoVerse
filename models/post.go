package models

import (
	"time"
)

// Post 帖子的结构体
type Post struct {
	ID          int64     `json:"id,string" db:"post_id"`
	AuthorID    int64     `json:"author_id,string" db:"author_id"`
	CommunityID int64     `json:"community_id,string" db:"community_id" binding:"required"`
	Status      int32     `json:"status,string" db:"status"`
	Title       string    `json:"title" db:"title" binding:"required"`
	Content     string    `json:"content" db:"content" binding:"required"`
	CreateTime  time.Time `json:"create_time" db:"create_time"`
	Likes       int64     `json:"likes,string" db:"likes"`
	DisLikes    int64     `json:"dislikes,string" db:"d"`
}

// TableName 方法用于指定 GORM 使用的表名
func (Post) TableName() string {
	return "post" // 确保使用 "images" 表
}

type ApiPostDetail struct {
	AuthorName       string `json:"author_name" db:"author_name"`
	VoteNum          int64  `json:"votes"`
	*Post            `json:"post_detail"`
	*CommunityDetail `json:"community_detail"`
}

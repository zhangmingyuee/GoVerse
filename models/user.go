package models

import "time"

type User struct {
	UserID    int64     `gorm:"primaryKey;autoIncrement" db:"user_id" json:"user_id"`   // 用户唯一 ID，自增主键
	Username  string    `gorm:"uniqueIndex;size:100" db:"username" json:"username"`     // 用户名，唯一
	Password  string    `gorm:"size:255" db:"password" json:"-"`                        // 加密后的密码
	Token     string    `gorm:"-" json:"token,omitempty"`                               // JWT Token（不会存入数据库）
	Email     string    `gorm:"uniqueIndex;size:255" db:"email" json:"email,omitempty"` // 用户邮箱，唯一
	CreatedAt time.Time `gorm:"autoCreateTime" db:"create_time" json:"created_at"`      // 账户创建时间
	UpdatedAt time.Time `gorm:"autoUpdateTime" db:"update_time" json:"updated_at"`      // 账户最后更新时间
	AvatarURL string    `gorm:"size:255" db:"avatar_url" json:"avatar_url"`             // 头像 URL
	Bio       string    `gorm:"size:500" db:"bio" json:"bio,omitempty"`                 // 个人简介
}

type UserSafe struct {
	UserID    int64      `gorm:"primaryKey;autoIncrement" db:"user_id" json:"user_id"`        // 用户唯一 ID，自增主键
	Username  string     `gorm:"uniqueIndex;size:100" db:"username" json:"username"`          // 用户名，唯一
	Email     string     `gorm:"uniqueIndex;size:255" db:"email" json:"email,omitempty"`      // 用户邮箱，唯一
	CreatedAt *time.Time `gorm:"autoCreateTime" db:"create_time" json:"created_at,omitempty"` // 账户创建时间
	UpdatedAt *time.Time `gorm:"autoUpdateTime" db:"update_time" json:"updated_at,omitempty"` // 账户最后更新时间
	AvatarURL string     `gorm:"size:255" db:"avatar_url" json:"avatar_url"`                  // 头像 URL
	Bio       string     `gorm:"size:500" db:"bio" json:"bio,omitempty"`                      // 个人简介
}

// 用户和全部帖子
type UserPost struct {
	User  *UserSafe
	Posts []*Post
}

// JWTBlacklist 记录已失效的Token
type JWTBlacklist struct {
	ID        uint   `gorm:"primaryKey"`
	Token     string `gorm:"type:varchar(512);uniqueIndex"` // 改成 VARCHAR(512)
	CreatedAt time.Time
}

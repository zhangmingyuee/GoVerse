package mysql

import (
	"bluebell/models"
	"database/sql"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// encryptPassword 使用 bcrypt 加密密码
func encryptPassword(password string) (string, error) {
	// 生成 bcrypt 哈希（使用默认成本）
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckUserExist 检查指定用户名的用户是否存在
func CheckUserExist(username string) (err error) {
	sqlStr := `select count(user_id) from user where username=?`
	var count int
	if err = db.Get(&count, sqlStr, username); err != nil {
		return err
	}
	if count > 0 {
		return ErrorUserExist
	}
	return
}

// InsertUser 向数据库中插入一条新的用户记录
func InsertUser(user *models.User) (err error) {
	// 对密码加密
	user.Password, err = encryptPassword(user.Password)
	if err != nil {
		return err
	}
	// 执行SQL语句入库
	sqlStr := `insert into user (user_id,username,password,create_time,avatar_url) values(?,?,?,?,?)`
	_, err = db.Exec(sqlStr, user.UserID, user.Username, user.Password, user.CreatedAt, user.AvatarURL)
	return
}

// CheckLogin 检查是否登录成功
func CheckLogin(user *models.User) (err error) {
	pwd := user.Password
	sqlStr := `select user_id, username, password from user where username=?`
	err = db.Get(user, sqlStr, user.Username)
	if err == sql.ErrNoRows {
		return ErrorUserNotExist
	}
	if err != nil {
		return err
	}
	// 比较数据库中的密码和输入密码是否匹配
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pwd))
	if err != nil {
		return ErrorUserPassword
	}
	return
}

// GetUserByID 根据ID查询用户信息
func GetUserByID(userID int64) (user *models.UserSafe, err error) {
	user = new(models.UserSafe)
	sqlStr := `select user_id, username, COALESCE(email,'') as email, create_time, update_time, avatar_url, COALESCE(bio,'') as bio from user where user_id=?`
	err = db.Get(user, sqlStr, userID)
	return
}

// GetExistUser 判断是否存在用户
func GetExistUser(username string) (userID int64, err error) {
	//fmt.Println("username: ", username)
	sqlStr := `select user_id from user where username=?`
	err = db.Get(&userID, sqlStr, username)
	//fmt.Println("userID", userID)
	return

}

// GetUserInfo 获取用户的全部信息
func GetUserInfo(userID int64) (user *models.User, err error) {
	user = new(models.User)
	//fmt.Println(userID)
	sqlStr := `select user_id, username, COALESCE(email, '') as email, create_time, update_time, COALESCE(avatar_url, '') as avatar_url, COALESCE(bio, '') as bio from user where user_id=?`
	err = db.Get(user, sqlStr, userID)
	if err != nil {
		return nil, err
	}
	return
}

// ModifyUserInfo 修改用户信息
func ModifyUserInfo(userID int64, updates map[string]interface{}) (err error) {
	return gormdb.Table("user").Where("user_id=?", userID).Updates(updates).Error
}

// AddToBlacklist 将本次登录token增加到黑名单
func AddToBlacklist(token string) error {
	blacklistEntry := &models.JWTBlacklist{Token: token}
	return gormdb.Create(blacklistEntry).Error
}

// IsTokenBlacklisted 检查 Token 是否在黑名单
func IsTokenBlacklisted(token string) bool {
	var count int64
	gormdb.Model(&models.JWTBlacklist{}).Where("token = ?", token).Count(&count)
	return count > 0
}

// CheckPassword 验证密码是否正确
func CheckPassword(userID int64, password string) (flag bool, err error) {
	//fmt.Println(userID)
	var entrptPass string
	query := "SELECT password FROM user WHERE user_id = ?"
	// 执行 SQL 语句
	if err = db.Get(&entrptPass, query, userID); err != nil {
		return false, err
	}
	//fmt.Println("entrptPass: ", entrptPass)

	//// ✅ 去除可能的空格或换行符
	//entrptPass = strings.TrimSpace(entrptPass)
	//password = strings.TrimSpace(password)
	// 检查密码是否正确
	err = bcrypt.CompareHashAndPassword([]byte(entrptPass), []byte(password))

	if err != nil {
		zap.L().Error("bcrypt.CompareHashAndPassword Failed")
		return false, err
	}
	return true, nil
}

// 修改密码
func ModifyPassword(userID int64, password string) (err error) {
	// 写入数据库
	encrypwd, err := encryptPassword(password)
	if err != nil {
		return
	}
	sqlStr := `UPDATE user SET password=? WHERE id=?`
	_, err = db.Exec(sqlStr, encrypwd, userID)
	return
}

// GetUserByName 通过用户名得到某个信息
func GetUserByName(username string) (user *models.UserSafe, err error) {
	user = new(models.UserSafe)
	sqlStr := `select user_id, username, COALESCE(email,'') as email, create_time, update_time, avatar_url, COALESCE(bio,'') as bio from user where username=?`
	err = db.Get(user, sqlStr, username)
	if err != nil {
		return nil, err
	}
	return
}

// GetUserByName 通过用户名得到匹配用户们的信息
func GetUsersByName(username string) ([]*models.UserSafe, error) {
	users := make([]*models.UserSafe, 0)
	query := "SELECT user_id, username, avatar_url FROM user WHERE username LIKE ? LIMIT 20"
	err := db.Select(&users, query, "%"+username+"%")
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetUsers 获得全部用户
func GetUsers() ([]*models.UserSafe, error) {
	users := make([]*models.UserSafe, 0)
	query := "SELECT user_id, username, avatar_url FROM user LIMIT 20"
	err := db.Select(&users, query)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetPostListByUserID 根据用户查询其全部帖子
func GetPostListByUserID(userID int64) (posts []*models.Post, err error) {
	posts = make([]*models.Post, 0)
	sqlStr := `SELECT post_id, community_id, status, title, content, create_time FROM post
				WHERE author_id = ?
				ORDER BY create_time DESC;`
	err = db.Select(&posts, sqlStr, userID)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

package logic

import (
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/models"
	"bluebell/pkg/jwt"
	"bluebell/pkg/snowflake"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"math/big"
	"time"
)

const (
	UserUpdateTime   = "update_time"
	DefaultAvatarURL = "bfkflsfksnfjke27fnjekrfnk" // 默认的用户头像URL
)

var (
	ErrorOPTExpired      = errors.New("验证码过期")
	ErrorOPTInvalid      = errors.New("验证码错误")
	ErrorModifyNil       = errors.New("不允许修改")
	ErrorPasswordInvalid = errors.New("密码错误")
	ErrorPassNotEqual    = errors.New("两次输入密码不同")
)

// JSON 字段到数据库字段的映射
var UserFieldMapping = map[string]string{
	"user_id":    "user_id",
	"username":   "username",
	"email":      "email",
	"avatar_url": "avatar_url",
	"bio":        "bio",
	"created_at": "create_time",
	"updated_at": "update_time",
}

// generateOTP 用户登录时生成6位数字验证码
func GenerateOTP(userID int64, c *gin.Context) (err error, otp string) {
	// 生成过程
	otp = ""
	for i := 0; i < 6; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(10))
		otp += fmt.Sprintf("%d", num)
	}
	// 存入 Redis（有效期 5 分钟）
	if err = redis.SetOTP(c, userID, otp); err != nil {
		zap.L().Error("redis StoreOTP error", zap.Error(err))
		return err, ""
	}

	return nil, otp
}

// GetExistUser 判断用户是否存在，返回id
func GetExistUser(req *models.ParamUsernameRequest) (int64, bool) {
	userID, err := mysql.GetExistUser(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			zap.L().Error("mysql.GetExistUser not exsit", zap.Error(err))
			return 0, false // 用户不存在
		}
		zap.L().Error("mysql.GetExistUser error", zap.Error(err))
		return 0, false
	}
	return userID, true
}

// Signup 用户注册的逻辑处理
func Signup(p *models.ParamSignUp) (err error) {
	// 判断用户是否存在
	if err = mysql.CheckUserExist(p.Username); err != nil {
		return err
	}
	// 2. 生成uid
	userID := snowflake.GenID()

	// 构造一个User实例
	user := &models.User{
		UserID:    userID,
		Username:  p.Username,
		Password:  p.Password,
		CreatedAt: time.Now(),
		AvatarURL: DefaultAvatarURL,
	}
	// 3. 保存入数据库
	return mysql.InsertUser(user)
}

// LogIn 用户登录的逻辑处理
func LogIn(c *gin.Context, p *models.ParamLogIn) (User *models.User, err error) {
	user := &models.User{
		Username: p.Username,
		Password: p.Password,
	}
	// 判断用户和密码是否正确(传递指针，拿到user)
	if err = mysql.CheckLogin(user); err != nil {
		return nil, err
	}
	// 判断验证码是否正确
	var otp string
	if otp, err = redis.GetOTP(c, user.UserID); err != nil {
		zap.L().Error("redis GetOTP error", zap.Error(err))
		return nil, ErrorOPTExpired
	}
	if otp != p.IdentifyCode {
		return nil, ErrorOPTInvalid
	}

	// 登录成功，生成JWT
	var token string
	token, err = jwt.GenToken(user.UserID, user.Username)
	if err != nil {
		return nil, err
	}
	user.Token = token
	return user, nil
}

func LogOut(token string) error {
	return mysql.AddToBlacklist(token)
}

// GetUserInfo 获取当前用户的信息
func GetUserInfo(userID int64) (*models.User, error) {
	return mysql.GetUserInfo(userID)
}

// ModifyUserInfo 修改用户基本信息
func ModifyUserInfo(userID int64, updates map[string]interface{}) error {
	// 过滤不允许更新的字段（如 user_id, created_at, token, password）
	disallowedFields := map[string]bool{
		"user_id": true, "created_at": true, "token": true, "password": true,
	}
	for key := range updates {
		if disallowedFields[key] {
			delete(updates, key) // 移除不允许修改的字段
		}
	}

	if len(updates) == 0 {
		zap.L().Error("存在不允许修改字段")
		return ErrorModifyNil
	}

	// 转换 JSON 字段为 DB 字段
	validUpdates := make(map[string]interface{})
	for jsonField, value := range updates {
		if dbField, exists := UserFieldMapping[jsonField]; exists {
			validUpdates[dbField] = value // 映射 JSON 字段到数据库字段
		}
	}

	// 修改更新信息时间
	validUpdates[UserUpdateTime] = time.Now()
	// 更新mysql数据库表
	if err := mysql.ModifyUserInfo(userID, validUpdates); err != nil {
		zap.L().Error("mysql.ModifyUserInfo error", zap.Error(err))
		return err
	}
	return nil
}

// ModifyPassword 修改用户密码
func ModifyPassword(userID int64, p *models.ParamPassword) error {
	// 判断原密码是否正确
	flag, err := mysql.CheckPassword(userID, p.OrPassword)
	if err != nil {
		zap.L().Error("mysql.CheckPassword error", zap.Error(err))
		return err
	}
	if !flag {
		zap.L().Error("the password is incorrect", zap.Error(ErrorOPTInvalid))
		return ErrorPasswordInvalid
	}

	// 判断新密码两次输入是否相等
	if p.Password != p.RePassword {
		zap.L().Error("the passwords are not equal", zap.Error(ErrorOPTInvalid))
		return ErrorPassNotEqual
	}
	// 更新密码
	if err := mysql.ModifyPassword(userID, p.Password); err != nil {
		zap.L().Error("mysql.ModifyPassword error", zap.Error(err))
		return err
	}
	return nil
}

func GetUserByName(username string) (*models.UserSafe, error) {
	user, err := mysql.GetUserByName(username)
	if err != nil {
		zap.L().Error("mysql.GetUserByName error", zap.Error(err))
		return nil, err
	}
	return user, nil
}

func GetUserByID(userID int64) (*models.UserSafe, error) {
	user, err := mysql.GetUserByID(userID)
	if err != nil {
		zap.L().Error("mysql.GetUserByName error", zap.Error(err))
		return nil, err
	}
	return user, nil
}

func GetUsersByName(username string) ([]*models.UserSafe, error) {
	users, err := mysql.GetUsersByName(username)
	if err != nil {
		zap.L().Error("mysql.GetUsersByName error", zap.Error(err))
		return nil, err
	}
	return users, nil
}

// GetUsers 获取全部的用户
func GetUsers() ([]*models.UserSafe, error) {
	users, err := mysql.GetUsers()
	if err != nil {
		zap.L().Error("mysql.GetUsers error", zap.Error(err))
		return nil, err
	}
	return users, nil
}

// GetUserPosts 获取用户以及其对应的全部帖子
func GetUserPosts(userID int64) (*models.UserPost, error) {
	// 获取用户
	user, err := GetUserByID(userID)
	if err != nil {
		zap.L().Error("mysql.GetUserByID error", zap.Error(err))
		return nil, err
	}
	// 获取帖子
	posts, err := mysql.GetPostListByUserID(userID)
	if err == sql.ErrNoRows {
		zap.L().Warn("mysql.GetPostListByUserID get nil", zap.Error(err))
		err = nil
	}
	if err != nil {
		zap.L().Error("mysql.GetPostListByUserID error", zap.Error(err))
		return nil, err
	}
	// 拼装得到的信息
	userpost := &models.UserPost{
		User:  user,
		Posts: posts,
	}
	return userpost, nil
}

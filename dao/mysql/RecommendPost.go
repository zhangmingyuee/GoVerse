package mysql

import (
	"bluebell/models"
	"fmt"
	"gorm.io/gorm"
)

// 获取所有用户-文章行为数据
func GetUserPostBehavior() ([]models.UserPostBehavior, error) {
	var behaviors []models.UserPostBehavior
	err := gormdb.Find(&behaviors).Error
	return behaviors, err
}

// CheckBehavior 新增用户帖子行为分数，返回查到的结构体
func CheckBehavior(userID, postId int64) (*models.UserPostBehavior, error) {
	// 使用 GORM 查询对应条件的记录
	var behavior models.UserPostBehavior
	err := gormdb.Model(&models.UserPostBehavior{}).Where("user_id = ? AND post_id = ?", userID, postId).First(&behavior).Error
	if err != nil {
		// 如果没有找到对应的记录或发生其他错误，返回错误
		return nil, err
	}

	// 返回查到的结构体
	return &behavior, nil
}

// CreateBehavior 创建用户帖子行为
func CreateBehavior(userID, postID int64, behaviorCode int8) error {
	fmt.Println(userID)
	// 创建新的用户帖子行为记录
	behavior := models.UserPostBehavior{
		UserID: userID,
		PostID: postID,
	}
	if behaviorCode == BehaviorBrowser {
		behavior.Browse = 1
		behavior.Rate = 1
	}
	if behaviorCode == BehaviorLike {
		behavior.Like = 1
		behavior.Rate = 3
	}
	if behaviorCode == BehaviorComment {
		behavior.Comment = 1
		behavior.Rate = 3
	}

	// 使用 GORM 的 Create 方法插入数据
	if err := gormdb.Create(&behavior).Error; err != nil {
		// 如果插入失败，返回错误
		return err
	}
	// 返回 nil 表示成功
	return nil
}

// UpdateBehavior 更新用户帖子行为
func UpdateBehavior(userID, postID int64, behaviorCode int8) error {
	rateIncr := 1
	behav := "browse"
	if behaviorCode == BehaviorComment {
		behav = "comment"
		rateIncr = 2
	} else if behaviorCode == BehaviorLike {
		behav = "like"
		rateIncr = 2
	}

	// 使用 gorm.Expr 构建 SQL 表达式，更新 rate 和对应行为字段
	updateData := map[string]interface{}{
		"rate":   gorm.Expr("rate + ?", rateIncr),
		"browse": 1,
		behav:    1, // 将对应行为字段设为 1
	}

	result := gormdb.Model(&models.UserPostBehavior{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Updates(updateData) // 使用 Updates 更新多个字段

	// 检查更新是否成功
	if err := result.Error; err != nil {
		// 如果发生错误，返回错误信息
		return err
	}

	// 成功时返回 nil
	return nil
}

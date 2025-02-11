package mysql

import (
	"bluebell/models"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var gormdb *gorm.DB // 全局 GORM 连接

// **初始化 GORM，复用 `sqlx` 连接**
func InitGORM() (err error) {
	if db == nil {
		zap.L().Error("sqlx DB is not initialized, GORM cannot be initialized")
		return
	}

	// 使用 `sqlx.DB` 作为 GORM 的连接
	gormdb, err = gorm.Open(mysql.New(mysql.Config{
		Conn: db.DB, // 复用 sqlx 的 `sql.DB`
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 关闭 GORM 日志
	})

	if err != nil {
		zap.L().Error("failed to initialize GORM", zap.Error(err))
		return
	}

	// 如果 jwt_blacklist 表不存在，它会自动创建一个 名为 jwt_blacklist 的表。
	// 如果表已存在，但结构不同，GORM 会尝试 更新表结构 以匹配 JWTBlacklist 结构体。
	// 如果表结构完全匹配，它不会执行任何操作
	if err = gormdb.AutoMigrate(&models.JWTBlacklist{}); err != nil {
		zap.L().Error("failed to auto migrate JWT Blacklist", zap.Error(err))
		return
	}

	zap.L().Info("GORM initialized successfully")
	return
}

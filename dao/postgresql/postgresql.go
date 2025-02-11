package postgresql

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var db *sqlx.DB

func Init() (err error) {
	// 组装 PostgreSQL 连接字符串
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("postgres.host"),
		viper.GetInt("postgres.port"),
		viper.GetString("postgres.user"),
		viper.GetString("postgres.password"),
		viper.GetString("postgres.dbname"),
		viper.GetString("postgres.sslmode"))

	db, err = sqlx.Connect("postgres", connStr)
	if err != nil {
		zap.L().Error("connect postgresql failed", zap.Error(err))
		return
	}
	zap.L().Info("connect psotgresql success")
	return nil
}

func Close() {
	if db != nil {
		db.Close()
	}
	//_ = db.Close()
}

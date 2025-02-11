package mysql

import (
	"bluebell/models"
	"database/sql"
	"errors"
	"go.uber.org/zap"
)

// GetCommunityList 查询全部社区
func GetCommunityList() ([]*models.Community, error) {
	sqlStr := "SELECT community_id, community_name FROM community"
	var communityList []*models.Community // 先声明切片变量
	err := db.Select(&communityList, sqlStr)
	if err != nil {
		zap.L().Error("query community list failed", zap.Error(err))
		return nil, err
	}

	if len(communityList) == 0 { // 查询结果为空时打印日志
		zap.L().Warn("there is no community in db")
	}

	return communityList, nil
}

// GetCommunityById 根据id查询社区详情
func GetCommunityById(id int64) (*models.CommunityDetail, error) {
	sqlStr := "SELECT community_id, community_name, introduction, create_time FROM community WHERE community_id = ?"

	var communityDetail models.CommunityDetail // 不要用指针的指针，直接定义结构体变量
	err := db.Get(&communityDetail, sqlStr, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // 使用 errors.Is 进行错误比较，兼容性更好
			zap.L().Warn("there is no community in db", zap.Int64("community_id", id))
			err = ErrorInvalidID
		}
		return nil, err // 其他 SQL 运行错误，直接返回
	}

	return &communityDetail, nil
}

// CreateCommunity 创建社区
func CreateCommunity(comm *models.CommunityDetail) error {
	sqlStr := "INSERT INTO community (community_id, community_name, introduction) VALUES (?, ?, ?)"
	_, err := db.Exec(sqlStr, comm.ID, comm.Name, comm.Introduction)
	if err != nil {
		return err
	}
	return nil
}

package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"fmt"
	"go.uber.org/zap"
)

// GetCommunity 获得全部的社区
func GetCommunity() ([]*models.Community, error) {
	// 查找全部的community并返回
	return mysql.GetCommunityList()

}

// GetCommunityDetail 根据社区id查询社区信息
func GetCommunityDetail(id int64) (*models.CommunityDetail, error) {
	return mysql.GetCommunityById(id)
}

// CreateCommunity 创建社区
func CreateCommunity(comm *models.CommunityDetail) error {
	fmt.Printf("id=%d, name=%s,intr=%s\n", comm.ID, comm.Name, comm.Introduction)
	if err := mysql.CreateCommunity(comm); err != nil {
		zap.L().Error("CreateCommunity Failed", zap.Error(err))
		return err
	}
	return nil
}

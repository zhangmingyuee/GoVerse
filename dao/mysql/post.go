package mysql

import (
	"bluebell/models"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"strings"
)

/*	一些代码小备注：
	db.Exec：执行 不返回结果的 SQL（INSERT/UPDATE/DELETE）。
	db.Get：查询 单行数据 并自动映射到结构体（SELECT）。
	db.Select：查询 多行数据 并自动映射到 []struct（SELECT）
*/

// CreatePost 向数据库插入一个帖子
func CreatePost(p *models.Post) (err error) {
	sqlStr := `insert into post (post_id, title, content, author_id, community_id) values (?,?,?,?,?)`
	_, err = db.Exec(sqlStr, p.ID, p.Title, p.Content, p.AuthorID, p.CommunityID)
	return
}

// GetPostByID 根据id查询单个帖子的详细信息
func GetPostByID(pid int64) (p *models.Post, err error) {
	p = new(models.Post)
	sqlStr := `select post_id, title, content, author_id, community_id, create_time from post where post_id = ?`
	err = db.Get(p, sqlStr, pid)
	return
}

// GetPostList 获得数据库中全部帖子信息
func GetPostList(offset, limit int64) ([]*models.Post, error) {
	posts := make([]*models.Post, 0)
	sqlStr := `select post_id, title, content, author_id, community_id, status, create_time from post order by create_time desc limit ?, ?`
	err := db.Select(&posts, sqlStr, ((offset - 1) * limit), limit)
	//fmt.Println("len", len(posts))
	if err != nil {
		zap.L().Error("mysql.GetPostList failed", zap.Error(err))
		return nil, err
	}
	if len(posts) == 0 {
		zap.L().Warn("there is no post in db")
		return nil, nil
	}
	return posts, err
}

// 根据给定的id列表查询帖子数据
func GetPostsListByIds(ids []string) (postlist []*models.Post, err error) {
	sqlStr := `select post_id, title, content, author_id, community_id, create_time from post where post_id in (?)
				order by FIND_IN_SET(post_id, ?)`
	query, args, err := sqlx.In(sqlStr, ids, strings.Join(ids, ","))
	if err != nil {
		return nil, err
	}
	query = db.Rebind(query)

	err = db.Select(&postlist, query, args...)
	return
}

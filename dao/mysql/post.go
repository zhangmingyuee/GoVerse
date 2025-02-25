package mysql

import (
	"bluebell/models"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"log"
	"strings"
	"time"
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

// GetPostsListByIds 根据给定的 ID 列表查询帖子数据，支持传入 int64 切片
func GetPostsListByInt64Ids(ids []int64) (postlist []*models.Post, err error) {
	// 定义 SQL 查询语句，使用 FIND_IN_SET 进行排序
	sqlStr := `select post_id, title, content, author_id, community_id, create_time from post where post_id in (?) 
				order by FIND_IN_SET(post_id, ?)`

	// 使用 sqlx.In 将切片参数绑定到查询语句中
	query, args, err := sqlx.In(sqlStr, ids, strings.Join(toStringSlice(ids), ","))
	if err != nil {
		return nil, err
	}
	// Rebind 查询语句，使其兼容不同数据库驱动的占位符样式（如 ? 或 $1, $2）
	query = db.Rebind(query)
	// 执行查询并将结果映射到 postlist
	err = db.Select(&postlist, query, args...)
	return
}

// 将 []int64 转换为 []string
func toStringSlice(ids []int64) []string {
	result := make([]string, len(ids))
	for i, v := range ids {
		result[i] = fmt.Sprintf("%d", v)
	}
	return result
}

// 根据给定的id列表查询对应社区的数据
func GetPostsListByIdsAndComm(comm_id int64, ids []string) (postlist []*models.Post, err error) {
	sqlStr := `SELECT post_id, title, content, author_id, community_id, create_time
				FROM post
				WHERE post_id IN (?) AND community_id = ?
				ORDER BY FIND_IN_SET(post_id, ?);`
	query, args, err := sqlx.In(sqlStr, ids, comm_id, strings.Join(ids, ","))
	if err != nil {
		return nil, err
	}
	query = db.Rebind(query)

	err = db.Select(&postlist, query, args...)
	return
}

// 按照时间顺序查询帖子
func GetPostIdsInTime(p *models.ParamPostList) (post []*models.Post, err error) {
	post = make([]*models.Post, 0)
	sqlStr := `SELECT post_id, title, content, author_id, community_id, create_time
				FROM post
				ORDER BY create_time DESC
				LIMIT ?,?;`
	err = db.Select(&post, sqlStr, (p.Offset-1)*p.Limit, p.Limit)
	return
}

// 按照时间和社区查询帖子
func GetPostIdsInCommTime(p *models.ParamPostList) (post []*models.Post, err error) {
	post = make([]*models.Post, 0)
	sqlStr := `SELECT post_id, title, content, author_id, community_id, create_time
				FROM post
				WHERE community_id = ? 
				ORDER BY create_time DESC
				LIMIT ?,?;`
	err = db.Select(&post, sqlStr, p.Community_id, (p.Offset-1)*p.Limit, p.Limit)
	return
}

// 查询数据库获得对应帖子的创建时间
func GetPostCreateTime(postID int64) (ctimestamp int64, err error) {
	var ct time.Time
	sqlStr := `select create_time from post where post_id = ?`
	err = db.Get(&ct, sqlStr, postID)
	if err != nil {
		return 0, err
	}
	ctimestamp = ct.Unix()
	return ctimestamp, nil
}

// 开始MySQL事务，并将数据存入mysql数据库
func BatchInsertPostHotScores(postUpdateScores []redis.Z) error {
	// 开启事务
	tx, err := db.Beginx()
	if err != nil {
		zap.L().Error("开启mysql事务失败", zap.Error(err))
		return err
	}

	// 构造批量插入 SQL
	stmt := "INSERT INTO post_hot_scores (post_id, hot_score, updated_at) VALUES "
	values := []interface{}{}
	placeholders := []string{}

	for _, post := range postUpdateScores {
		postID := post.Member.(string) // 获取帖子的 ID
		hotScore := post.Score         // 获取热度
		placeholders = append(placeholders, "(?, ?, NOW())")
		values = append(values, postID, hotScore)
	}

	// 拼接 SQL 语句
	stmt += strings.Join(placeholders, ",")
	stmt += " ON DUPLICATE KEY UPDATE hot_score = VALUES(hot_score), updated_at = NOW();"

	// 执行批量插入
	_, err = tx.Exec(stmt, values...)
	if err != nil {
		zap.L().Error("批量插入MySQL失败", zap.Error(err))
		tx.Rollback()
		return err
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		zap.L().Error("提交MySQL事务失败", zap.Error(err))
		return err
	}
	return nil
}

// BatchInsertLikes 开始MySQL事务，并将更新的点赞数据存入mysql数据库
func BatchInsertLikes(postUpdateScores []redis.Z) error {
	// 开启事务
	tx, err := db.Beginx()
	if err != nil {
		zap.L().Error("开启mysql事务失败", zap.Error(err))
		return err
	}

	// 构造批量插入 SQL
	// 1️⃣ 预编译 SQL 语句
	stmt, err := tx.Preparex("UPDATE post SET likes = ? WHERE post_id = ?")
	if err != nil {
		zap.L().Error("SQL预编译失败", zap.Error(err))
		tx.Rollback() // 事务回滚
		return err
	}
	defer stmt.Close()

	// 2️⃣ 批量执行更新
	for _, post := range postUpdateScores {
		postID, ok := post.Member.(string) // Redis ZSET Member 可能是 string
		if !ok {
			zap.L().Warn("跳过无效的postID", zap.Any("post", post))
			continue
		}

		_, err := stmt.Exec(post.Score, postID)
		if err != nil {
			zap.L().Error("更新帖子点赞数失败", zap.String("postID", postID), zap.Error(err))
			tx.Rollback() // 事务回滚
			return err
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		zap.L().Error("提交MySQL事务失败", zap.Error(err))
		return err
	}
	return nil
}

// BatchInsertDisLikes 开始MySQL事务，并将更新的点赞数据存入mysql数据库
func BatchInsertDisLikes(postUpdateScores []redis.Z) error {
	// 开启事务
	tx, err := db.Beginx()
	if err != nil {
		zap.L().Error("开启mysql事务失败", zap.Error(err))
		return err
	}

	// 构造批量插入 SQL
	// 1️⃣ 预编译 SQL 语句
	stmt, err := tx.Preparex("UPDATE post SET dislikes = ? WHERE post_id = ?")
	if err != nil {
		zap.L().Error("SQL预编译失败", zap.Error(err))
		tx.Rollback() // 事务回滚
		return err
	}
	defer stmt.Close()

	// 2️⃣ 批量执行更新
	for _, post := range postUpdateScores {
		postID, ok := post.Member.(string) // Redis ZSET Member 可能是 string
		if !ok {
			zap.L().Warn("跳过无效的postID", zap.Any("post", post))
			continue
		}

		_, err := stmt.Exec(post.Score, postID)
		if err != nil {
			zap.L().Error("更新帖子点踩数失败", zap.String("postID", postID), zap.Error(err))
			tx.Rollback() // 事务回滚
			return err
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		zap.L().Error("提交MySQL事务失败", zap.Error(err))
		return err
	}
	return nil
}

// SaveImageToDB 插入图像信息
func SaveImageToDB(image *models.ParamImage) error {
	// 使用 GORM 的 Create 方法插入数据
	if err := gormdb.Create(image).Error; err != nil {
		log.Println("Failed to insert image:", err)
		return err
	}
	return nil
}

// 查询数据库获得对应帖子的作者
func GetPostAuthor(postID int64) (int64, error) {
	var userID int64
	sqlStr := `select author_id from post where post_id = ?`
	err := db.Get(&userID, sqlStr, postID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

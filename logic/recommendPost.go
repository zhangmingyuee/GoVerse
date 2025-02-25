package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"go.uber.org/zap"
	"math"
	"sort"
)

// 排序的结构体类型，用于存储 map 的键值对
type ByValue []struct {
	Key   int64
	Value float64
}

func (a ByValue) Len() int           { return len(a) }
func (a ByValue) Less(i, j int) bool { return a[i].Value > a[j].Value } // 降序排序
func (a ByValue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func GetTopK(similaruser map[int64]float64, k int) ([]int64, []float64) {
	// 将 map 转换为切片
	var kvPairs []struct {
		Key   int64
		Value float64
	}
	for key, value := range similaruser {
		kvPairs = append(kvPairs, struct {
			Key   int64
			Value float64
		}{Key: key, Value: value})
	}

	// 按照 Value（相似度）进行排序
	sort.Sort(ByValue(kvPairs))

	// 获取前 k 个
	if k > len(kvPairs) {
		k = len(kvPairs) // 防止 k 大于总数
	}

	// 提取前 k 个用户ID和相似度
	var topKUsers []int64
	var topKSimilarities []float64
	for i := 0; i < k; i++ {
		topKUsers = append(topKUsers, kvPairs[i].Key)
		topKSimilarities = append(topKSimilarities, kvPairs[i].Value)
	}

	return topKUsers, topKSimilarities
}

// cosineSimilarity 计算两个帖子之间的余弦相似度
// vec1 和 vec2 是两个物品的评分向量--> 例如用户对这些物品的评分），通过余弦相似度计算它们之间的相似度
func cosineSimilarity(vec1, vec2 map[int64]int8) float64 {
	var dotProduct, normVec1, normVec2 float64

	// 计算点积（dot product）
	// 对于两个物品的评分向量，计算相同用户的评分乘积并累加，得到点积
	for pid, count1 := range vec1 {
		count2, exists := vec2[pid] // 如果物品2的评分向量也有该用户的评分
		if exists {
			dotProduct += float64(count1 * count2) // 累加两个物品在相同用户上的评分乘积
		}
	}

	// 计算第两个帖子（vec1）的模长（norm）
	for pid, count1 := range vec1 {
		normVec1 += float64(count1 * count1) // 累加每个评分的平方
		if count2, exist := vec2[pid]; exist {
			normVec2 += float64(count2 * count1)
		}
	}

	// 返回余弦相似度：dotProduct / (normVec1 * normVec2)
	// 余弦相似度是通过点积除以两个物品的模长的乘积得到的，结果范围为 [-1, 1]，值越大表示越相似
	return dotProduct / (math.Sqrt(normVec1) * math.Sqrt(normVec2))
}

// RecommendArticles 根据用户的历史行为推荐文章
// 根据物品（文章）之间的相似度以及用户的历史评分，向用户推荐物品
func RecommendArticles(userID int64) ([]*models.Post, error) {
	// 获取用户行为数据：用户对物品（文章）的评分或行为（如点赞）
	behaviors, err := mysql.GetUserPostBehavior()
	if err != nil {
		zap.L().Error("mysql.GetUserPostBehavior", zap.Error(err))
		return nil, err // 如果获取行为数据失败，返回错误
	}
	if len(behaviors) == 0 {
		zap.L().Warn("behaviors are empty", zap.Error(err))
		// 获取全部帖子
		return mysql.GetPostList(1, 10)
	}

	// 构建用户-文章评分矩阵：userItemMatrix[userID][articleID] = rate
	// 用来存储每个用户对每篇文章的评分
	userItemMatrix := make(map[int64]map[int64]int8)
	for _, behavior := range behaviors {
		if _, exists := userItemMatrix[behavior.UserID]; !exists {
			userItemMatrix[behavior.UserID] = make(map[int64]int8) // 初始化每个用户的评分矩阵
		}
		userItemMatrix[behavior.UserID][behavior.PostID] = behavior.Rate // 记录用户对文章的评分
	}

	// 其他用户与该用户行为的相似度
	similaruser := make(map[int64]float64)
	for uID, upost := range userItemMatrix {
		if userID == uID {
			continue
		}
		similaruser[uID] = cosineSimilarity(userItemMatrix[userID], upost)
	}

	// 找出最相似的k个用户（k=10）
	kuser := 10
	users, similar := GetTopK(similaruser, kuser)

	// 计算最终的评分
	postsimilar := make(map[int64]float64)
	for i, user := range users {
		for pid, rate := range userItemMatrix[user] {
			postsimilar[pid] += float64(rate) * similar[i]
		}
	}

	kpost := 10
	pos, _ := GetTopK(postsimilar, kpost)
	postss, err := mysql.GetPostsListByInt64Ids(pos)
	if err != nil {
		zap.L().Error("mysql.GetPostsListByInt64Ids", zap.Error(err))
		return nil, err
	}
	return postss, nil

}

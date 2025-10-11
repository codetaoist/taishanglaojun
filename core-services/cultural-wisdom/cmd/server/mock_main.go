package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CulturalWisdom 文化智慧内容模型
type CulturalWisdom struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Summary       string    `json:"summary"`
	Author        string    `json:"author"`
	Category      string    `json:"category"`
	School        string    `json:"school"`
	Tags          []string  `json:"tags"`
	Difficulty    string    `json:"difficulty"`
	ViewCount     int64     `json:"view_count"`
	LikeCount     int64     `json:"like_count"`
	ShareCount    int64     `json:"share_count"`
	CommentCount  int64     `json:"comment_count"`
	IsFeatured    bool      `json:"is_featured"`
	IsRecommended bool      `json:"is_recommended"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Category 分类模型
type Category struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    *int   `json:"parent_id"`
	SortOrder   int    `json:"sort_order"`
	IsActive    bool   `json:"is_active"`
}

// WisdomTag 智慧标签模型
type WisdomTag struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	UsageCount  int    `json:"usage_count"`
	IsActive    bool   `json:"is_active"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Wisdoms    []CulturalWisdom `json:"wisdoms"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// 模拟数据
var wisdoms []CulturalWisdom
var categories []Category
var tags []WisdomTag

func initMockData() {
	// 初始化分类数�?
	categories = []Category{
		{ID: 1, Name: "儒家", Description: "儒家思想和经�?, SortOrder: 1, IsActive: true},
		{ID: 2, Name: "道家", Description: "道家思想和经�?, SortOrder: 2, IsActive: true},
		{ID: 3, Name: "佛家", Description: "佛家思想和经�?, SortOrder: 3, IsActive: true},
		{ID: 4, Name: "法家", Description: "法家思想和经�?, SortOrder: 4, IsActive: true},
		{ID: 5, Name: "兵家", Description: "兵家思想和经�?, SortOrder: 5, IsActive: true},
	}

	// 初始化标签数�?
	tags = []WisdomTag{
		{ID: 1, Name: "修身", Description: "个人修养", Color: "#007bff", UsageCount: 15, IsActive: true},
		{ID: 2, Name: "治国", Description: "国家治理", Color: "#28a745", UsageCount: 12, IsActive: true},
		{ID: 3, Name: "齐家", Description: "家庭和睦", Color: "#ffc107", UsageCount: 8, IsActive: true},
		{ID: 4, Name: "平天�?, Description: "天下太平", Color: "#dc3545", UsageCount: 6, IsActive: true},
		{ID: 5, Name: "智慧", Description: "人生智慧", Color: "#6f42c1", UsageCount: 20, IsActive: true},
	}

	// 初始化智慧内容数�?
	wisdoms = []CulturalWisdom{
		{
			ID:            uuid.New().String(),
			Title:         "学而时习之，不亦说�?,
			Content:       "子曰：学而时习之，不亦说乎？有朋自远方来，不亦乐乎？人不知而不愠，不亦君子乎？",
			Summary:       "孔子论学习的快乐和君子的品格",
			Author:        "孔子",
			Category:      "儒家",
			School:        "儒家",
			Tags:          []string{"修身", "学习", "君子"},
			Difficulty:    "初级",
			ViewCount:     1250,
			LikeCount:     89,
			ShareCount:    23,
			CommentCount:  15,
			IsFeatured:    true,
			IsRecommended: true,
			CreatedAt:     time.Now().Add(-time.Hour * 24 * 30),
			UpdatedAt:     time.Now().Add(-time.Hour * 24 * 5),
		},
		{
			ID:            uuid.New().String(),
			Title:         "道可道，非常�?,
			Content:       "道可道，非常道；名可名，非常名。无名天地之始，有名万物之母�?,
			Summary:       "老子论道的本质和万物的起�?,
			Author:        "老子",
			Category:      "道家",
			School:        "道家",
			Tags:          []string{"�?, "哲学", "宇宙�?},
			Difficulty:    "高级",
			ViewCount:     980,
			LikeCount:     67,
			ShareCount:    18,
			CommentCount:  12,
			IsFeatured:    true,
			IsRecommended: false,
			CreatedAt:     time.Now().Add(-time.Hour * 24 * 25),
			UpdatedAt:     time.Now().Add(-time.Hour * 24 * 3),
		},
		{
			ID:            uuid.New().String(),
			Title:         "四谛八正�?,
			Content:       "苦集灭道四谛，正见正思维正语正业正命正精进正念正定八正道�?,
			Summary:       "佛陀教导的解脱之�?,
			Author:        "释迦牟尼",
			Category:      "佛家",
			School:        "佛家",
			Tags:          []string{"解脱", "修行", "智慧"},
			Difficulty:    "中级",
			ViewCount:     756,
			LikeCount:     45,
			ShareCount:    12,
			CommentCount:  8,
			IsFeatured:    false,
			IsRecommended: true,
			CreatedAt:     time.Now().Add(-time.Hour * 24 * 20),
			UpdatedAt:     time.Now().Add(-time.Hour * 24 * 2),
		},
		{
			ID:            uuid.New().String(),
			Title:         "法不阿贵，绳不挠�?,
			Content:       "法不阿贵，绳不挠曲。法之所加，智者弗能辞，勇者弗敢争。刑过不避大臣，赏善不遗匹夫�?,
			Summary:       "韩非子论法治的公正�?,
			Author:        "韩非�?,
			Category:      "法家",
			School:        "法家",
			Tags:          []string{"法治", "公正", "治国"},
			Difficulty:    "中级",
			ViewCount:     623,
			LikeCount:     38,
			ShareCount:    9,
			CommentCount:  6,
			IsFeatured:    false,
			IsRecommended: false,
			CreatedAt:     time.Now().Add(-time.Hour * 24 * 15),
			UpdatedAt:     time.Now().Add(-time.Hour * 24 * 1),
		},
		{
			ID:            uuid.New().String(),
			Title:         "知己知彼，百战不�?,
			Content:       "知己知彼，百战不殆；不知彼而知己，一胜一负；不知彼不知己，每战必败�?,
			Summary:       "孙子兵法中的战略智慧",
			Author:        "孙武",
			Category:      "兵家",
			School:        "兵家",
			Tags:          []string{"战略", "智慧", "谋略"},
			Difficulty:    "初级",
			ViewCount:     1456,
			LikeCount:     102,
			ShareCount:    34,
			CommentCount:  21,
			IsFeatured:    true,
			IsRecommended: true,
			CreatedAt:     time.Now().Add(-time.Hour * 24 * 10),
			UpdatedAt:     time.Now().Add(-time.Hour),
		},
	}
}

// 获取智慧列表
func getWisdomList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	category := c.Query("category")
	school := c.Query("school")
	difficulty := c.Query("difficulty")

	filteredWisdoms := make([]CulturalWisdom, 0)
	for _, wisdom := range wisdoms {
		if category != "" && wisdom.Category != category {
			continue
		}
		if school != "" && wisdom.School != school {
			continue
		}
		if difficulty != "" && wisdom.Difficulty != difficulty {
			continue
		}
		filteredWisdoms = append(filteredWisdoms, wisdom)
	}

	total := int64(len(filteredWisdoms))
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(filteredWisdoms) {
		end = len(filteredWisdoms)
	}
	if start > len(filteredWisdoms) {
		start = len(filteredWisdoms)
	}

	result := SearchResult{
		Wisdoms:    filteredWisdoms[start:end],
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取智慧列表成功",
		"data":    result,
	})
}

// 获取智慧详情
func getWisdomDetail(c *gin.Context) {
	id := c.Param("id")
	
	for i, wisdom := range wisdoms {
		if wisdom.ID == id {
			// 增加浏览�?
			wisdoms[i].ViewCount++
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "获取智慧详情成功",
				"data":    wisdoms[i],
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"code":    404,
		"message": "智慧内容不存�?,
		"data":    nil,
	})
}

// 搜索智慧内容
func searchWisdom(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	filteredWisdoms := make([]CulturalWisdom, 0)
	for _, wisdom := range wisdoms {
		if keyword == "" || 
		   strings.Contains(wisdom.Title, keyword) || 
		   strings.Contains(wisdom.Content, keyword) || 
		   strings.Contains(wisdom.Summary, keyword) ||
		   strings.Contains(wisdom.Author, keyword) {
			filteredWisdoms = append(filteredWisdoms, wisdom)
		}
	}

	total := int64(len(filteredWisdoms))
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(filteredWisdoms) {
		end = len(filteredWisdoms)
	}
	if start > len(filteredWisdoms) {
		start = len(filteredWisdoms)
	}

	result := SearchResult{
		Wisdoms:    filteredWisdoms[start:end],
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "搜索成功",
		"data":    result,
	})
}

// 获取分类列表
func getCategories(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取分类列表成功",
		"data":    categories,
	})
}

// 获取标签列表
func getTags(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取标签列表成功",
		"data":    tags,
	})
}

// 获取推荐内容
func getRecommendations(c *gin.Context) {
	userID := c.Param("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	// 简单推荐逻辑：返回推荐标记的内容
	recommendations := make([]CulturalWisdom, 0)
	for _, wisdom := range wisdoms {
		if wisdom.IsRecommended && len(recommendations) < limit {
			recommendations = append(recommendations, wisdom)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取推荐内容成功",
		"data": gin.H{
			"user_id":        userID,
			"recommendations": recommendations,
			"total":          len(recommendations),
		},
	})
}

// 获取统计信息
func getWisdomStats(c *gin.Context) {
	totalWisdoms := len(wisdoms)
	totalCategories := len(categories)
	totalTags := len(tags)
	
	var totalViews, totalLikes, totalShares int64
	for _, wisdom := range wisdoms {
		totalViews += wisdom.ViewCount
		totalLikes += wisdom.LikeCount
		totalShares += wisdom.ShareCount
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取统计信息成功",
		"data": gin.H{
			"total_wisdoms":    totalWisdoms,
			"total_categories": totalCategories,
			"total_tags":       totalTags,
			"total_views":      totalViews,
			"total_likes":      totalLikes,
			"total_shares":     totalShares,
		},
	})
}

func main() {
	// 初始化模拟数�?
	initMockData()

	// 创建Gin路由�?
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API路由�?
	api := r.Group("/api/v1")
	{
		// 智慧内容相关路由
		api.GET("/cultural-wisdom/list", getWisdomList)
		api.GET("/cultural-wisdom/detail/:id", getWisdomDetail)
		api.GET("/cultural-wisdom/search", searchWisdom)
		api.GET("/cultural-wisdom/stats", getWisdomStats)
		
		// 分类和标签路�?
		api.GET("/cultural-wisdom/categories", getCategories)
		api.GET("/cultural-wisdom/tags", getTags)
		
		// 推荐路由
		api.GET("/cultural-wisdom/recommendations/:user_id", getRecommendations)
	}

	// 健康检�?
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "cultural-wisdom",
			"version": "1.0.0",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	log.Println("文化智慧服务 (Mock版本) 启动在端�?8082")
	log.Fatal(r.Run(":8082"))
}

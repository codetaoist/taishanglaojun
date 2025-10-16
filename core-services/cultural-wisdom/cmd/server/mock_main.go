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

// CulturalWisdom 
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

// Category 
type Category struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    *int   `json:"parent_id"`
	SortOrder   int    `json:"sort_order"`
	IsActive    bool   `json:"is_active"`
}

// WisdomTag 
type WisdomTag struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	UsageCount  int    `json:"usage_count"`
	IsActive    bool   `json:"is_active"`
}

// SearchResult 
type SearchResult struct {
	Wisdoms    []CulturalWisdom `json:"wisdoms"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// 
var wisdoms []CulturalWisdom
var categories []Category
var tags []WisdomTag

func initMockData() {
	// 
	categories = []Category{
		{ID: 1, Name: "", Description: "", SortOrder: 1, IsActive: true},
		{ID: 2, Name: "", Description: "", SortOrder: 2, IsActive: true},
		{ID: 3, Name: "", Description: "", SortOrder: 3, IsActive: true},
		{ID: 4, Name: "", Description: "", SortOrder: 4, IsActive: true},
		{ID: 5, Name: "", Description: "", SortOrder: 5, IsActive: true},
	}

	// 
	tags = []WisdomTag{
		{ID: 1, Name: "", Description: "", Color: "#007bff", UsageCount: 15, IsActive: true},
		{ID: 2, Name: "", Description: "", Color: "#28a745", UsageCount: 12, IsActive: true},
		{ID: 3, Name: "", Description: "", Color: "#ffc107", UsageCount: 8, IsActive: true},
		{ID: 4, Name: "", Description: "", Color: "#dc3545", UsageCount: 6, IsActive: true},
		{ID: 5, Name: "", Description: "", Color: "#6f42c1", UsageCount: 20, IsActive: true},
	}

	// 
	wisdoms = []CulturalWisdom{
		{
			ID:            uuid.New().String(),
			Title:         "",
			Content:       "㳲",
			Summary:       "",
			Author:        "",
			Category:      "",
			School:        "",
			Tags:          []string{"", "", ""},
			Difficulty:    "",
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
			Title:         "",
			Content:       "",
			Summary:       "",
			Author:        "",
			Category:      "",
			School:        "",
			Tags:          []string{"", "", ""},
			Difficulty:    "",
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
			Title:         "",
			Content:       "",
			Summary:       "",
			Author:        "",
			Category:      "",
			School:        "",
			Tags:          []string{"", "", ""},
			Difficulty:    "",
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
			Title:         "",
			Content:       "",
			Summary:       "",
			Author:        "",
			Category:      "",
			School:        "",
			Tags:          []string{"", "", ""},
			Difficulty:    "",
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
			Title:         "",
			Content:       "",
			Summary:       "",
			Author:        "",
			Category:      "",
			School:        "",
			Tags:          []string{"", "", ""},
			Difficulty:    "",
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

// 
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
		"message": "",
		"data":    result,
	})
}

// 
func getWisdomDetail(c *gin.Context) {
	id := c.Param("id")

	for i, wisdom := range wisdoms {
		if wisdom.ID == id {
			// 
			wisdoms[i].ViewCount++
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "",
				"data":    wisdoms[i],
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"code":    404,
		"message": "",
		"data":    nil,
	})
}

// 
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
		"message": "",
		"data":    result,
	})
}

// 
func getCategories(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "",
		"data":    categories,
	})
}

// 
func getTags(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "",
		"data":    tags,
	})
}

// 
func getRecommendations(c *gin.Context) {
	userID := c.Param("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	// 
	recommendations := make([]CulturalWisdom, 0)
	for _, wisdom := range wisdoms {
		if wisdom.IsRecommended && len(recommendations) < limit {
			recommendations = append(recommendations, wisdom)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "",
		"data": gin.H{
			"user_id":         userID,
			"recommendations": recommendations,
			"total":           len(recommendations),
		},
	})
}

// 
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
		"message": "",
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
	// 
	initMockData()

	// Gin
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API
	api := r.Group("/api/v1")
	{
		// 
		api.GET("/cultural-wisdom/list", getWisdomList)
		api.GET("/cultural-wisdom/detail/:id", getWisdomDetail)
		api.GET("/cultural-wisdom/search", searchWisdom)
		api.GET("/cultural-wisdom/stats", getWisdomStats)

		// 
		api.GET("/cultural-wisdom/categories", getCategories)
		api.GET("/cultural-wisdom/tags", getTags)

		// 
		api.GET("/cultural-wisdom/recommendations/:user_id", getRecommendations)
	}

	// 
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "cultural-wisdom",
			"version": "1.0.0",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	log.Println(" (Mock汾)  8082")
	log.Fatal(r.Run(":8082"))
}


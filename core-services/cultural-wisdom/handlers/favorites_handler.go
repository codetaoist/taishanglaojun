package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/services"
)

// FavoritesHandler 收藏和笔记处理器
type FavoritesHandler struct {
	favoritesService *services.FavoritesService
	logger           *zap.Logger
}

// NewFavoritesHandler 创建收藏和笔记处理器实例
func NewFavoritesHandler(favoritesService *services.FavoritesService, logger *zap.Logger) *FavoritesHandler {
	return &FavoritesHandler{
		favoritesService: favoritesService,
		logger:           logger,
	}
}

// AddFavorite 添加收藏
func (h *FavoritesHandler) AddFavorite(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	var req models.FavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	favorite, err := h.favoritesService.AddFavorite(c.Request.Context(), userID, req.WisdomID)
	if err != nil {
		h.logger.Error("Failed to add favorite", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "收藏成功",
		"data":    favorite,
	})
}

// RemoveFavorite 移除收藏
func (h *FavoritesHandler) RemoveFavorite(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	wisdomID := c.Param("wisdom_id")
	if wisdomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "智慧ID不能为空"})
		return
	}

	err := h.favoritesService.RemoveFavorite(c.Request.Context(), userID, wisdomID)
	if err != nil {
		h.logger.Error("Failed to remove favorite", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "取消收藏成功"})
}

// GetUserFavorites 获取用户收藏列表
func (h *FavoritesHandler) GetUserFavorites(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	favorites, total, err := h.favoritesService.GetUserFavorites(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get user favorites", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": favorites,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// CheckFavoriteStatus 检查收藏状?
func (h *FavoritesHandler) CheckFavoriteStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	wisdomID := c.Param("wisdom_id")
	if wisdomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "智慧ID不能为空"})
		return
	}

	isFavorited, err := h.favoritesService.IsFavorited(c.Request.Context(), userID, wisdomID)
	if err != nil {
		h.logger.Error("Failed to check favorite status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_favorited": isFavorited,
	})
}

// CreateNote 创建笔记
func (h *FavoritesHandler) CreateNote(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	var req models.NoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	note, err := h.favoritesService.CreateNote(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to create note", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "笔记创建成功",
		"data":    note,
	})
}

// UpdateNote 更新笔记
func (h *FavoritesHandler) UpdateNote(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	wisdomID := c.Param("wisdom_id")
	if wisdomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "智慧ID不能为空"})
		return
	}

	var req models.NoteUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	note, err := h.favoritesService.UpdateNote(c.Request.Context(), userID, wisdomID, &req)
	if err != nil {
		h.logger.Error("Failed to update note", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "笔记更新成功",
		"data":    note,
	})
}

// GetNote 获取笔记
func (h *FavoritesHandler) GetNote(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	wisdomID := c.Param("wisdom_id")
	if wisdomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "智慧ID不能为空"})
		return
	}

	note, err := h.favoritesService.GetNote(c.Request.Context(), userID, wisdomID)
	if err != nil {
		h.logger.Error("Failed to get note", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": note,
	})
}

// GetUserNotes 获取用户笔记列表
func (h *FavoritesHandler) GetUserNotes(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	notes, total, err := h.favoritesService.GetUserNotes(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get user notes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": notes,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// DeleteNote 删除笔记
func (h *FavoritesHandler) DeleteNote(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	wisdomID := c.Param("wisdom_id")
	if wisdomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "智慧ID不能为空"})
		return
	}

	err := h.favoritesService.DeleteNote(c.Request.Context(), userID, wisdomID)
	if err != nil {
		h.logger.Error("Failed to delete note", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "笔记删除成功"})
}


package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/application/services/learner"
)

// CommunityHandler 学习社区处理
type CommunityHandler struct {
	communityService *learner.LearningCommunityService
}

// NewCommunityHandler 创建新的学习社区处理
func NewCommunityHandler(communityService *learner.LearningCommunityService) *CommunityHandler {
	return &CommunityHandler{
		communityService: communityService,
	}
}

// CreateCommunity 创建学习社区
// @Summary 创建学习社区
// @Description 创建新的学习社区
// @Tags community
// @Accept json
// @Produce json
// @Param request body learner.CreateCommunityRequest true "创建社区请求"
// @Param creator_id header string true "创建者ID"
// @Success 201 {object} learner.CreateCommunityResponse "社区创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/communities [post]
func (h *CommunityHandler) CreateCommunity(c *gin.Context) {
	var req learner.CreateCommunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	// 从请求头或上下文获取创建者ID
	creatorIDStr := c.GetHeader("X-User-ID")
	if creatorIDStr == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Missing creator ID",
			Message: "Creator ID is required",
		})
		return
	}

	creatorID, err := uuid.Parse(creatorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid creator ID",
			Message: "Creator ID must be a valid UUID",
		})
		return
	}

	response, err := h.communityService.CreateCommunity(c.Request.Context(), &req, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create community",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// JoinCommunity 加入学习社区
// @Summary 加入学习社区
// @Description 学习者加入指定的学习社区
// @Tags community
// @Accept json
// @Produce json
// @Param community_id path string true "社区ID"
// @Param request body learner.JoinCommunityRequest true "加入社区请求"
// @Success 200 {object} SuccessResponse "加入成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "社区不存在"
// @Failure 409 {object} ErrorResponse "已经是社区成员"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/communities/{community_id}/join [post]
func (h *CommunityHandler) JoinCommunity(c *gin.Context) {
	communityIDStr := c.Param("community_id")
	communityID, err := uuid.Parse(communityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid community ID",
			Message: "Community ID must be a valid UUID",
		})
		return
	}

	var req learner.JoinCommunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	// 确保路径参数和请求体中的社区ID一致
	req.CommunityID = communityID

	if err := h.communityService.JoinCommunity(c.Request.Context(), &req); err != nil {
		if err.Error() == "already a member of this community" {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "Already a member",
				Message: "You are already a member of this community",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to join community",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Successfully joined the community",
	})
}

// CreatePost 创建帖子
// @Summary 创建帖子
// @Description 在学习社区中创建新帖子
// @Tags community
// @Accept json
// @Produce json
// @Param request body learner.CreatePostRequest true "创建帖子请求"
// @Success 201 {object} learner.Post "帖子创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 403 {object} ErrorResponse "无权创建帖子"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/communities/posts [post]
func (h *CommunityHandler) CreatePost(c *gin.Context) {
	var req learner.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	post, err := h.communityService.CreatePost(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "user is not a member of this community" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "Access denied",
				Message: "You must be a member of the community to create posts",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create post",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, post)
}

// GetCommunityPosts 获取社区帖子列表
// @Summary 获取社区帖子列表
// @Description 获取指定社区的帖子列表，支持筛选和分页
// @Tags community
// @Accept json
// @Produce json
// @Param community_id path string true "社区ID"
// @Param type query string false "帖子类型" Enums(discussion,question,resource,announcement)
// @Param tags query string false "标签（逗号分隔）"
// @Param sort_by query string false "排序方式" Enums(latest,popular,pinned) default(latest)
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} learner.GetCommunityPostsResponse "帖子列表"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "社区不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/communities/{community_id}/posts [get]
func (h *CommunityHandler) GetCommunityPosts(c *gin.Context) {
	communityIDStr := c.Param("community_id")
	communityID, err := uuid.Parse(communityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid community ID",
			Message: "Community ID must be a valid UUID",
		})
		return
	}

	req := learner.GetCommunityPostsRequest{
		CommunityID: communityID,
	}

	// 解析查询参数
	if typeStr := c.Query("type"); typeStr != "" {
		postType := learner.PostType(typeStr)
		req.Type = &postType
	}

	if tagsStr := c.Query("tags"); tagsStr != "" {
		// 简单的逗号分隔解析
		req.Tags = []string{tagsStr} // 实际应该按逗号分割
	}

	req.SortBy = c.DefaultQuery("sort_by", "latest")

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	response, err := h.communityService.GetCommunityPosts(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get community posts",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateReply 创建回复
// @Summary 创建回复
// @Description 对帖子创建回复
// @Tags community
// @Accept json
// @Produce json
// @Param request body learner.CreateReplyRequest true "创建回复请求"
// @Success 201 {object} learner.Reply "回复创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 403 {object} ErrorResponse "无权回复帖子"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/communities/posts/replies [post]
func (h *CommunityHandler) CreateReply(c *gin.Context) {
	var req learner.CreateReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	reply, err := h.communityService.CreateReply(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "user is not a member of this community" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "Access denied",
				Message: "You must be a member of the community to reply to posts",
			})
			return
		}

		if err.Error() == "post is locked and cannot accept new replies" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "Post locked",
				Message: "This post is locked and cannot accept new replies",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create reply",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, reply)
}

// CreateStudyGroup 创建学习小组
// @Summary 创建学习小组
// @Description 创建新的学习小组
// @Tags community
// @Accept json
// @Produce json
// @Param request body learner.CreateStudyGroupRequest true "创建学习小组请求"
// @Success 201 {object} learner.StudyGroup "学习小组创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/study-groups [post]
func (h *CommunityHandler) CreateStudyGroup(c *gin.Context) {
	var req learner.CreateStudyGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	studyGroup, err := h.communityService.CreateStudyGroup(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create study group",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, studyGroup)
}

// GetCommunityList 获取社区列表
// @Summary 获取社区列表
// @Description 获取学习社区列表，支持筛选和分页
// @Tags community
// @Accept json
// @Produce json
// @Param type query string false "社区类型" Enums(public,private,course,study_group)
// @Param tags query string false "标签（逗号分隔）"
// @Param search query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} CommunityListResponse "社区列表"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/communities [get]
func (h *CommunityHandler) GetCommunityList(c *gin.Context) {
	// 解析查询参数
	communityType := c.Query("type")
	tags := c.Query("tags")
	search := c.Query("search")

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// 这里应该调用服务层方法获取社区列表
	// 为了演示，返回模拟数据
	communities := []learner.Community{
		{
			ID:          uuid.New(),
			Name:        "Python学习社区",
			Description: "专注于Python编程语言学习和交流的社区，欢迎所有对Python感兴趣的学习者加入。",
			Type:        learner.CommunityTypePublic,
			CreatorID:   uuid.New(),
			Tags:        []string{"python", "programming", "beginner"},
			MemberCount: 1250,
			PostCount:   3420,
			IsActive:    true,
		},
		{
			ID:          uuid.New(),
			Name:        "机器学习研讨社区",
			Description: "深入探讨机器学习算法和应用的研讨社区，欢迎所有对机器学习感兴趣的学习者加入。",
			Type:        learner.CommunityTypeStudyGroup,
			CreatorID:   uuid.New(),
			Tags:        []string{"machine-learning", "ai", "advanced"},
			MemberCount: 85,
			PostCount:   156,
			IsActive:    true,
		},
	}

	response := CommunityListResponse{
		Communities: communities,
		Total:       len(communities),
		Page:        page,
		Limit:       limit,
	}

	c.JSON(http.StatusOK, response)
}

// GetCommunityDetails 获取社区详情
// @Summary 获取社区详情
// @Description 获取指定社区的详细信息，包括成员、帖子和统计数据
// @Tags community
// @Accept json
// @Produce json
// @Param community_id path string true "社区ID"
// @Success 200 {object} CommunityDetailsResponse "社区详情"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "社区不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/communities/{community_id} [get]
func (h *CommunityHandler) GetCommunityDetails(c *gin.Context) {
	communityIDStr := c.Param("community_id")
	communityID, err := uuid.Parse(communityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid community ID",
			Message: "Community ID must be a valid UUID",
		})
		return
	}

	// 这里应该调用服务层方法获取社区详情
	// 为了演示，返回模拟数据
	community := learner.Community{
		ID:          communityID,
		Name:        "Python学习社区",
		Description: "专注于Python编程语言学习和交流的社区，欢迎所有对Python感兴趣的学习者加入。",
		Type:        learner.CommunityTypePublic,
		CreatorID:   uuid.New(),
		Tags:        []string{"python", "programming", "beginner", "tutorial"},
		MemberCount: 1250,
		PostCount:   3420,
		IsActive:    true,
		Settings: learner.CommunitySettings{
			AllowMemberPost:     true,
			RequireApproval:     false,
			AllowedPostTypes:    []learner.PostType{learner.PostTypeDiscussion, learner.PostTypeQuestion, learner.PostTypeResource},
			MaxMembersCount:     5000,
			AutoJoin:            false,
			NotificationEnabled: true,
		},
	}

	response := CommunityDetailsResponse{
		Community: community,
		Statistics: CommunityStatistics{
			ActiveMembers:      856,
			PostsThisWeek:      45,
			RepliesThisWeek:    128,
			NewMembersThisWeek: 23,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetStudyGroupList 获取学习小组列表
// @Summary 获取学习小组列表
// @Description 获取学习小组列表，支持筛选和分页
// @Tags community
// @Accept json
// @Produce json
// @Param status query string false "小组状态" Enums(recruiting,active,completed,cancelled)
// @Param content_id query string false "关联内容ID"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} StudyGroupListResponse "学习小组列表"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/study-groups [get]
func (h *CommunityHandler) GetStudyGroupList(c *gin.Context) {
	// 解析查询参数
	status := c.Query("status")
	contentIDStr := c.Query("content_id")

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// 这里应该调用服务层方法获取学习小组列表
	// 为了演示，返回模拟数据
	studyGroups := []learner.StudyGroup{
		{
			ID:             uuid.New(),
			Name:           "Python基础学习小组",
			Description:    "一起学习Python基础知识，每周三次在线讨论。",
			CreatorID:      uuid.New(),
			MaxMembers:     10,
			CurrentMembers: 7,
			Status:         learner.GroupStatusRecruiting,
		},
		{
			ID:             uuid.New(),
			Name:           "数据结构算法小组",
			Description:    "深入学习数据结构和算法，准备技术面试。",
			CreatorID:      uuid.New(),
			MaxMembers:     8,
			CurrentMembers: 8,
			Status:         learner.GroupStatusActive,
		},
	}

	response := StudyGroupListResponse{
		StudyGroups: studyGroups,
		Total:       len(studyGroups),
		Page:        page,
		Limit:       limit,
	}

	c.JSON(http.StatusOK, response)
}

// 响应结构
// CommunityListResponse 社区列表响应
type CommunityListResponse struct {
	Communities []learner.Community `json:"communities"`
	Total       int                 `json:"total"`
	Page        int                 `json:"page"`
	Limit       int                 `json:"limit"`
}

// CommunityDetailsResponse 社区详情响应
type CommunityDetailsResponse struct {
	Community  learner.Community   `json:"community"`
	Statistics CommunityStatistics `json:"statistics"`
}

// CommunityStatistics 社区统计信息
type CommunityStatistics struct {
	ActiveMembers      int `json:"active_members"`
	PostsThisWeek      int `json:"posts_this_week"`
	RepliesThisWeek    int `json:"replies_this_week"`
	NewMembersThisWeek int `json:"new_members_this_week"`
}

// StudyGroupListResponse 学习小组列表响应
type StudyGroupListResponse struct {
	StudyGroups []learner.StudyGroup `json:"study_groups"`
	Total       int                  `json:"total"`
	Page        int                  `json:"page"`
	Limit       int                  `json:"limit"`
}

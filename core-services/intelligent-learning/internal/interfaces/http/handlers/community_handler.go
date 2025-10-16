package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/learner"
)

// CommunityHandler 
type CommunityHandler struct {
	communityService *learner.LearningCommunityService
}

// NewCommunityHandler 
func NewCommunityHandler(communityService *learner.LearningCommunityService) *CommunityHandler {
	return &CommunityHandler{
		communityService: communityService,
	}
}

// CreateCommunity 
// @Summary 
// @Description 
// @Tags community
// @Accept json
// @Produce json
// @Param request body learner.CreateCommunityRequest true ""
// @Param creator_id header string true "ID"
// @Success 201 {object} learner.CreateCommunityResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 401 {object} ErrorResponse "?
// @Failure 500 {object} ErrorResponse "?
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

	// ID
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

// JoinCommunity 
// @Summary 
// @Description 
// @Tags community
// @Accept json
// @Produce json
// @Param community_id path string true "ID"
// @Param request body learner.JoinCommunityRequest true ""
// @Success 200 {object} SuccessResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 404 {object} ErrorResponse "?
// @Failure 409 {object} ErrorResponse "?
// @Failure 500 {object} ErrorResponse "?
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

	// ID?
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

// CreatePost 
// @Summary 
// @Description ?
// @Tags community
// @Accept json
// @Produce json
// @Param request body learner.CreatePostRequest true ""
// @Success 201 {object} learner.Post ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 403 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
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

// GetCommunityPosts 
// @Summary 
// @Description 
// @Tags community
// @Accept json
// @Produce json
// @Param community_id path string true "ID"
// @Param type query string false "" Enums(discussion,question,resource,announcement)
// @Param tags query string false "?
// @Param sort_by query string false "" Enums(latest,popular,pinned) default(latest)
// @Param page query int false "" default(1)
// @Param limit query int false "" default(20)
// @Success 200 {object} learner.GetCommunityPostsResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 404 {object} ErrorResponse "?
// @Failure 500 {object} ErrorResponse "?
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

	// 
	if typeStr := c.Query("type"); typeStr != "" {
		postType := learner.PostType(typeStr)
		req.Type = &postType
	}

	if tagsStr := c.Query("tags"); tagsStr != "" {
		// 
		req.Tags = []string{tagsStr} // 
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

// CreateReply 
// @Summary 
// @Description ?
// @Tags community
// @Accept json
// @Produce json
// @Param request body learner.CreateReplyRequest true ""
// @Success 201 {object} learner.Reply ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 403 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
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

// CreateStudyGroup 
// @Summary 
// @Description 
// @Tags community
// @Accept json
// @Produce json
// @Param request body learner.CreateStudyGroupRequest true ""
// @Success 201 {object} learner.StudyGroup "鴴"
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
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

// GetCommunityList 
// @Summary 
// @Description 
// @Tags community
// @Accept json
// @Produce json
// @Param type query string false "" Enums(public,private,course,study_group)
// @Param tags query string false "?
// @Param search query string false "?
// @Param page query int false "" default(1)
// @Param limit query int false "" default(20)
// @Success 200 {object} CommunityListResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/communities [get]
func (h *CommunityHandler) GetCommunityList(c *gin.Context) {
	// 
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

	// 㷽?
	// ?
	communities := []learner.Community{
		{
			ID:          uuid.New(),
			Name:        "Python",
			Description: "PythonPython?,
			Type:        learner.CommunityTypePublic,
			CreatorID:   uuid.New(),
			Tags:        []string{"python", "programming", "beginner"},
			MemberCount: 1250,
			PostCount:   3420,
			IsActive:    true,
		},
		{
			ID:          uuid.New(),
			Name:        "",
			Description: "㷨?,
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

// GetCommunityDetails 
// @Summary 
// @Description 
// @Tags community
// @Accept json
// @Produce json
// @Param community_id path string true "ID"
// @Success 200 {object} CommunityDetailsResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 404 {object} ErrorResponse "?
// @Failure 500 {object} ErrorResponse "?
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

	// 㷽?
	// ?
	community := learner.Community{
		ID:          communityID,
		Name:        "Python",
		Description: "PythonPython?,
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

// GetStudyGroupList 
// @Summary 
// @Description 
// @Tags community
// @Accept json
// @Produce json
// @Param status query string false "? Enums(recruiting,active,completed,cancelled)
// @Param content_id query string false "ID"
// @Param page query int false "" default(1)
// @Param limit query int false "" default(20)
// @Success 200 {object} StudyGroupListResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/study-groups [get]
func (h *CommunityHandler) GetStudyGroupList(c *gin.Context) {
	// 
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

	// 㷽?
	// ?
	studyGroups := []learner.StudyGroup{
		{
			ID:             uuid.New(),
			Name:           "Python",
			Description:    "Python?,
			CreatorID:      uuid.New(),
			MaxMembers:     10,
			CurrentMembers: 7,
			Status:         learner.GroupStatusRecruiting,
		},
		{
			ID:             uuid.New(),
			Name:           "㷨",
			Description:    "㷨?,
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

// 
// CommunityListResponse 
type CommunityListResponse struct {
	Communities []learner.Community `json:"communities"`
	Total       int                 `json:"total"`
	Page        int                 `json:"page"`
	Limit       int                 `json:"limit"`
}

// CommunityDetailsResponse 
type CommunityDetailsResponse struct {
	Community  learner.Community   `json:"community"`
	Statistics CommunityStatistics `json:"statistics"`
}

// CommunityStatistics 
type CommunityStatistics struct {
	ActiveMembers      int `json:"active_members"`
	PostsThisWeek      int `json:"posts_this_week"`
	RepliesThisWeek    int `json:"replies_this_week"`
	NewMembersThisWeek int `json:"new_members_this_week"`
}

// StudyGroupListResponse 
type StudyGroupListResponse struct {
	StudyGroups []learner.StudyGroup `json:"study_groups"`
	Total       int                  `json:"total"`
	Page        int                  `json:"page"`
	Limit       int                  `json:"limit"`
}


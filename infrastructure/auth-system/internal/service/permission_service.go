package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/models"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/repository"
)

var (
	ErrPermissionNotFound     = errors.New("permission not found")
	ErrPermissionExists       = errors.New("permission already exists")
	ErrRolePermissionExists   = errors.New("role permission already exists")
	ErrRolePermissionNotFound = errors.New("role permission not found")
	ErrUserPermissionExists   = errors.New("user permission already exists")
	ErrUserPermissionNotFound = errors.New("user permission not found")
)

// PermissionService 权限管理服务接口
type PermissionService interface {
	// 权限管理
	CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*models.Permission, error)
	GetPermission(ctx context.Context, id uuid.UUID) (*models.Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	ListPermissions(ctx context.Context, req *ListPermissionsRequest) ([]*models.Permission, int64, error)
	UpdatePermission(ctx context.Context, id uuid.UUID, req *UpdatePermissionRequest) (*models.Permission, error)
	DeletePermission(ctx context.Context, id uuid.UUID) error

	// 角色权限管理
	GetRolePermissions(ctx context.Context, role models.UserRole) ([]*models.Permission, error)
	AssignRolePermission(ctx context.Context, role models.UserRole, permissionID uuid.UUID) error
	RevokeRolePermission(ctx context.Context, role models.UserRole, permissionID uuid.UUID) error
	HasRolePermission(ctx context.Context, role models.UserRole, resource, action string) (bool, error)

	// 用户权限管理
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.Permission, error)
	GetUserEffectivePermissions(ctx context.Context, userID uuid.UUID) ([]*models.Permission, error)
	AssignUserPermission(ctx context.Context, userID, permissionID uuid.UUID, granted bool) error
	RevokeUserPermission(ctx context.Context, userID, permissionID uuid.UUID) error
	HasUserPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)

	// 权限检查
	CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
}

// 请求结构体
type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=255"`
	Resource    string `json:"resource" validate:"required,min=1,max=100"`
	Action      string `json:"action" validate:"required,min=1,max=50"`
}

type UpdatePermissionRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=255"`
	Resource    *string `json:"resource,omitempty" validate:"omitempty,min=1,max=100"`
	Action      *string `json:"action,omitempty" validate:"omitempty,min=1,max=50"`
}

type ListPermissionsRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	Resource string `json:"resource,omitempty"`
	Action   string `json:"action,omitempty"`
	Search   string `json:"search,omitempty"`
}

// permissionService 权限服务实现
type permissionService struct {
	db                   *gorm.DB
	permissionRepo       repository.PermissionRepository
	rolePermissionRepo   repository.RolePermissionRepository
	userPermissionRepo   repository.UserPermissionRepository
	userRepo             repository.UserRepository
	logger               *zap.Logger
}

// NewPermissionService 创建权限服务实例
func NewPermissionService(
	db *gorm.DB,
	permissionRepo repository.PermissionRepository,
	rolePermissionRepo repository.RolePermissionRepository,
	userPermissionRepo repository.UserPermissionRepository,
	userRepo repository.UserRepository,
	logger *zap.Logger,
) PermissionService {
	return &permissionService{
		db:                   db,
		permissionRepo:       permissionRepo,
		rolePermissionRepo:   rolePermissionRepo,
		userPermissionRepo:   userPermissionRepo,
		userRepo:             userRepo,
		logger:               logger,
	}
}

// CreatePermission 创建权限
func (s *permissionService) CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*models.Permission, error) {
	// 检查权限是否已存在
	existing, err := s.permissionRepo.GetByName(ctx, req.Name)
	if err != nil && !errors.Is(err, repository.ErrPermissionNotFound) {
		s.logger.Error("Failed to check existing permission", zap.Error(err))
		return nil, err
	}
	if existing != nil {
		return nil, ErrPermissionExists
	}

	// 创建权限
	permission := &models.Permission{
		Name:        req.Name,
		Description: req.Description,
		Resource:    req.Resource,
		Action:      req.Action,
	}

	if err := s.permissionRepo.Create(ctx, permission); err != nil {
		s.logger.Error("Failed to create permission", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Permission created successfully",
		zap.String("permission_id", permission.ID.String()),
		zap.String("name", permission.Name),
	)

	return permission, nil
}

// GetPermission 获取权限
func (s *permissionService) GetPermission(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	permission, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrPermissionNotFound) {
			return nil, ErrPermissionNotFound
		}
		s.logger.Error("Failed to get permission", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}
	return permission, nil
}

// GetPermissionByName 根据名称获取权限
func (s *permissionService) GetPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	permission, err := s.permissionRepo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, repository.ErrPermissionNotFound) {
			return nil, ErrPermissionNotFound
		}
		s.logger.Error("Failed to get permission by name", zap.String("name", name), zap.Error(err))
		return nil, err
	}
	return permission, nil
}

// ListPermissions 获取权限列表
func (s *permissionService) ListPermissions(ctx context.Context, req *ListPermissionsRequest) ([]*models.Permission, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	permissions, total, err := s.permissionRepo.List(ctx, &repository.ListPermissionsOptions{
		Page:     req.Page,
		PageSize: req.PageSize,
		Resource: req.Resource,
		Action:   req.Action,
		Search:   req.Search,
	})
	if err != nil {
		s.logger.Error("Failed to list permissions", zap.Error(err))
		return nil, 0, err
	}

	return permissions, total, nil
}

// UpdatePermission 更新权限
func (s *permissionService) UpdatePermission(ctx context.Context, id uuid.UUID, req *UpdatePermissionRequest) (*models.Permission, error) {
	permission, err := s.GetPermission(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Name != nil {
		// 检查名称是否已被其他权限使用
		existing, err := s.permissionRepo.GetByName(ctx, *req.Name)
		if err != nil && !errors.Is(err, repository.ErrPermissionNotFound) {
			s.logger.Error("Failed to check existing permission", zap.Error(err))
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrPermissionExists
		}
		permission.Name = *req.Name
	}
	if req.Description != nil {
		permission.Description = *req.Description
	}
	if req.Resource != nil {
		permission.Resource = *req.Resource
	}
	if req.Action != nil {
		permission.Action = *req.Action
	}

	if err := s.permissionRepo.Update(ctx, permission); err != nil {
		s.logger.Error("Failed to update permission", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}

	s.logger.Info("Permission updated successfully",
		zap.String("permission_id", permission.ID.String()),
		zap.String("name", permission.Name),
	)

	return permission, nil
}

// DeletePermission 删除权限
func (s *permissionService) DeletePermission(ctx context.Context, id uuid.UUID) error {
	// 检查权限是否存在
	_, err := s.GetPermission(ctx, id)
	if err != nil {
		return err
	}

	// 删除相关的角色权限和用户权限
	if err := s.rolePermissionRepo.DeleteByPermissionID(ctx, id); err != nil {
		s.logger.Error("Failed to delete role permissions", zap.String("permission_id", id.String()), zap.Error(err))
		return err
	}

	if err := s.userPermissionRepo.DeleteByPermissionID(ctx, id); err != nil {
		s.logger.Error("Failed to delete user permissions", zap.String("permission_id", id.String()), zap.Error(err))
		return err
	}

	// 删除权限
	if err := s.permissionRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete permission", zap.String("id", id.String()), zap.Error(err))
		return err
	}

	s.logger.Info("Permission deleted successfully", zap.String("permission_id", id.String()))
	return nil
}

// GetRolePermissions 获取角色权限
func (s *permissionService) GetRolePermissions(ctx context.Context, role models.UserRole) ([]*models.Permission, error) {
	permissions, err := s.rolePermissionRepo.GetPermissionsByRole(ctx, role)
	if err != nil {
		s.logger.Error("Failed to get role permissions", zap.String("role", string(role)), zap.Error(err))
		return nil, err
	}
	return permissions, nil
}

// AssignRolePermission 分配角色权限
func (s *permissionService) AssignRolePermission(ctx context.Context, role models.UserRole, permissionID uuid.UUID) error {
	// 检查权限是否存在
	_, err := s.GetPermission(ctx, permissionID)
	if err != nil {
		return err
	}

	// 检查角色权限是否已存在
	exists, err := s.rolePermissionRepo.Exists(ctx, role, permissionID)
	if err != nil {
		s.logger.Error("Failed to check role permission existence", zap.Error(err))
		return err
	}
	if exists {
		return ErrRolePermissionExists
	}

	// 创建角色权限
	rolePermission := &models.RolePermission{
		Role:         role,
		PermissionID: permissionID,
	}

	if err := s.rolePermissionRepo.Create(ctx, rolePermission); err != nil {
		s.logger.Error("Failed to assign role permission", zap.Error(err))
		return err
	}

	s.logger.Info("Role permission assigned successfully",
		zap.String("role", string(role)),
		zap.String("permission_id", permissionID.String()),
	)

	return nil
}

// RevokeRolePermission 撤销角色权限
func (s *permissionService) RevokeRolePermission(ctx context.Context, role models.UserRole, permissionID uuid.UUID) error {
	// 检查角色权限是否存在
	exists, err := s.rolePermissionRepo.Exists(ctx, role, permissionID)
	if err != nil {
		s.logger.Error("Failed to check role permission existence", zap.Error(err))
		return err
	}
	if !exists {
		return ErrRolePermissionNotFound
	}

	if err := s.rolePermissionRepo.Delete(ctx, role, permissionID); err != nil {
		s.logger.Error("Failed to revoke role permission", zap.Error(err))
		return err
	}

	s.logger.Info("Role permission revoked successfully",
		zap.String("role", string(role)),
		zap.String("permission_id", permissionID.String()),
	)

	return nil
}

// HasRolePermission 检查角色是否有指定权限
func (s *permissionService) HasRolePermission(ctx context.Context, role models.UserRole, resource, action string) (bool, error) {
	return s.rolePermissionRepo.HasPermission(ctx, role, resource, action)
}

// GetUserPermissions 获取用户直接分配的权限
func (s *permissionService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.Permission, error) {
	permissions, err := s.userPermissionRepo.GetPermissionsByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user permissions", zap.String("user_id", userID.String()), zap.Error(err))
		return nil, err
	}
	return permissions, nil
}

// GetUserEffectivePermissions 获取用户有效权限（角色权限 + 用户权限）
func (s *permissionService) GetUserEffectivePermissions(ctx context.Context, userID uuid.UUID) ([]*models.Permission, error) {
	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.String("user_id", userID.String()), zap.Error(err))
		return nil, err
	}

	// 获取角色权限
	rolePermissions, err := s.GetRolePermissions(ctx, user.Role)
	if err != nil {
		return nil, err
	}

	// 获取用户直接权限
	userPermissions, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 合并权限（去重）
	permissionMap := make(map[uuid.UUID]*models.Permission)
	
	// 添加角色权限
	for _, perm := range rolePermissions {
		permissionMap[perm.ID] = perm
	}
	
	// 添加用户权限（可能覆盖角色权限）
	for _, perm := range userPermissions {
		permissionMap[perm.ID] = perm
	}

	// 转换为切片
	var effectivePermissions []*models.Permission
	for _, perm := range permissionMap {
		effectivePermissions = append(effectivePermissions, perm)
	}

	return effectivePermissions, nil
}

// AssignUserPermission 分配用户权限
func (s *permissionService) AssignUserPermission(ctx context.Context, userID, permissionID uuid.UUID, granted bool) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return repository.ErrUserNotFound
		}
		s.logger.Error("Failed to get user", zap.String("user_id", userID.String()), zap.Error(err))
		return err
	}

	// 检查权限是否存在
	_, err = s.GetPermission(ctx, permissionID)
	if err != nil {
		return err
	}

	// 检查用户权限是否已存在
	exists, err := s.userPermissionRepo.Exists(ctx, userID, permissionID)
	if err != nil {
		s.logger.Error("Failed to check user permission existence", zap.Error(err))
		return err
	}

	if exists {
		// 更新现有权限
		if err := s.userPermissionRepo.UpdateGranted(ctx, userID, permissionID, granted); err != nil {
			s.logger.Error("Failed to update user permission", zap.Error(err))
			return err
		}
	} else {
		// 创建新权限
		userPermission := &models.UserPermission{
			UserID:       userID,
			PermissionID: permissionID,
			Granted:      granted,
		}

		if err := s.userPermissionRepo.Create(ctx, userPermission); err != nil {
			s.logger.Error("Failed to assign user permission", zap.Error(err))
			return err
		}
	}

	s.logger.Info("User permission assigned successfully",
		zap.String("user_id", userID.String()),
		zap.String("permission_id", permissionID.String()),
		zap.Bool("granted", granted),
	)

	return nil
}

// RevokeUserPermission 撤销用户权限
func (s *permissionService) RevokeUserPermission(ctx context.Context, userID, permissionID uuid.UUID) error {
	// 检查用户权限是否存在
	exists, err := s.userPermissionRepo.Exists(ctx, userID, permissionID)
	if err != nil {
		s.logger.Error("Failed to check user permission existence", zap.Error(err))
		return err
	}
	if !exists {
		return ErrUserPermissionNotFound
	}

	if err := s.userPermissionRepo.Delete(ctx, userID, permissionID); err != nil {
		s.logger.Error("Failed to revoke user permission", zap.Error(err))
		return err
	}

	s.logger.Info("User permission revoked successfully",
		zap.String("user_id", userID.String()),
		zap.String("permission_id", permissionID.String()),
	)

	return nil
}

// HasUserPermission 检查用户是否有指定权限
func (s *permissionService) HasUserPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	return s.userPermissionRepo.HasPermission(ctx, userID, resource, action)
}

// CheckPermission 检查用户权限（综合角色权限和用户权限）
func (s *permissionService) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.String("user_id", userID.String()), zap.Error(err))
		return false, err
	}

	// 检查角色权限
	hasRolePermission, err := s.HasRolePermission(ctx, user.Role, resource, action)
	if err != nil {
		s.logger.Error("Failed to check role permission", zap.Error(err))
		return false, err
	}

	// 检查用户直接权限
	hasUserPermission, err := s.HasUserPermission(ctx, userID, resource, action)
	if err != nil {
		s.logger.Error("Failed to check user permission", zap.Error(err))
		return false, err
	}

	// 用户权限优先级高于角色权限
	return hasRolePermission || hasUserPermission, nil
}
package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
	
	"github.com/codetaoist/taishanglaojun/core-services/internal/models"
)

// DatabaseConnectionService 数据库连接管理服务
type DatabaseConnectionService struct {
	db         *gorm.DB
	encryptKey []byte // 32字节的加密密钥
}

// NewDatabaseConnectionService 创建数据库连接管理服务实例
func NewDatabaseConnectionService(db *gorm.DB, encryptKey string) *DatabaseConnectionService {
	// 确保密钥长度为32字节
	key := make([]byte, 32)
	copy(key, []byte(encryptKey))
	
	return &DatabaseConnectionService{
		db:         db,
		encryptKey: key,
	}
}

// encrypt 加密密码
func (s *DatabaseConnectionService) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt 解密密码
func (s *DatabaseConnectionService) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// CreateConnection 创建数据库连接配置
func (s *DatabaseConnectionService) CreateConnection(ctx context.Context, form *models.DatabaseConnectionForm, userID string) (*models.DatabaseConnection, error) {
	// 加密密码
	encryptedPassword, err := s.encrypt(form.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}

	// 序列化标签
	tagsJSON, _ := json.Marshal(form.Tags)

	// 创建连接配置
	connection := &models.DatabaseConnection{
		ID:                uuid.New().String(),
		Name:              form.Name,
		Type:              form.Type,
		Host:              form.Host,
		Port:              form.Port,
		Database:          form.Database,
		Username:          form.Username,
		Password:          encryptedPassword,
		SSL:               form.SSL != nil && *form.SSL,
		ConnectionTimeout: 30,
		MaxConnections:    10,
		Description:       form.Description,
		Tags:              string(tagsJSON),
		IsDefault:         form.IsDefault != nil && *form.IsDefault,
		CreatedBy:         userID,
	}

	if form.ConnectionTimeout != nil {
		connection.ConnectionTimeout = *form.ConnectionTimeout
	}
	if form.MaxConnections != nil {
		connection.MaxConnections = *form.MaxConnections
	}

	// 如果设置为默认连接，先取消其他默认连接
	if connection.IsDefault {
		if err := s.db.Model(&models.DatabaseConnection{}).
			Where("is_default = ? AND type = ?", true, connection.Type).
			Update("is_default", false).Error; err != nil {
			return nil, fmt.Errorf("failed to update existing default connections: %w", err)
		}
	}

	// 保存到数据库
	if err := s.db.Create(connection).Error; err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// 创建初始状态记录
	status := &models.DatabaseConnectionStatus{
		ID:           uuid.New().String(),
		ConnectionID: connection.ID,
		Status:       models.ConnectionStatusUnknown,
		LastChecked:  time.Now(),
	}
	s.db.Create(status)

	// 记录事件
	s.logEvent(connection.ID, "create", true, "Database connection created", userID, "", "")

	// 隐藏密码后返回
	connection.Password = "***"
	return connection, nil
}

// GetConnections 获取数据库连接列表
func (s *DatabaseConnectionService) GetConnections(ctx context.Context, query *models.DatabaseConnectionQuery) ([]*models.DatabaseConnection, int64, error) {
	var connections []*models.DatabaseConnection
	var total int64

	db := s.db.Model(&models.DatabaseConnection{})

	// 应用搜索过滤
	if query.Search != "" {
		searchTerm := "%" + query.Search + "%"
		db = db.Where("name LIKE ? OR description LIKE ? OR host LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// 应用类型过滤
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}

	// 应用标签过滤
	if query.Tags != "" {
		tags := strings.Split(query.Tags, ",")
		for _, tag := range tags {
			db = db.Where("tags LIKE ?", "%"+strings.TrimSpace(tag)+"%")
		}
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count connections: %w", err)
	}

	// 应用排序
	sortBy := "created_at"
	sortOrder := "desc"
	if query.SortBy != "" {
		sortBy = query.SortBy
	}
	if query.SortOrder != "" {
		sortOrder = query.SortOrder
	}
	db = db.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// 应用分页
	page := 1
	pageSize := 20
	if query.Page > 0 {
		page = query.Page
	}
	if query.PageSize > 0 {
		pageSize = query.PageSize
	}
	offset := (page - 1) * pageSize
	db = db.Offset(offset).Limit(pageSize)

	// 预加载状态信息
	if err := db.Preload("Status").Find(&connections).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get connections: %w", err)
	}

	// 隐藏密码
	for _, conn := range connections {
		conn.Password = "***"
	}

	return connections, total, nil
}

// GetConnection 获取单个数据库连接配置
func (s *DatabaseConnectionService) GetConnection(ctx context.Context, id string) (*models.DatabaseConnection, error) {
	var connection models.DatabaseConnection
	if err := s.db.Preload("Status").First(&connection, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("connection not found")
		}
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	// 隐藏密码
	connection.Password = "***"
	return &connection, nil
}

// UpdateConnection 更新数据库连接配置
func (s *DatabaseConnectionService) UpdateConnection(ctx context.Context, id string, form *models.DatabaseConnectionForm, userID string) (*models.DatabaseConnection, error) {
	var connection models.DatabaseConnection
	if err := s.db.First(&connection, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("connection not found")
		}
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	// 更新字段
	updates := map[string]interface{}{
		"name":        form.Name,
		"type":        form.Type,
		"host":        form.Host,
		"port":        form.Port,
		"database":    form.Database,
		"username":    form.Username,
		"description": form.Description,
		"updated_at":  time.Now(),
	}

	// 如果提供了新密码，加密后更新
	if form.Password != "" && form.Password != "***" {
		encryptedPassword, err := s.encrypt(form.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %w", err)
		}
		updates["password"] = encryptedPassword
	}

	// 更新可选字段
	if form.SSL != nil {
		updates["ssl"] = *form.SSL
	}
	if form.ConnectionTimeout != nil {
		updates["connection_timeout"] = *form.ConnectionTimeout
	}
	if form.MaxConnections != nil {
		updates["max_connections"] = *form.MaxConnections
	}
	if form.IsDefault != nil {
		updates["is_default"] = *form.IsDefault
		
		// 如果设置为默认连接，先取消其他默认连接
		if *form.IsDefault {
			s.db.Model(&models.DatabaseConnection{}).
				Where("is_default = ? AND type = ? AND id != ?", true, form.Type, id).
				Update("is_default", false)
		}
	}

	// 更新标签
	if form.Tags != nil {
		tagsJSON, _ := json.Marshal(form.Tags)
		updates["tags"] = string(tagsJSON)
	}

	// 执行更新
	if err := s.db.Model(&connection).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update connection: %w", err)
	}

	// 记录事件
	s.logEvent(id, "update", true, "Database connection updated", userID, "", "")

	// 重新获取更新后的连接
	if err := s.db.Preload("Status").First(&connection, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get updated connection: %w", err)
	}

	// 隐藏密码
	connection.Password = "***"
	return &connection, nil
}

// DeleteConnection 删除数据库连接配置
func (s *DatabaseConnectionService) DeleteConnection(ctx context.Context, id string, userID string) error {
	var connection models.DatabaseConnection
	if err := s.db.First(&connection, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("connection not found")
		}
		return fmt.Errorf("failed to get connection: %w", err)
	}

	// 删除相关的状态记录
	s.db.Delete(&models.DatabaseConnectionStatus{}, "connection_id = ?", id)

	// 删除连接配置
	if err := s.db.Delete(&connection).Error; err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	// 记录事件
	s.logEvent(id, "delete", true, "Database connection deleted", userID, "", "")

	return nil
}

// TestConnection 测试数据库连接
func (s *DatabaseConnectionService) TestConnection(ctx context.Context, form *models.DatabaseConnectionForm) (*models.DatabaseConnectionTest, error) {
	startTime := time.Now()
	
	// 构建连接字符串
	connStr, err := s.buildConnectionString(form)
	if err != nil {
		return &models.DatabaseConnectionTest{
			Success:      false,
			ResponseTime: int(time.Since(startTime).Milliseconds()),
			ErrorMessage: fmt.Sprintf("Failed to build connection string: %v", err),
		}, nil
	}

	// 尝试连接
	db, err := sql.Open(string(form.Type), connStr)
	if err != nil {
		return &models.DatabaseConnectionTest{
			Success:      false,
			ResponseTime: int(time.Since(startTime).Milliseconds()),
			ErrorMessage: fmt.Sprintf("Failed to open connection: %v", err),
		}, nil
	}
	defer db.Close()

	// 设置连接超时
	timeout := 30 * time.Second
	if form.ConnectionTimeout != nil {
		timeout = time.Duration(*form.ConnectionTimeout) * time.Second
	}
	
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 测试连接
	if err := db.PingContext(ctx); err != nil {
		return &models.DatabaseConnectionTest{
			Success:      false,
			ResponseTime: int(time.Since(startTime).Milliseconds()),
			ErrorMessage: fmt.Sprintf("Connection test failed: %v", err),
		}, nil
	}

	// 获取服务器信息
	serverInfo := &models.DatabaseServerInfo{}
	s.getServerInfo(db, form.Type, serverInfo)

	return &models.DatabaseConnectionTest{
		Success:      true,
		ResponseTime: int(time.Since(startTime).Milliseconds()),
		ServerInfo:   serverInfo,
	}, nil
}

// TestSavedConnection 测试已保存的数据库连接
func (s *DatabaseConnectionService) TestSavedConnection(ctx context.Context, id string, userID string) (*models.DatabaseConnectionTest, error) {
	var connection models.DatabaseConnection
	if err := s.db.First(&connection, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("connection not found")
		}
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	// 解密密码
	password, err := s.decrypt(connection.Password)
	if err != nil {
		return &models.DatabaseConnectionTest{
			Success:      false,
			ResponseTime: 0,
			ErrorMessage: "Failed to decrypt password",
		}, nil
	}

	// 构建表单
	form := &models.DatabaseConnectionForm{
		Type:              connection.Type,
		Host:              connection.Host,
		Port:              connection.Port,
		Database:          connection.Database,
		Username:          connection.Username,
		Password:          password,
		SSL:               &connection.SSL,
		ConnectionTimeout: &connection.ConnectionTimeout,
	}

	// 测试连接
	result, err := s.TestConnection(ctx, form)
	if err != nil {
		return nil, err
	}

	// 更新连接状态
	s.updateConnectionStatus(id, result.Success, result.ResponseTime, result.ErrorMessage, result.ServerInfo)

	// 记录事件
	action := "test"
	message := "Connection test completed"
	if !result.Success {
		message = fmt.Sprintf("Connection test failed: %s", result.ErrorMessage)
	}
	s.logEvent(id, action, result.Success, message, userID, "", "")

	return result, nil
}

// GetConnectionsStatus 获取所有连接状态
func (s *DatabaseConnectionService) GetConnectionsStatus(ctx context.Context) ([]*models.DatabaseConnectionStatus, error) {
	var statuses []*models.DatabaseConnectionStatus
	if err := s.db.Preload("Connection").Find(&statuses).Error; err != nil {
		return nil, fmt.Errorf("failed to get connection statuses: %w", err)
	}

	// 隐藏密码
	for _, status := range statuses {
		status.Connection.Password = "***"
	}

	return statuses, nil
}

// RefreshConnectionStatus 刷新连接状态
func (s *DatabaseConnectionService) RefreshConnectionStatus(ctx context.Context, id string, userID string) (*models.DatabaseConnectionStatus, error) {
	// 测试连接
	_, err := s.TestSavedConnection(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// 获取更新后的状态
	var status models.DatabaseConnectionStatus
	if err := s.db.Preload("Connection").First(&status, "connection_id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get connection status: %w", err)
	}

	// 隐藏密码
	status.Connection.Password = "***"
	return &status, nil
}

// GetConnectionStats 获取连接统计信息
func (s *DatabaseConnectionService) GetConnectionStats(ctx context.Context) (*models.DatabaseConnectionStats, error) {
	var total int64
	s.db.Model(&models.DatabaseConnection{}).Count(&total)

	var active int64
	s.db.Model(&models.DatabaseConnectionStatus{}).
		Where("status = ?", models.ConnectionStatusConnected).
		Count(&active)

	// 按类型统计
	connectionsByType := make(map[models.DatabaseType]int)
	var typeStats []struct {
		Type  models.DatabaseType
		Count int
	}
	s.db.Model(&models.DatabaseConnection{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&typeStats)
	for _, stat := range typeStats {
		connectionsByType[stat.Type] = stat.Count
	}

	// 按状态统计
	connectionsByStatus := make(map[models.ConnectionStatus]int)
	var statusStats []struct {
		Status models.ConnectionStatus
		Count  int
	}
	s.db.Model(&models.DatabaseConnectionStatus{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusStats)
	for _, stat := range statusStats {
		connectionsByStatus[stat.Status] = stat.Count
	}

	// 平均响应时间
	var avgResponseTime float64
	s.db.Model(&models.DatabaseConnectionStatus{}).
		Where("response_time IS NOT NULL").
		Select("AVG(response_time)").
		Scan(&avgResponseTime)

	return &models.DatabaseConnectionStats{
		TotalConnections:    int(total),
		ActiveConnections:   int(active),
		ConnectionsByType:   connectionsByType,
		ConnectionsByStatus: connectionsByStatus,
		AverageResponseTime: avgResponseTime,
		LastUpdated:         time.Now(),
	}, nil
}

// buildConnectionString 构建连接字符串
func (s *DatabaseConnectionService) buildConnectionString(form *models.DatabaseConnectionForm) (string, error) {
	switch form.Type {
	case models.DatabaseTypeMySQL, models.DatabaseTypeMariaDB:
		ssl := "false"
		if form.SSL != nil && *form.SSL {
			ssl = "true"
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=%s", 
			form.Username, form.Password, form.Host, form.Port, form.Database, ssl), nil
	case models.DatabaseTypePostgreSQL:
		ssl := "disable"
		if form.SSL != nil && *form.SSL {
			ssl = "require"
		}
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			form.Username, form.Password, form.Host, form.Port, form.Database, ssl), nil
	case models.DatabaseTypeSQLite:
		return form.Database, nil
	default:
		return "", fmt.Errorf("unsupported database type: %s", form.Type)
	}
}

// getServerInfo 获取服务器信息
func (s *DatabaseConnectionService) getServerInfo(db *sql.DB, dbType models.DatabaseType, info *models.DatabaseServerInfo) {
	switch dbType {
	case models.DatabaseTypeMySQL, models.DatabaseTypeMariaDB:
		db.QueryRow("SELECT VERSION()").Scan(&info.Version)
		db.QueryRow("SELECT @@character_set_server").Scan(&info.Charset)
		db.QueryRow("SELECT @@time_zone").Scan(&info.Timezone)
	case models.DatabaseTypePostgreSQL:
		db.QueryRow("SELECT version()").Scan(&info.Version)
		db.QueryRow("SHOW timezone").Scan(&info.Timezone)
	}
}

// updateConnectionStatus 更新连接状态
func (s *DatabaseConnectionService) updateConnectionStatus(connectionID string, success bool, responseTime int, errorMessage string, serverInfo *models.DatabaseServerInfo) {
	status := models.ConnectionStatusConnected
	if !success {
		status = models.ConnectionStatusError
	}

	updates := map[string]interface{}{
		"status":        status,
		"last_checked":  time.Now(),
		"response_time": responseTime,
		"error_message": errorMessage,
		"updated_at":    time.Now(),
	}

	if serverInfo != nil {
		updates["server_version"] = serverInfo.Version
	}

	s.db.Model(&models.DatabaseConnectionStatus{}).
		Where("connection_id = ?", connectionID).
		Updates(updates)
}

// logEvent 记录事件日志
func (s *DatabaseConnectionService) logEvent(connectionID, action string, success bool, message, userID, ipAddress, userAgent string) {
	event := &models.DatabaseConnectionEvent{
		ID:           uuid.New().String(),
		ConnectionID: connectionID,
		Action:       action,
		Success:      success,
		Message:      message,
		UserID:       userID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}
	s.db.Create(event)
}
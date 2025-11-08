package vector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
)

// Service 向量数据库服务
type Service struct {
	db     VectorDatabase
	config *DatabaseConfig
	mu     sync.RWMutex
}

// NewService 创建一个新的向量数据库服务
func NewService(config *DatabaseConfig) (*Service, error) {
	if config == nil {
		return nil, ErrInvalidConfig("config cannot be nil")
	}

	return &Service{
		config: config,
	}, nil
}

// Disconnect 断开向量数据库连接
func (s *Service) Disconnect(ctx context.Context) error {
	return s.Close()
}

// Connect 连接到向量数据库
func (s *Service) Connect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		// 如果已经连接，先关闭
		if err := s.db.Close(); err != nil {
			return errors.Wrap(err, "failed to close existing connection")
		}
		s.db = nil
	}

	factory := NewVectorDatabaseFactory()
	db, err := factory.ConnectAndCreate(ctx, s.config)
	if err != nil {
		return errors.Wrap(err, "failed to connect to vector database")
	}

	s.db = db
	return nil
}

// IsConnected 检查是否已连接到向量数据库
func (s *Service) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.db != nil
}

// GetDB 获取向量数据库客户端
func (s *Service) GetDB() VectorDatabase {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.db
}

// CreateCollection 创建向量集合
func (s *Service) CreateCollection(ctx context.Context, req *models.CreateCollectionRequest) error {
	if !s.IsConnected() {
		return ErrConnectionFailed("not connected to vector database")
	}

	// 从请求中提取参数
	name := req.CollectionName
	if name == "" {
		name = req.Name // 使用兼容字段
	}

	description := req.Description

	// 获取索引参数
	indexParams := IndexParams{
		IndexType: IndexType(req.IndexType),
		MetricType: MetricType(req.MetricType),
		Params: make(map[string]interface{}),
	}
	
	if req.IndexParams.Nlist > 0 {
		indexParams.Params["nlist"] = req.IndexParams.Nlist
	}
	if req.IndexParams.M > 0 {
		indexParams.Params["M"] = req.IndexParams.M
	}
	if req.IndexParams.EfConstruction > 0 {
		indexParams.Params["efConstruction"] = req.IndexParams.EfConstruction
	}

	return s.db.CreateCollection(ctx, name, description, req.Dimension, indexParams)
}

// DropCollection 删除向量集合
func (s *Service) DropCollection(ctx context.Context, name string) error {
	if !s.IsConnected() {
		return ErrConnectionFailed("not connected to vector database")
	}

	return s.db.DropCollection(ctx, name)
}

// HasCollection 检查集合是否存在
func (s *Service) HasCollection(ctx context.Context, name string) (bool, error) {
	if !s.IsConnected() {
		return false, ErrConnectionFailed("not connected to vector database")
	}

	return s.db.HasCollection(ctx, name)
}

// ListCollections 列出所有集合
func (s *Service) ListCollections(ctx context.Context) ([]string, error) {
	if !s.IsConnected() {
		return nil, ErrConnectionFailed("not connected to vector database")
	}

	return s.db.ListCollections(ctx)
}

// GetCollectionInfo 获取集合信息
func (s *Service) GetCollectionInfo(ctx context.Context, name string) (*CollectionInfo, error) {
	if !s.IsConnected() {
		return nil, ErrConnectionFailed("not connected to vector database")
	}

	return s.db.GetCollectionInfo(ctx, name)
}

// RebuildIndex 重建索引
func (s *Service) RebuildIndex(ctx context.Context, collectionName, fieldName string) error {
	if !s.IsConnected() {
		return ErrConnectionFailed("not connected to vector database")
	}

	// 检查VectorDatabase接口是否有RebuildIndex方法
	type RebuildIndexDB interface {
		RebuildIndex(ctx context.Context, collectionName, fieldName string) error
	}

	// 类型断言检查是否实现了RebuildIndex方法
	if db, ok := s.db.(RebuildIndexDB); ok {
		return db.RebuildIndex(ctx, collectionName, fieldName)
	}

	return ErrUnsupportedOperation("RebuildIndex")
}

// GetCollectionStats 获取集合统计信息
func (s *Service) GetCollectionStats(ctx context.Context, collectionName string) (*models.CollectionStats, error) {
	if !s.IsConnected() {
		return nil, ErrConnectionFailed("not connected to vector database")
	}

	// 调用底层实现
	stats, err := s.db.GetCollectionStats(ctx, collectionName)
	if err != nil {
		return nil, err
	}

	// 转换为models.CollectionStats
	return &models.CollectionStats{
		Name:    stats.Name,
		Count:   stats.VectorCount,
		Size:    stats.SizeInBytes,
		Indexed: true, // 假设有索引，实际应该检查索引状态
	}, nil
}

// CreateIndex 创建索引
func (s *Service) CreateIndex(ctx context.Context, req *models.CreateIndexRequest) error {
	if !s.IsConnected() {
		return ErrConnectionFailed("not connected to vector database")
	}

	return s.db.CreateIndex(ctx, req)
}

// DropIndex 删除索引
func (s *Service) DropIndex(ctx context.Context, collectionName, fieldName string) error {
	if !s.IsConnected() {
		return ErrConnectionFailed("not connected to vector database")
	}

	// 检查VectorDatabase接口是否有DropIndex方法
	type DropIndexDB interface {
		DropIndex(ctx context.Context, collectionName, fieldName string) error
	}

	// 类型断言检查是否实现了DropIndex方法
	if db, ok := s.db.(DropIndexDB); ok {
		return db.DropIndex(ctx, collectionName, fieldName)
	}

	// 如果没有实现带fieldName的DropIndex，尝试调用不带fieldName的版本
	return s.db.DropIndex(ctx, collectionName)
}

// HasIndex 检查索引是否存在
func (s *Service) HasIndex(ctx context.Context, collectionName string) (bool, error) {
	if !s.IsConnected() {
		return false, ErrConnectionFailed("not connected to vector database")
	}

	return s.db.HasIndex(ctx, collectionName)
}

// Insert 插入向量
func (s *Service) Insert(ctx context.Context, collectionName string, vectors []Vector) error {
	if !s.IsConnected() {
		return ErrConnectionFailed("not connected to vector database")
	}

	return s.db.Insert(ctx, collectionName, vectors)
}

// Upsert 更新或插入向量
func (s *Service) Upsert(ctx context.Context, collectionName string, vectors []Vector) error {
	if !s.IsConnected() {
		return ErrConnectionFailed("not connected to vector database")
	}

	return s.db.Upsert(ctx, collectionName, vectors)
}

// Delete 删除向量
func (s *Service) Delete(ctx context.Context, collectionName string, ids []string) error {
	if !s.IsConnected() {
		return ErrConnectionFailed("not connected to vector database")
	}

	return s.db.Delete(ctx, collectionName, ids)
}

// SearchVectors 搜索向量
func (s *Service) SearchVectors(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	if !s.IsConnected() {
		return nil, ErrConnectionFailed("not connected to vector database")
	}

	// 将[]float64转换为[]float32
	queryVector := make([]float32, len(req.QueryVector))
	for i, v := range req.QueryVector {
		queryVector[i] = float32(v)
	}

	// 设置搜索选项
	opts := SearchOptions{
		TopK:   req.TopK,
		Filter: req.Filter,
	}

	// 调用底层实现
	results, err := s.db.Search(ctx, req.CollectionName, queryVector, opts)
	if err != nil {
		return nil, err
	}

	// 转换结果
	searchResults := make([]models.VectorSearchResult, len(results))
	for i, result := range results {
		// 将[]float32转换为[]float64
		var vector []float64
		if result.Vector != nil {
			vector = make([]float64, len(result.Vector))
			for j, v := range result.Vector {
				vector[j] = float64(v)
			}
		}

		searchResults[i] = models.VectorSearchResult{
			ID:       result.ID,
			Score:    float64(result.Score),
			Vector:   vector,
			Metadata: result.Metadata,
		}
	}

	return &models.SearchResponse{
		Results: searchResults,
		Total:   len(searchResults),
	}, nil
}

// Search 搜索向量
func (s *Service) Search(ctx context.Context, collectionName string, queryVector []float32, opts SearchOptions) ([]SearchResult, error) {
	if !s.IsConnected() {
		return nil, ErrConnectionFailed("not connected to vector database")
	}

	return s.db.Search(ctx, collectionName, queryVector, opts)
}

// UpsertVectors 批量插入/更新向量
func (s *Service) UpsertVectors(ctx context.Context, req *models.UpsertVectorsRequest) (*models.UpsertResponse, error) {
	if !s.IsConnected() {
		return nil, ErrConnectionFailed("not connected to vector database")
	}

	// 将models.VectorData转换为vector.Vector
	vectors := make([]Vector, len(req.Vectors))
	for i, v := range req.Vectors {
		// 将[]float64转换为[]float32
		vector32 := make([]float32, len(v.Vector))
		for j, val := range v.Vector {
			vector32[j] = float32(val)
		}

		vectors[i] = Vector{
			ID:       v.ID,
			Vector:   vector32,
			Metadata: v.Metadata,
		}
	}

	// 调用Upsert方法
	err := s.Upsert(ctx, req.CollectionName, vectors)
	if err != nil {
		return nil, err
	}

	// 提取ID
	ids := make([]string, len(vectors))
	for i, v := range vectors {
		ids[i] = v.ID
	}

	// 返回响应
	return &models.UpsertResponse{
		InsertedCount: len(vectors),
		UpdatedCount:  0,
		Ids:           ids,
		SuccessCount:  len(vectors),
		FailedCount:   0,
	}, nil
}

// QueryVectors 搜索向量
func (s *Service) QueryVectors(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	if !s.IsConnected() {
		return nil, ErrConnectionFailed("not connected to vector database")
	}

	// 将[]float64转换为[]float32
	queryVector32 := make([]float32, len(req.QueryVector))
	for i, val := range req.QueryVector {
		queryVector32[i] = float32(val)
	}

	// 设置搜索选项
	opts := SearchOptions{
		TopK:          req.TopK,
		IncludeVector: true,
		Filter:        req.Filter,
	}

	// 调用Search方法
	results, err := s.Search(ctx, req.CollectionName, queryVector32, opts)
	if err != nil {
		return nil, err
	}

	// 将vector.SearchResult转换为models.VectorSearchResult
	searchResults := make([]models.VectorSearchResult, len(results))
	for i, r := range results {
		// 将[]float32转换为[]float64
		vector64 := make([]float64, len(r.Vector))
		for j, val := range r.Vector {
			vector64[j] = float64(val)
		}

		searchResults[i] = models.VectorSearchResult{
			ID:       r.ID,
			Score:    float64(r.Score),
			Vector:   vector64,
			Metadata: r.Metadata,
		}
	}

	// 返回响应
	return &models.SearchResponse{
		Results: searchResults,
		Total:   len(searchResults),
	}, nil
}

// DescribeIndex 获取索引详细信息
func (s *Service) DescribeIndex(ctx context.Context, collectionName, fieldName string) (*models.VectorIndex, error) {
	if !s.IsConnected() {
		return nil, ErrConnectionFailed("not connected to vector database")
	}

	// 检查VectorDatabase接口是否有DescribeIndex方法
	type DescribeIndexDB interface {
		DescribeIndex(ctx context.Context, collectionName, fieldName string) (*models.VectorIndex, error)
	}

	// 类型断言检查是否实现了DescribeIndex方法
	if db, ok := s.db.(DescribeIndexDB); ok {
		return db.DescribeIndex(ctx, collectionName, fieldName)
	}

	return nil, ErrUnsupportedOperation("DescribeIndex")
}

// DescribeCollection 获取集合详细信息
func (s *Service) DescribeCollection(ctx context.Context, collectionName string) (*models.VectorCollection, error) {
	if !s.IsConnected() {
		return nil, ErrConnectionFailed("not connected to vector database")
	}

	// 检查VectorDatabase接口是否有DescribeCollection方法
	type DescribeCollectionDB interface {
		DescribeCollection(ctx context.Context, collectionName string) (*models.VectorCollection, error)
	}

	// 类型断言检查是否实现了DescribeCollection方法
	if db, ok := s.db.(DescribeCollectionDB); ok {
		return db.DescribeCollection(ctx, collectionName)
	}

	return nil, ErrUnsupportedOperation("DescribeCollection")
}

// DeleteVectors 删除向量
func (s *Service) DeleteVectors(ctx context.Context, req *models.DeleteVectorsRequest) (*models.DeleteResponse, error) {
	if !s.IsConnected() {
		return nil, ErrConnectionFailed("not connected to vector database")
	}

	// 调用底层实现
	err := s.db.Delete(ctx, req.CollectionName, req.Ids)
	if err != nil {
		return nil, err
	}

	// 返回删除响应
	return &models.DeleteResponse{
		DeletedCount: len(req.Ids),
	}, nil
}

// GetVector 根据ID获取向量
func (s *Service) GetVector(ctx context.Context, collectionName, vectorID string) (*models.VectorData, error) {
	if !s.IsConnected() {
		return nil, ErrConnectionFailed("not connected to vector database")
	}

	// 调用GetByID方法
	vector, err := s.GetByID(ctx, collectionName, vectorID)
	if err != nil {
		return nil, err
	}

	// 将[]float32转换为[]float64
	vector64 := make([]float64, len(vector.Vector))
	for i, val := range vector.Vector {
		vector64[i] = float64(val)
	}

	// 返回响应
	return &models.VectorData{
		ID:       vector.ID,
		Vector:   vector64,
		Metadata: vector.Metadata,
	}, nil
}

// GetByID 根据ID获取向量
func (s *Service) GetByID(ctx context.Context, collectionName string, id string) (*Vector, error) {
	if !s.IsConnected() {
		return nil, ErrConnectionFailed("not connected to vector database")
	}

	return s.db.GetByID(ctx, collectionName, id)
}

// Health 检查向量数据库健康状态
func (s *Service) Health(ctx context.Context) error {
	if !s.IsConnected() {
		return ErrConnectionFailed("not connected to vector database")
	}

	return s.db.Health(ctx)
}

// Close 关闭向量数据库连接
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		err := s.db.Close()
		s.db = nil
		return err
	}

	return nil
}

// Compact 压缩集合以清理已删除的数据并回收空间
func (s *Service) Compact(ctx context.Context, collectionName string) error {
	if !s.IsConnected() {
		return ErrConnectionFailed("not connected to vector database")
	}

	// 检查VectorDatabase接口是否有Compact方法
	type CompactDB interface {
		Compact(ctx context.Context, collectionName string) error
	}

	// 类型断言检查是否实现了Compact方法
	if db, ok := s.db.(CompactDB); ok {
		return db.Compact(ctx, collectionName)
	}

	return ErrUnsupportedOperation("Compact")
}

// Ping 检查向量数据库连接状态
func (s *Service) Ping(ctx context.Context) error {
	if !s.IsConnected() {
		return ErrConnectionFailed("not connected to vector database")
	}

	// 使用健康检查来验证连接
	return s.Health(ctx)
}

// Reconnect 重新连接到向量数据库
func (s *Service) Reconnect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果已经连接，先关闭
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return errors.Wrap(err, "failed to close existing connection")
		}
		s.db = nil
	}

	// 重新连接
	factory := NewVectorDatabaseFactory()
	db, err := factory.ConnectAndCreate(ctx, s.config)
	if err != nil {
		return errors.Wrap(err, "failed to reconnect to vector database")
	}

	s.db = db
	return nil
}

// GetConfig 获取向量数据库配置
func (s *Service) GetConfig() *DatabaseConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 返回配置的副本
	if s.config == nil {
		return nil
	}

	configCopy := *s.config
	if s.config.Milvus != nil {
		milvusConfigCopy := *s.config.Milvus
		configCopy.Milvus = &milvusConfigCopy
	}

	return &configCopy
}

// UpdateConfig 更新向量数据库配置
func (s *Service) UpdateConfig(config *DatabaseConfig) error {
	if config == nil {
		return ErrInvalidConfig("config cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果已经连接，需要先关闭
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return errors.Wrap(err, "failed to close existing connection")
		}
		s.db = nil
	}

	// 更新配置
	s.config = config
	return nil
}

// GetDatabaseType 获取向量数据库类型
func (s *Service) GetDatabaseType() DatabaseType {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config == nil {
		return ""
	}

	return s.config.Type
}

// GetConnectionInfo 获取连接信息
func (s *Service) GetConnectionInfo() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info := map[string]interface{}{
		"type":         s.GetDatabaseType(),
		"connected":    s.IsConnected(),
		"connected_at": time.Now(), // 这里应该记录实际连接时间
	}

	if s.config != nil {
		switch s.config.Type {
		case DatabaseTypeMilvus:
			if s.config.Milvus != nil {
				info["address"] = fmt.Sprintf("%s:%d", s.config.Milvus.Address, s.config.Milvus.Port)
				info["database"] = s.config.Milvus.Database
				info["enable_tls"] = s.config.Milvus.EnableTLS
			}
		}
	}

	return info
}

// VectorServiceFactory 创建向量服务实例
type VectorServiceFactory struct{}

// NewVectorServiceFactory 创建向量服务工厂
func NewVectorServiceFactory() *VectorServiceFactory {
	return &VectorServiceFactory{}
}

// CreateService 根据配置创建向量服务
func (f *VectorServiceFactory) CreateService(config *DatabaseConfig) (VectorDatabase, error) {
	if config == nil {
		return nil, ErrInvalidConfig("config cannot be nil")
	}

	switch config.Type {
	case DatabaseTypeMilvus:
		if config.Milvus == nil {
			return nil, ErrInvalidConfig("milvus config cannot be nil")
		}
		return NewMilvusClient(config.Milvus)
	case DatabaseTypeQdrant:
		// TODO: 实现Qdrant客户端
		return nil, ErrUnsupportedDB("Qdrant not yet supported")
	case DatabaseTypeWeaviate:
		// TODO: 实现Weaviate客户端
		return nil, ErrUnsupportedDB("Weaviate not yet supported")
	default:
		return nil, ErrUnsupportedDB(string(config.Type))
	}
}

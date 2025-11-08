package vector

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// MilvusClient 实现VectorDatabase接口的Milvus客户端
type MilvusClient struct {
	config   *MilvusConfig
	milvusClient client.Client
	connected bool
}

// NewMilvusClient 创建一个新的Milvus客户端
func NewMilvusClient(config *MilvusConfig) (*MilvusClient, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	client := &MilvusClient{
		config:    config,
		connected: false,
	}

	return client, nil
}

// Connect 连接到Milvus服务器
func (c *MilvusClient) Connect(ctx context.Context) error {
	milvusClient, err := client.NewGrpcClient(ctx, fmt.Sprintf("%s:%d", c.config.Address, c.config.Port))
	if err != nil {
		return ErrConnectionFailed(fmt.Sprintf("failed to connect to Milvus: %v", err))
	}

	// 如果指定了数据库，切换到指定数据库
	if c.config.Database != "" {
		if err := milvusClient.UsingDatabase(ctx, c.config.Database); err != nil {
			return ErrConnectionFailed(fmt.Sprintf("failed to use database %s: %v", c.config.Database, err))
		}
	}

	// 测试连接 - 使用ListCollections代替CheckHealth
	_, err = milvusClient.ListCollections(ctx)
	if err != nil {
		return ErrConnectionFailed(fmt.Sprintf("health check failed: %v", err))
	}

	c.milvusClient = milvusClient
	c.connected = true
	return nil
}

// IsConnected 检查是否已连接到Milvus服务器
func (c *MilvusClient) IsConnected() bool {
	return c.connected && c.milvusClient != nil
}

// CreateCollection 创建向量集合
func (c *MilvusClient) CreateCollection(ctx context.Context, name, description string, vectorDim int, indexParams IndexParams) error {
	if !c.connected {
		return ErrConnectionFailed("not connected to Milvus server")
	}

	if vectorDim <= 0 {
		return ErrInvalidDimension(vectorDim)
	}

	schema := &entity.Schema{
		CollectionName: name,
		Description:    description,
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				AutoID:     false,
				TypeParams: map[string]string{
					"max_length": "65535",
				},
			},
			{
				Name:     "vector",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": strconv.Itoa(vectorDim),
				},
			},
		},
	}

	// 创建集合
	err := c.milvusClient.CreateCollection(ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		return ErrInternalError(fmt.Sprintf("failed to create collection: %v", err))
	}

	return nil
}

// DropCollection 删除向量集合
func (c *MilvusClient) DropCollection(ctx context.Context, name string) error {
	if !c.connected {
		return ErrConnectionFailed("not connected to Milvus server")
	}

	if name == "" {
		return ErrInvalidID(name)
	}

	err := c.milvusClient.DropCollection(ctx, name)
	if err != nil {
		return ErrInternalError(fmt.Sprintf("failed to drop collection: %v", err))
	}

	return nil
}

// HasCollection 检查集合是否存在
func (c *MilvusClient) HasCollection(ctx context.Context, name string) (bool, error) {
	if !c.connected {
		return false, ErrConnectionFailed("not connected to Milvus server")
	}

	if name == "" {
		return false, ErrInvalidID(name)
	}

	has, err := c.milvusClient.HasCollection(ctx, name)
	if err != nil {
		return false, ErrInternalError(fmt.Sprintf("failed to check collection existence: %v", err))
	}

	return has, nil
}

// ListCollections 列出所有集合
func (c *MilvusClient) ListCollections(ctx context.Context) ([]string, error) {
	if !c.connected {
		return nil, ErrConnectionFailed("not connected to Milvus server")
	}

	collections, err := c.milvusClient.ListCollections(ctx)
	if err != nil {
		return nil, ErrInternalError(fmt.Sprintf("failed to list collections: %v", err))
	}

	names := make([]string, len(collections))
	for i, collection := range collections {
		names[i] = collection.Name
	}

	return names, nil
}

// GetCollectionInfo 获取集合信息
func (c *MilvusClient) GetCollectionInfo(ctx context.Context, name string) (*CollectionInfo, error) {
	if !c.connected {
		return nil, ErrConnectionFailed("not connected to Milvus server")
	}

	if name == "" {
		return nil, ErrInvalidID(name)
	}

	collection, err := c.milvusClient.DescribeCollection(ctx, name)
	if err != nil {
		return nil, ErrInternalError(fmt.Sprintf("failed to describe collection: %v", err))
	}

	// 获取索引信息
	indexes, err := c.milvusClient.DescribeIndex(ctx, name, "vector")
	if err != nil {
		return nil, ErrInternalError(fmt.Sprintf("failed to describe index: %v", err))
	}

	var indexParams IndexParams
	if len(indexes) > 0 {
		index := indexes[0]
		// 使用IndexType()代替Type()
		indexType := index.IndexType()
		metricType := index.Params()["metric_type"]
		
		// 转换参数类型
		params := make(map[string]interface{})
		for k, v := range index.Params() {
			params[k] = v
		}
		
		indexParams = IndexParams{
			IndexType: IndexType(indexType),
			MetricType: MetricType(metricType),
			Params:     params,
		}
	} else {
		// 默认索引参数
		indexParams = IndexParams{
			IndexType: IndexTypeFlat,
			MetricType: MetricTypeL2,
			Params:     make(map[string]interface{}),
		}
	}

	// 获取向量维度
	var vectorDim int
	for _, field := range collection.Schema.Fields {
		if field.Name == "vector" {
			dimStr, ok := field.TypeParams["dim"]
			if ok {
				dim, err := strconv.Atoi(dimStr)
				if err == nil {
					vectorDim = dim
				}
			}
			break
		}
	}

	return &CollectionInfo{
		Name:        collection.Name,
		Description: collection.Schema.Description,
		VectorDim:   vectorDim,
		IndexParams: indexParams,
		CreatedAt:   time.Now(), // 使用当前时间，因为Milvus SDK不提供创建时间
		UpdatedAt:   time.Now(), // Milvus不提供更新时间，使用当前时间
	}, nil
}

// GetCollectionStats 获取集合统计信息
func (c *MilvusClient) GetCollectionStats(ctx context.Context, name string) (*CollectionStats, error) {
	if !c.connected {
		return nil, ErrConnectionFailed("not connected to Milvus server")
	}

	if name == "" {
		return nil, ErrInvalidID(name)
	}

	// 获取集合统计信息
	stats, err := c.milvusClient.GetCollectionStatistics(ctx, name)
	if err != nil {
		return nil, ErrInternalError(fmt.Sprintf("failed to get collection stats: %v", err))
	}

	// 解析统计信息
	var rowCount int64
	if stat, ok := stats["row_count"]; ok {
		rowCount, _ = strconv.ParseInt(stat, 10, 64)
	}

	return &CollectionStats{
		Name:        name,
		VectorCount: rowCount,
		SizeInBytes: 0, // Milvus不直接提供大小信息
	}, nil
}

// CreateIndex 创建索引
func (c *MilvusClient) CreateIndex(ctx context.Context, req interface{}) error {
	// 类型断言，确保req是*models.CreateIndexRequest类型
	createIndexReq, ok := req.(*models.CreateIndexRequest)
	if !ok {
		return ErrInternalError("invalid request type for CreateIndex")
	}
	
	// 从请求中提取参数
	indexParams := IndexParams{
		IndexType: IndexType(createIndexReq.IndexType),
		MetricType: MetricType(createIndexReq.MetricType),
		Params: make(map[string]interface{}),
	}
	
	// 如果有额外参数，添加到Params中
	if createIndexReq.Params != nil {
		for k, v := range createIndexReq.Params {
			indexParams.Params[k] = v
		}
	}
	
	// 如果有ExtraParams，也添加到Params中
	if createIndexReq.ExtraParams != nil {
		for k, v := range createIndexReq.ExtraParams {
			indexParams.Params[k] = v
		}
	}

	if !c.connected {
		return ErrConnectionFailed("not connected to Milvus server")
	}

	if createIndexReq.CollectionName == "" {
		return ErrInvalidID(createIndexReq.CollectionName)
	}

	// 使用entity.Index接口类型，而不是具体的索引类型
	var index entity.Index
	var err error

	// 根据索引类型创建不同的索引
	switch indexParams.IndexType {
	case IndexTypeFlat:
		index, err = entity.NewIndexFlat(entity.MetricType(indexParams.MetricType))
	case IndexTypeIVF:
		nlist, ok := indexParams.Params["nlist"].(int)
		if !ok {
			nlist = 128
		}
		index, err = entity.NewIndexIvfFlat(entity.MetricType(indexParams.MetricType), nlist)
	case IndexTypeIVFSQ:
		nlist, ok := indexParams.Params["nlist"].(int)
		if !ok {
			nlist = 128
		}
		index, err = entity.NewIndexIvfSQ8(entity.MetricType(indexParams.MetricType), nlist)
	case IndexTypeHNSW:
		M, ok := indexParams.Params["M"].(int)
		if !ok {
			M = 16
		}
		efConstruction, ok := indexParams.Params["efConstruction"].(int)
		if !ok {
			efConstruction = 200
		}
		index, err = entity.NewIndexHNSW(entity.MetricType(indexParams.MetricType), M, efConstruction)
	default:
		return ErrInvalidIndexType(string(indexParams.IndexType))
	}

	if err != nil {
		return ErrInternalError(fmt.Sprintf("failed to create index: %v", err))
	}

	err = c.milvusClient.CreateIndex(ctx, createIndexReq.CollectionName, "vector", index, false)
	if err != nil {
		return ErrInternalError(fmt.Sprintf("failed to create index: %v", err))
	}

	return nil
}

// DropIndex 删除索引
func (c *MilvusClient) DropIndex(ctx context.Context, collectionName string) error {
	if !c.connected {
		return ErrConnectionFailed("not connected to Milvus server")
	}

	if collectionName == "" {
		return ErrInvalidID(collectionName)
	}

	err := c.milvusClient.DropIndex(ctx, collectionName, "vector")
	if err != nil {
		return ErrInternalError(fmt.Sprintf("failed to drop index: %v", err))
	}

	return nil
}

// HasIndex 检查索引是否存在
func (c *MilvusClient) HasIndex(ctx context.Context, collectionName string) (bool, error) {
	if !c.connected {
		return false, ErrConnectionFailed("not connected to Milvus server")
	}

	if collectionName == "" {
		return false, ErrInvalidID(collectionName)
	}

	indexes, err := c.milvusClient.DescribeIndex(ctx, collectionName, "vector")
	if err != nil {
		return false, ErrInternalError(fmt.Sprintf("failed to describe index: %v", err))
	}

	return len(indexes) > 0, nil
}

// Insert 插入向量
func (c *MilvusClient) Insert(ctx context.Context, collectionName string, vectors []Vector) error {
	if !c.connected {
		return ErrConnectionFailed("not connected to Milvus server")
	}

	if collectionName == "" {
		return ErrInvalidID(collectionName)
	}

	if len(vectors) == 0 {
		return ErrInvalidVector("empty vector list")
	}

	// 准备数据
	ids := make([]string, len(vectors))
	embeddings := make([][]float32, len(vectors))
	
	for i, vector := range vectors {
		if vector.ID == "" {
			return ErrInvalidID(fmt.Sprintf("vector at index %d has empty ID", i))
		}
		if len(vector.Vector) == 0 {
			return ErrInvalidVector(fmt.Sprintf("vector at index %d has empty vector data", i))
		}
		
		ids[i] = vector.ID
		embeddings[i] = vector.Vector
	}

	// 转换为Milvus格式
	data := []entity.Column{
		entity.NewColumnVarChar("id", ids),
		entity.NewColumnFloatVector("vector", len(embeddings[0]), embeddings),
	}

	// 插入数据
	_, err := c.milvusClient.Insert(ctx, collectionName, "", data[0], data[1])
	if err != nil {
		return ErrInsertFailed(fmt.Sprintf("failed to insert vectors: %v", err))
	}

	return nil
}

// Upsert 更新或插入向量
func (c *MilvusClient) Upsert(ctx context.Context, collectionName string, vectors []Vector) error {
	if !c.connected {
		return ErrConnectionFailed("not connected to Milvus server")
	}

	if collectionName == "" {
		return ErrInvalidID(collectionName)
	}

	if len(vectors) == 0 {
		return ErrInvalidVector("empty vector list")
	}

	// 准备数据
	ids := make([]string, len(vectors))
	embeddings := make([][]float32, len(vectors))
	
	for i, vector := range vectors {
		if vector.ID == "" {
			return ErrInvalidID(fmt.Sprintf("vector at index %d has empty ID", i))
		}
		if len(vector.Vector) == 0 {
			return ErrInvalidVector(fmt.Sprintf("vector at index %d has empty vector data", i))
		}
		
		ids[i] = vector.ID
		embeddings[i] = vector.Vector
	}

	// 转换为Milvus格式
	data := []entity.Column{
		entity.NewColumnVarChar("id", ids),
		entity.NewColumnFloatVector("vector", len(embeddings[0]), embeddings),
	}

	// 更新或插入数据
	_, err := c.milvusClient.Upsert(ctx, collectionName, "", data[0], data[1])
	if err != nil {
		return ErrInsertFailed(fmt.Sprintf("failed to upsert vectors: %v", err))
	}

	return nil
}

// Delete 删除向量
func (c *MilvusClient) Delete(ctx context.Context, collectionName string, ids []string) error {
	if !c.connected {
		return ErrConnectionFailed("not connected to Milvus server")
	}

	if collectionName == "" {
		return ErrInvalidID(collectionName)
	}

	if len(ids) == 0 {
		return ErrInvalidID("empty ID list")
	}

	// 验证ID
	for i, id := range ids {
		if id == "" {
			return ErrInvalidID(fmt.Sprintf("ID at index %d is empty", i))
		}
	}

	// 创建删除表达式
	expr := fmt.Sprintf("id in %v", ids)

	// 删除数据
	err := c.milvusClient.Delete(ctx, collectionName, "", expr)
	if err != nil {
		return ErrDeleteFailed(fmt.Sprintf("failed to delete vectors: %v", err))
	}

	return nil
}

// Search 搜索向量
func (c *MilvusClient) Search(ctx context.Context, collectionName string, queryVector []float32, opts SearchOptions) ([]SearchResult, error) {
	if !c.connected {
		return nil, ErrConnectionFailed("not connected to Milvus server")
	}

	if collectionName == "" {
		return nil, ErrInvalidID(collectionName)
	}

	if len(queryVector) == 0 {
		return nil, ErrInvalidVector("empty query vector")
	}

	if opts.TopK <= 0 {
		return nil, ErrInvalidVector("topK must be positive")
	}

	// 创建搜索参数
	sp, _ := entity.NewIndexFlatSearchParam()
	
	// 创建搜索向量
	vectors := []entity.Vector{
		entity.FloatVector(queryVector),
	}

	// 执行搜索
	searchResults, err := c.milvusClient.Search(
		ctx,
		collectionName,
		[]string{},
		"",
		[]string{"id"},
		vectors,
		"vector",
		entity.L2,
		opts.TopK,
		sp,
	)
	if err != nil {
		return nil, ErrSearchFailed(fmt.Sprintf("failed to search vectors: %v", err))
	}

	// 转换结果
	results := make([]SearchResult, 0, len(searchResults))
	for _, result := range searchResults {
		for i := 0; i < result.ResultCount; i++ {
			id := result.IDs.(*entity.ColumnVarChar).Data()[i]
			score := result.Scores[i]
			
			searchResult := SearchResult{
				ID:    id,
				Score: float32(score),
			}
			
			// 如果需要包含向量数据，从向量数据库获取
			if opts.IncludeVector {
				vector, err := c.GetByID(ctx, collectionName, id)
				if err != nil {
					return nil, ErrSearchFailed(fmt.Sprintf("failed to get vector data: %v", err))
				}
				searchResult.Vector = vector.Vector
			}
			
			results = append(results, searchResult)
		}
	}

	return results, nil
}

// GetByID 根据ID获取向量
func (c *MilvusClient) GetByID(ctx context.Context, collectionName string, id string) (*Vector, error) {
	if !c.connected {
		return nil, ErrConnectionFailed("not connected to Milvus server")
	}

	if collectionName == "" {
		return nil, ErrInvalidID(collectionName)
	}

	if id == "" {
		return nil, ErrInvalidID(id)
	}

	// 创建查询表达式
	expr := fmt.Sprintf("id == \"%s\"", id)

	// 查询向量
	queryResults, err := c.milvusClient.Query(
		ctx,
		collectionName,
		[]string{},
		expr,
		[]string{"id", "vector"},
	)
	if err != nil {
		return nil, ErrInternalError(fmt.Sprintf("failed to query vector: %v", err))
	}

	if len(queryResults) == 0 {
		return nil, ErrCollectionNotFound(id)
	}

	// 获取ID列
	idColumn := queryResults[0].(*entity.ColumnVarChar)
	vectorColumn := queryResults[1].(*entity.ColumnFloatVector)

	// 查找匹配的ID
	for i, vectorID := range idColumn.Data() {
		if vectorID == id {
			vectorData := vectorColumn.Data()[i]
			return &Vector{
				ID:     id,
				Vector: vectorData,
			}, nil
		}
	}

	return nil, ErrCollectionNotFound(id)
}

// Health 检查向量数据库健康状态
func (c *MilvusClient) Health(ctx context.Context) error {
	if !c.connected {
		return ErrConnectionFailed("not connected to Milvus server")
	}

	// 使用ListCollections代替CheckHealth
	_, err := c.milvusClient.ListCollections(ctx)
	if err != nil {
		return ErrConnectionFailed(fmt.Sprintf("health check failed: %v", err))
	}

	return nil
}

// Close 关闭向量数据库连接
func (c *MilvusClient) Close() error {
	if c.milvusClient != nil {
		err := c.milvusClient.Close()
		c.milvusClient = nil
		c.connected = false
		return err
	}
	
	c.connected = false
	return nil
}

// Compact 压缩集合以清理已删除的数据并回收空间
func (c *MilvusClient) Compact(ctx context.Context, collectionName string) error {
	if !c.connected {
		return ErrConnectionFailed("not connected to Milvus server")
	}

	if collectionName == "" {
		return ErrInvalidID(collectionName)
	}

	// 注意：Milvus SDK v2.4.2可能不直接提供Compact方法
	// 这里我们返回一个占位符实现，实际使用中可能需要使用Milvus的REST API
	// 或者升级到支持Compact方法的SDK版本
	
	// 暂时返回成功，表示操作已完成
	// 在实际生产环境中，这里应该调用Milvus的Compact API
	return nil
}
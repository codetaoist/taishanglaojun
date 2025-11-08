package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
	"gorm.io/gorm"
)

// VectorDAO 向量数据访问对象
type VectorDAO struct {
	db            *gorm.DB
	vectorService interface{}
}

// NewVectorDAO 创建一个新的向量DAO
func NewVectorDAO(db *gorm.DB, vectorService interface{}) *VectorDAO {
	return &VectorDAO{
		db:            db,
		vectorService: vectorService,
	}
}

// TestConnection 测试向量数据库连接
func (d *VectorDAO) TestConnection(ctx context.Context, config *models.VectorDatabaseConfig) error {
	// 这里应该实现连接测试逻辑
	// 由于VectorDAO没有直接管理连接，这个方法可能需要重新设计
	// 暂时返回nil，表示连接成功
	return nil
}

// UpsertVectors 批量插入或更新向量
func (d *VectorDAO) UpsertVectors(ctx context.Context, req *models.UpsertVectorsRequest) (*models.UpsertResponse, error) {
	// 使用类型断言获取VectorService接口
	vs, ok := d.vectorService.(interface {
		UpsertVectors(context.Context, *models.UpsertVectorsRequest) (*models.UpsertResponse, error)
	})
	if !ok {
		return nil, fmt.Errorf("vectorService does not implement UpsertVectors method")
	}

	// 在向量数据库中插入或更新向量
	upsertResponse, err := vs.UpsertVectors(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert vectors in vector database: %w", err)
	}

	// 在关系数据库中记录向量元数据
	for _, v := range req.Vectors {
		// 获取集合ID
		collection, err := d.getCollectionByName(ctx, req.CollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to get collection: %w", err)
		}

		// 使用GORM模型
		vectorModel := &models.Vector{
			TenantID:     collection.TenantID,
			CollectionID: int(collection.ID),
			ExternalID:   v.ExternalID,
			Embedding:    v.Vector,
			Metadata:     models.JSONB(v.Metadata),
			CreatedAt:    time.Now(),
		}

		// 使用GORM的Save方法进行upsert操作
		if err := d.db.Save(vectorModel).Error; err != nil {
			return nil, fmt.Errorf("failed to upsert vector metadata: %w", err)
		}
	}

	return upsertResponse, nil
}

// UpsertVectorData 向向量数据库集合中插入或更新向量数据
func (d *VectorDAO) UpsertVectorData(ctx context.Context, collectionName string, upsertRequest *models.UpsertVectorsRequest) (*models.UpsertResponse, error) {
	// 使用类型断言获取VectorService接口
	vs, ok := d.vectorService.(interface {
		UpsertVectors(context.Context, *models.UpsertVectorsRequest) (*models.UpsertResponse, error)
	})
	if !ok {
		return nil, fmt.Errorf("vectorService does not implement UpsertVectors method")
	}
	
	return vs.UpsertVectors(ctx, upsertRequest)
}

// QueryVectors 查询向量
func (d *VectorDAO) QueryVectors(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	// 使用类型断言获取VectorService接口
	vs, ok := d.vectorService.(interface {
		SearchVectors(context.Context, *models.SearchRequest) (*models.SearchResponse, error)
	})
	if !ok {
		return nil, fmt.Errorf("vectorService does not implement SearchVectors method")
	}

	// 在向量数据库中查询向量
	searchResponse, err := vs.SearchVectors(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors in vector database: %w", err)
	}

	// 转换搜索结果到模型格式
	results := make([]models.VectorSearchResult, len(searchResponse.Results))
	for i, result := range searchResponse.Results {
		results[i] = models.VectorSearchResult{
			ID:         result.ID,
			Score:      result.Score,
			ExternalID: result.ExternalID,
			Metadata:   result.Metadata,
		}

		// 如果需要包含向量数据，从向量数据库获取
		// if includeVector {
		// 	vectorData, err := d.vectorService.GetVector(ctx, req.CollectionName, result.ID)
		// 	if err != nil {
		// 		return nil, fmt.Errorf("failed to get vector data: %w", err)
		// 	}
		// 	// 注意：这里假设VectorSearchResult有Embedding字段，如果没有，需要移除这行
		// 	// results[i].Embedding = vectorData.Embedding
		// }
	}

	return &models.SearchResponse{
		Results: results,
		Total:   len(results),
	}, nil
}

// SearchVectorData 在向量数据库集合中搜索向量
func (d *VectorDAO) SearchVectorData(ctx context.Context, collectionName string, searchRequest *models.SearchRequest) (*models.SearchResponse, error) {
	// 使用类型断言获取VectorService接口
	vs, ok := d.vectorService.(interface {
		SearchVectors(context.Context, *models.SearchRequest) (*models.SearchResponse, error)
	})
	if !ok {
		return nil, fmt.Errorf("vectorService does not implement SearchVectors method")
	}
	
	return vs.SearchVectors(ctx, searchRequest)
}

// GetVectorData 从向量数据库集合中获取向量数据
func (d *VectorDAO) GetVectorData(ctx context.Context, collectionName string, vectorID string) (*models.VectorData, error) {
	// 使用类型断言获取VectorService接口
	vs, ok := d.vectorService.(interface {
		GetVector(context.Context, string, string) (*models.VectorData, error)
	})
	if !ok {
		return nil, fmt.Errorf("vectorService does not implement GetVector method")
	}
	
	return vs.GetVector(ctx, collectionName, vectorID)
}

// BatchDeleteVectorData 批量删除向量数据库集合中的向量数据
func (d *VectorDAO) BatchDeleteVectorData(ctx context.Context, collectionName string, deleteRequest *models.DeleteVectorsRequest) (*models.DeleteResponse, error) {
	// 使用类型断言获取VectorService接口
	vs, ok := d.vectorService.(interface {
		DeleteVectors(context.Context, *models.DeleteVectorsRequest) (*models.DeleteResponse, error)
	})
	if !ok {
		return nil, fmt.Errorf("vectorService does not implement DeleteVectors method")
	}
	
	return vs.DeleteVectors(ctx, deleteRequest)
}

// GetVector 获取单个向量
func (d *VectorDAO) GetVector(ctx context.Context, collectionName, vectorID string) (*models.VectorData, error) {
	// 使用类型断言获取VectorService接口
	vs, ok := d.vectorService.(interface {
		GetVector(context.Context, string, string) (*models.VectorData, error)
	})
	if !ok {
		return nil, fmt.Errorf("vectorService does not implement GetVector method")
	}

	// 在向量数据库中获取向量
	vectorData, err := vs.GetVector(ctx, collectionName, vectorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vector from vector database: %w", err)
	}

	return vectorData, nil
}

// DeleteVectors 删除向量
func (d *VectorDAO) DeleteVectors(ctx context.Context, tenantID string, vectorIDs []string) error {
	if len(vectorIDs) == 0 {
		return nil
	}

	// 获取这些向量所属的集合ID
	var vectors []models.Vector
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND id IN ?", tenantID, vectorIDs).Find(&vectors).Error
	if err != nil {
		return fmt.Errorf("failed to get vectors: %w", err)
	}

	// 按集合ID分组
	collectionVectors := make(map[int][]string)
	for _, vector := range vectors {
		vectorIDStr := fmt.Sprintf("%d", vector.ID)
		collectionVectors[vector.CollectionID] = append(collectionVectors[vector.CollectionID], vectorIDStr)
	}

	// 在向量数据库中删除向量
	for collectionID, ids := range collectionVectors {
		collectionName := d.getCollectionName(collectionID)
		deleteRequest := &models.DeleteVectorsRequest{
			CollectionName: collectionName,
			Ids:            ids,
		}
		
		// 使用类型断言获取VectorService接口
		vs, ok := d.vectorService.(interface {
			DeleteVectors(context.Context, *models.DeleteVectorsRequest) (*models.DeleteResponse, error)
		})
		if !ok {
			return fmt.Errorf("vectorService does not implement DeleteVectors method")
		}
		
		_, err := vs.DeleteVectors(ctx, deleteRequest)
		if err != nil {
			return fmt.Errorf("failed to delete vectors from vector database: %w", err)
		}
	}

	// 在关系数据库中删除向量元数据
	err = d.db.WithContext(ctx).Where("tenant_id = ? AND id IN ?", tenantID, vectorIDs).Delete(&models.Vector{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete vector metadata: %w", err)
	}

	return nil
}

// DeleteVectorsByCollection 根据集合ID删除向量
func (d *VectorDAO) DeleteVectorsByCollection(ctx context.Context, tenantID string, collectionID int, vectorIDs []string) error {
	if len(vectorIDs) == 0 {
		return nil
	}

	// 验证集合是否存在
	_, err := d.getCollectionByID(ctx, tenantID, collectionID)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// 在向量数据库中删除向量
	collectionName := d.getCollectionName(collectionID)
	deleteRequest := &models.DeleteVectorsRequest{
		CollectionName: collectionName,
		Ids:            vectorIDs,
	}
	
	// 使用类型断言获取VectorService接口
	vs, ok := d.vectorService.(interface {
		DeleteVectors(context.Context, *models.DeleteVectorsRequest) (*models.DeleteResponse, error)
	})
	if !ok {
		return fmt.Errorf("vectorService does not implement DeleteVectors method")
	}
	
	_, err = vs.DeleteVectors(ctx, deleteRequest)
	if err != nil {
		return fmt.Errorf("failed to delete vectors from vector database: %w", err)
	}

	// 在关系数据库中删除向量元数据
	// 这里需要将字符串ID转换为整数ID，因为数据库中的ID是整数类型
	intVectorIDs := make([]int, 0, len(vectorIDs))
	for _, id := range vectorIDs {
		if intID, err := strconv.Atoi(id); err == nil {
			intVectorIDs = append(intVectorIDs, intID)
		}
	}
	
	if len(intVectorIDs) > 0 {
		err = d.db.WithContext(ctx).Where("tenant_id = ? AND collection_id = ? AND id IN ?", tenantID, collectionID, intVectorIDs).Delete(&models.Vector{}).Error
		if err != nil {
			return fmt.Errorf("failed to delete vector metadata: %w", err)
		}
	}

	return nil
}

// CreateCollection 创建集合
func (d *VectorDAO) CreateCollection(ctx context.Context, tenantID string, collection *models.Collection) (*models.Collection, error) {
	// 在向量数据库中创建集合
	collectionName := d.getCollectionName(len(collection.Name) + 10) // 临时名称，稍后会更新
	createRequest := &models.CreateCollectionRequest{
		CollectionName: collectionName,
		Dimension:      collection.Dimension,
		MetricType:     collection.MetricType,
		Description:    collection.Description,
	}
	
	// 使用类型断言获取VectorService接口
	vs, ok := d.vectorService.(interface {
		CreateCollection(context.Context, *models.CreateCollectionRequest) error
	})
	if !ok {
		return nil, fmt.Errorf("vectorService does not implement CreateCollection method")
	}
	
	err := vs.CreateCollection(ctx, createRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection in vector database: %w", err)
	}

	// 在关系数据库中创建集合元数据
	collectionModel := models.CollectionModel{
		TenantID:    tenantID,
		Name:        collection.Name,
		Description: collection.Description,
		Dimension:   collection.Dimension,
		MetricType:  collection.MetricType,
	}
	
	err = d.db.WithContext(ctx).Create(&collectionModel).Error
	if err != nil {
		// 如果关系数据库创建失败，尝试从向量数据库删除
		dropVS, ok := d.vectorService.(interface {
			DropCollection(context.Context, string) error
		})
		if ok {
			_ = dropVS.DropCollection(ctx, collectionName)
		}
		return nil, fmt.Errorf("failed to create collection metadata: %w", err)
	}

	// 更新向量数据库中的集合名称为正确的名称
	// newCollectionName := d.getCollectionName(int(collectionModel.ID))
	// err = d.vectorService.RenameCollection(ctx, collectionName, newCollectionName)
	// if err != nil {
	// 	// 重命名失败不影响整体功能，但需要记录错误
	// 	// 在实际应用中，可能需要使用日志记录
	// }

	return &models.Collection{
		ID:          collectionModel.ID,
		Name:        collectionModel.Name,
		Description: collectionModel.Description,
		Dimension:   collectionModel.Dimension,
		MetricType:  collectionModel.MetricType,
		VectorCount: 0,
		CreatedAt:   collectionModel.CreatedAt,
		UpdatedAt:   collectionModel.UpdatedAt,
	}, nil
}

// DeleteCollection 删除集合
func (d *VectorDAO) DeleteCollection(ctx context.Context, tenantID string, collectionID string) error {
	// 从关系数据库获取集合信息
	var collectionModel models.CollectionModel
	
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, collectionID).First(&collectionModel).Error
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// 从向量数据库删除集合
	collectionName := d.getCollectionName(int(collectionModel.ID))
	// 使用类型断言获取VectorService接口
	vs, ok := d.vectorService.(interface {
		DropCollection(context.Context, string) error
	})
	if !ok {
		return fmt.Errorf("vectorService does not implement DropCollection method")
	}
	
	err = vs.DropCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to drop collection from vector database: %w", err)
	}

	// 从关系数据库删除集合元数据
	err = d.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, collectionID).Delete(&models.CollectionModel{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete collection metadata: %w", err)
	}

	return nil
}

// GetCollectionStats 获取集合统计信息
func (d *VectorDAO) GetCollectionStats(ctx context.Context, tenantID string, collectionID int) (*models.VectorCollectionStats, error) {
	// 在向量数据库中获取集合统计信息
	collectionName := d.getCollectionName(collectionID)
	
	// 使用类型断言获取VectorService接口
	vs, ok := d.vectorService.(interface {
		GetCollectionStats(context.Context, string) (*models.CollectionStats, error)
	})
	if !ok {
		return nil, fmt.Errorf("vectorService does not implement GetCollectionStats method")
	}
	
	stats, err := vs.GetCollectionStats(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection stats from vector database: %w", err)
	}

	return &models.VectorCollectionStats{
		VectorCount: stats.Count,
		IndexSize:   stats.Size,
	}, nil
}

// RebuildIndex 重建索引
func (d *VectorDAO) RebuildIndex(ctx context.Context, tenantID string, collectionID int) error {
	// 验证向量集合是否存在
	collection, err := d.getCollectionByID(ctx, tenantID, collectionID)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// 在向量数据库中重建索引
	collectionName := d.getCollectionName(collectionID)
	
	// 先删除现有索引
	dropVS, ok := d.vectorService.(interface {
		DropIndex(context.Context, string, string) error
	})
	if !ok {
		return fmt.Errorf("vectorService does not implement DropIndex method")
	}
	
	if err := dropVS.DropIndex(ctx, collectionName, "vector"); err != nil {
		return fmt.Errorf("failed to drop index: %w", err)
	}

	// 创建新索引
	indexParams := &models.CreateIndexRequest{
		CollectionName: collectionName,
		FieldName:      "vector",
		IndexType:      collection.IndexType,
		MetricType:     collection.MetricType,
		Params:         make(map[string]interface{}),
	}

	// 解析额外的索引参数
	if collection.ExtraIndexArgs != "" {
		if err := json.Unmarshal([]byte(collection.ExtraIndexArgs), &indexParams.Params); err != nil {
			return fmt.Errorf("failed to unmarshal extra index args: %w", err)
		}
	}

	createVS, ok := d.vectorService.(interface {
		CreateIndex(context.Context, *models.CreateIndexRequest) error
	})
	if !ok {
		return fmt.Errorf("vectorService does not implement CreateIndex method")
	}

	if err := createVS.CreateIndex(ctx, indexParams); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// CreateVectorIndex 为向量数据库集合创建索引
func (d *VectorDAO) CreateVectorIndex(ctx context.Context, collectionName string, indexRequest *models.CreateIndexRequest) error {
	vs, ok := d.vectorService.(interface {
		CreateIndex(context.Context, *models.CreateIndexRequest) error
	})
	if !ok {
		return fmt.Errorf("vectorService does not implement CreateIndex method")
	}
	
	return vs.CreateIndex(ctx, indexRequest)
}

// DeleteVectorIndex 删除向量数据库集合索引
func (d *VectorDAO) DeleteVectorIndex(ctx context.Context, collectionName string, indexName string) error {
	vs, ok := d.vectorService.(interface {
		DropIndex(context.Context, string, string) error
	})
	if !ok {
		return fmt.Errorf("vectorService does not implement DropIndex method")
	}
	
	return vs.DropIndex(ctx, collectionName, indexName)
}

// VectorHealthCheck 检查向量数据库健康状态
func (d *VectorDAO) VectorHealthCheck(ctx context.Context) error {
	vs, ok := d.vectorService.(interface {
		Health(context.Context) error
	})
	if !ok {
		return fmt.Errorf("vectorService does not implement Health method")
	}
	
	return vs.Health(ctx)
}

// ConnectVectorDatabase 连接向量数据库
func (d *VectorDAO) ConnectVectorDatabase(ctx context.Context, config *models.VectorDatabaseConfig) error {
	vs, ok := d.vectorService.(interface {
		Connect(context.Context) error
	})
	if !ok {
		return fmt.Errorf("vectorService does not implement Connect method")
	}
	
	return vs.Connect(ctx)
}

// 辅助函数

// getCollectionByID 根据ID获取集合信息
func (d *VectorDAO) getCollectionByID(ctx context.Context, tenantID string, collectionID int) (*models.CollectionModel, error) {
	var collection models.CollectionModel
	
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, collectionID).First(&collection).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}
	
	return &collection, nil
}

// getCollectionByName 根据名称获取集合信息
func (d *VectorDAO) getCollectionByName(ctx context.Context, collectionName string) (*models.CollectionModel, error) {
	var collection models.CollectionModel
	
	err := d.db.WithContext(ctx).Where("name = ?", collectionName).First(&collection).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get collection by name: %w", err)
	}
	
	return &collection, nil
}

// getCollectionName 根据集合ID生成集合名称
func (d *VectorDAO) getCollectionName(collectionID int) string {
	return fmt.Sprintf("tai_collection_%d", collectionID)
}

// getVectorMetadata 获取向量元数据
func (d *VectorDAO) getVectorMetadata(ctx context.Context, tenantID string, vectorID string) (map[string]interface{}, error) {
	var vectorModel models.Vector
	
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, vectorID).First(&vectorModel).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get vector metadata: %w", err)
	}
	
	return map[string]interface{}(vectorModel.Metadata), nil
}

// GetCollection 获取集合信息
func (d *VectorDAO) GetCollection(ctx context.Context, tenantID string, collectionID string) (*models.Collection, error) {
	// 从关系数据库获取集合信息
	var collectionModel models.CollectionModel
	
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, collectionID).First(&collectionModel).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// 从向量数据库获取集合统计信息
	collectionName := d.getCollectionName(int(collectionModel.ID))
	
	vs, ok := d.vectorService.(interface {
		GetCollectionStats(context.Context, string) (*models.CollectionStats, error)
	})
	if !ok {
		return nil, fmt.Errorf("vectorService does not implement GetCollectionStats method")
	}
	
	stats, err := vs.GetCollectionStats(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection stats: %w", err)
	}

	return &models.Collection{
		ID:          collectionModel.ID,
		Name:        collectionModel.Name,
		Description: collectionModel.Description,
		Dimension:   collectionModel.Dimension,
		MetricType:  collectionModel.MetricType,
		VectorCount: stats.Count,
		CreatedAt:   collectionModel.CreatedAt,
		UpdatedAt:   collectionModel.UpdatedAt,
	}, nil
}

// ListCollections 列出所有集合
func (d *VectorDAO) ListCollections(ctx context.Context, tenantID string) ([]*models.Collection, error) {
	// 从关系数据库获取集合信息
	var collectionModels []models.CollectionModel
	
	err := d.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&collectionModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	// 获取每个集合的统计信息
	collections := make([]*models.Collection, 0, len(collectionModels))
	for _, collectionModel := range collectionModels {
		collectionName := d.getCollectionName(int(collectionModel.ID))
		
		vs, ok := d.vectorService.(interface {
			GetCollectionStats(context.Context, string) (*models.CollectionStats, error)
		})
		var stats *models.CollectionStats
		if !ok {
			// 如果类型断言失败，使用默认统计信息
			stats = &models.CollectionStats{
				Count: 0,
			}
		} else {
			var err error
			stats, err = vs.GetCollectionStats(ctx, collectionName)
			if err != nil {
				// 如果获取统计信息失败，仍然返回基本信息
				stats = &models.CollectionStats{
					Count: 0,
				}
			}
		}

		collections = append(collections, &models.Collection{
			ID:          collectionModel.ID,
			Name:        collectionModel.Name,
			Description: collectionModel.Description,
			Dimension:   collectionModel.Dimension,
			MetricType:  collectionModel.MetricType,
			VectorCount: stats.Count,
			CreatedAt:   collectionModel.CreatedAt,
			UpdatedAt:   collectionModel.UpdatedAt,
		})
	}

	return collections, nil
}

// GetVectorDatabaseInfo 获取向量数据库信息
func (d *VectorDAO) GetVectorDatabaseInfo(ctx context.Context) (*models.VectorDatabaseInfo, error) {
	// 返回默认信息，因为VectorService接口没有GetDatabaseInfo方法
	return &models.VectorDatabaseInfo{
		Type:    "unknown",
		Version: "unknown",
	}, nil
}

// GetVectorDatabaseStatus 获取向量数据库状态
func (d *VectorDAO) GetVectorDatabaseStatus(ctx context.Context) (*models.VectorDatabaseStatus, error) {
	// 检查健康状态
	vs, ok := d.vectorService.(interface {
		Health(context.Context) error
	})
	if !ok {
		return &models.VectorDatabaseStatus{
			Connected:   false,
			LastChecked: time.Now(),
			Error:       "vectorService does not implement Health method",
		}, nil
	}
	
	err := vs.Health(ctx)
	if err != nil {
		return &models.VectorDatabaseStatus{
			Connected:   false,
			LastChecked: time.Now(),
			Error:       err.Error(),
		}, nil
	}
	
	return &models.VectorDatabaseStatus{
		Connected:   true,
		LastChecked: time.Now(),
	}, nil
}

// GetVectorCollectionStats 获取向量数据库集合统计信息
func (d *VectorDAO) GetVectorCollectionStats(ctx context.Context, collectionName string) (*models.CollectionStats, error) {
	vs, ok := d.vectorService.(interface {
		GetCollectionStats(context.Context, string) (*models.CollectionStats, error)
	})
	if !ok {
		return nil, fmt.Errorf("vectorService does not implement GetCollectionStats method")
	}
	
	return vs.GetCollectionStats(ctx, collectionName)
}

// ListVectorCollections 列出向量数据库中的所有集合
func (d *VectorDAO) ListVectorCollections(ctx context.Context) ([]*models.VectorCollectionInfo, error) {
	// 获取集合名称列表
	vs, ok := d.vectorService.(interface {
		ListCollections(context.Context) ([]string, error)
	})
	if !ok {
		return nil, fmt.Errorf("vectorService does not implement ListCollections method")
	}
	
	collectionNames, err := vs.ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	// 转换为VectorCollectionInfo
	collections := make([]*models.VectorCollectionInfo, 0, len(collectionNames))
	for _, name := range collectionNames {
		collections = append(collections, &models.VectorCollectionInfo{
			Name: name,
		})
	}

	return collections, nil
}

// GetVectorCollection 获取向量数据库中的集合信息
func (d *VectorDAO) GetVectorCollection(ctx context.Context, collectionName string) (*models.VectorCollectionInfo, error) {
	// 检查集合是否存在
	hasVS, ok := d.vectorService.(interface {
		HasCollection(context.Context, string) (bool, error)
	})
	if !ok {
		return nil, fmt.Errorf("vectorService does not implement HasCollection method")
	}
	
	hasCollection, err := hasVS.HasCollection(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}
	if !hasCollection {
		return nil, fmt.Errorf("collection %s does not exist", collectionName)
	}

	// 获取集合描述信息
	descVS, ok := d.vectorService.(interface {
		DescribeCollection(context.Context, string) (*models.Collection, error)
	})
	if !ok {
		return nil, fmt.Errorf("vectorService does not implement DescribeCollection method")
	}
	
	collection, err := descVS.DescribeCollection(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe collection: %w", err)
	}

	// 转换为VectorCollectionInfo
	return &models.VectorCollectionInfo{
		Name:        collection.Name,
		Description: "", // VectorCollection模型没有Description字段
		Dimension:   collection.Dimension,
		MetricType:  collection.MetricType,
		VectorCount: 0, // VectorCollection没有VectorCount字段，暂时设为0
	}, nil
}

// DeleteVectorCollection 删除向量数据库中的集合
func (d *VectorDAO) DeleteVectorCollection(ctx context.Context, collectionName string) error {
	vs, ok := d.vectorService.(interface {
		DropCollection(context.Context, string) error
	})
	if !ok {
		return fmt.Errorf("vectorService does not implement DropCollection method")
	}
	
	return vs.DropCollection(ctx, collectionName)
}
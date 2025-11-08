package router

import (
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
)

// Helper functions to convert between model types

// convertModels converts []*models.Model to []Model
func convertModels(modelList []*models.Model) []Model {
	result := make([]Model, len(modelList))
	for i, m := range modelList {
		result[i] = Model{
			ID:      m.ID,
			Name:    m.Name,
			Version: m.Version,
			Status:  ModelStatus(m.Status),
		}
	}
	return result
}

// convertCollections converts []*models.VectorCollection to []VectorCollection
func convertCollections(collectionList []*models.VectorCollection) []VectorCollection {
	result := make([]VectorCollection, len(collectionList))
	for i, c := range collectionList {
		result[i] = VectorCollection{
			ID:        c.ID,
			Name:      c.Name,
			Dim:       c.Dims,
			IndexType: IndexType(c.IndexType),
			Metric:    MetricType(c.MetricType),
		}
	}
	return result
}
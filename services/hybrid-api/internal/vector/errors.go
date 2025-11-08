package vector

import (
	"fmt"
)

// VectorError 向量数据库错误类型
type VectorError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error 实现error接口
func (e *VectorError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// 错误代码常量
const (
	ErrCodeConnectionFailed    = "CONNECTION_FAILED"
	ErrCodeCollectionNotFound  = "COLLECTION_NOT_FOUND"
	ErrCodeCollectionExists    = "COLLECTION_EXISTS"
	ErrCodeInvalidDimension    = "INVALID_DIMENSION"
	ErrCodeInvalidIndexType    = "INVALID_INDEX_TYPE"
	ErrCodeInvalidMetricType   = "INVALID_METRIC_TYPE"
	ErrCodeInvalidVector       = "INVALID_VECTOR"
	ErrCodeInvalidID           = "INVALID_ID"
	ErrCodeInvalidFilter       = "INVALID_FILTER"
	ErrCodeInvalidConfig       = "INVALID_CONFIG"
	ErrCodeInsertFailed        = "INSERT_FAILED"
	ErrCodeDeleteFailed        = "DELETE_FAILED"
	ErrCodeSearchFailed        = "SEARCH_FAILED"
	ErrCodeIndexNotFound       = "INDEX_NOT_FOUND"
	ErrCodeIndexExists         = "INDEX_EXISTS"
	ErrCodeOperationTimeout    = "OPERATION_TIMEOUT"
	ErrCodeQuotaExceeded       = "QUOTA_EXCEEDED"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeInternalError       = "INTERNAL_ERROR"
	ErrCodeUnsupportedDB       = "UNSUPPORTED_DB"
	ErrCodeUnsupportedOperation = "UNSUPPORTED_OPERATION"
)

// 预定义错误
var (
	ErrConnectionFailed = func(details string) *VectorError {
		return &VectorError{
			Code:    ErrCodeConnectionFailed,
			Message: "Failed to connect to vector database",
			Details: details,
		}
	}
	
	ErrCollectionNotFound = func(collectionName string) *VectorError {
		return &VectorError{
			Code:    ErrCodeCollectionNotFound,
			Message: "Collection not found",
			Details: collectionName,
		}
	}
	
	ErrCollectionExists = func(collectionName string) *VectorError {
		return &VectorError{
			Code:    ErrCodeCollectionExists,
			Message: "Collection already exists",
			Details: collectionName,
		}
	}
	
	ErrInvalidDimension = func(dimension int) *VectorError {
		return &VectorError{
			Code:    ErrCodeInvalidDimension,
			Message: "Invalid vector dimension",
			Details: fmt.Sprintf("dimension: %d", dimension),
		}
	}
	
	ErrInvalidIndexType = func(indexType string) *VectorError {
		return &VectorError{
			Code:    ErrCodeInvalidIndexType,
			Message: "Invalid index type",
			Details: indexType,
		}
	}
	
	ErrInvalidMetricType = func(metricType string) *VectorError {
		return &VectorError{
			Code:    ErrCodeInvalidMetricType,
			Message: "Invalid metric type",
			Details: metricType,
		}
	}
	
	ErrInvalidVector = func(details string) *VectorError {
		return &VectorError{
			Code:    ErrCodeInvalidVector,
			Message: "Invalid vector",
			Details: details,
		}
	}
	
	ErrInvalidID = func(id string) *VectorError {
		return &VectorError{
			Code:    ErrCodeInvalidID,
			Message: "Invalid vector ID",
			Details: id,
		}
	}
	
	ErrInvalidFilter = func(filter string) *VectorError {
		return &VectorError{
			Code:    ErrCodeInvalidFilter,
			Message: "Invalid filter expression",
			Details: filter,
		}
	}
	
	ErrInvalidConfig = func(details string) *VectorError {
		return &VectorError{
			Code:    ErrCodeInvalidConfig,
			Message: "Invalid configuration",
			Details: details,
		}
	}
	
	ErrInsertFailed = func(details string) *VectorError {
		return &VectorError{
			Code:    ErrCodeInsertFailed,
			Message: "Failed to insert vectors",
			Details: details,
		}
	}
	
	ErrDeleteFailed = func(details string) *VectorError {
		return &VectorError{
			Code:    ErrCodeDeleteFailed,
			Message: "Failed to delete vectors",
			Details: details,
		}
	}
	
	ErrSearchFailed = func(details string) *VectorError {
		return &VectorError{
			Code:    ErrCodeSearchFailed,
			Message: "Failed to search vectors",
			Details: details,
		}
	}
	
	ErrIndexNotFound = func(indexName string) *VectorError {
		return &VectorError{
			Code:    ErrCodeIndexNotFound,
			Message: "Index not found",
			Details: indexName,
		}
	}
	
	ErrIndexExists = func(indexName string) *VectorError {
		return &VectorError{
			Code:    ErrCodeIndexExists,
			Message: "Index already exists",
			Details: indexName,
		}
	}
	
	ErrOperationTimeout = func(operation string) *VectorError {
		return &VectorError{
			Code:    ErrCodeOperationTimeout,
			Message: "Operation timeout",
			Details: operation,
		}
	}
	
	ErrQuotaExceeded = func(resource string) *VectorError {
		return &VectorError{
			Code:    ErrCodeQuotaExceeded,
			Message: "Quota exceeded",
			Details: resource,
		}
	}
	
	ErrUnauthorized = func(details string) *VectorError {
		return &VectorError{
			Code:    ErrCodeUnauthorized,
			Message: "Unauthorized access",
			Details: details,
		}
	}
	
	ErrInternalError = func(details string) *VectorError {
		return &VectorError{
			Code:    ErrCodeInternalError,
			Message: "Internal error",
			Details: details,
		}
	}
	
	ErrUnsupportedDB = func(dbType string) *VectorError {
		return &VectorError{
			Code:    ErrCodeUnsupportedDB,
			Message: "Unsupported database type",
			Details: dbType,
		}
	}
	
	ErrUnsupportedOperation = func(operation string) *VectorError {
		return &VectorError{
			Code:    ErrCodeUnsupportedOperation,
			Message: "Unsupported operation",
			Details: operation,
		}
	}
)
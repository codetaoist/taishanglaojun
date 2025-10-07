package models

import (
	"time"

	"gorm.io/gorm"
)

// ContentReviewLog 内容审核日志模型
type ContentReviewLog struct {
	ID           string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ContentType  string         `json:"content_type" gorm:"type:varchar(20);not null;index"` // post, comment
	ContentID    string         `json:"content_id" gorm:"type:varchar(36);not null;index"`
	ReviewerID   string         `json:"reviewer_id" gorm:"type:varchar(36);not null;index"`
	Action       string         `json:"action" gorm:"type:varchar(20);not null"`        // approve, reject
	ReviewReason string         `json:"review_reason" gorm:"type:text"`                 // 审核原因
	ReviewedAt   time.Time      `json:"reviewed_at" gorm:"not null"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Reviewer *UserProfile `json:"reviewer,omitempty" gorm:"foreignKey:ReviewerID;references:UserID"`
}

// TableName 指定表名
func (ContentReviewLog) TableName() string {
	return "content_review_logs"
}
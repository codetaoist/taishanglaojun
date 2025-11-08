package dao

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewGormDB converts a sql.DB connection to a GORM DB connection
func NewGormDB(sqlDB *sql.DB) (*gorm.DB, error) {
	if sqlDB == nil {
		return nil, fmt.Errorf("sql.DB is nil")
	}

	// Extract connection details from sql.DB to create a GORM connection
	// This is a simplified approach - in a real application, you might want to
	// store the original connection string and use it here
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GORM: %w", err)
	}

	return gormDB, nil
}
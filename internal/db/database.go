package db

import (
	"fmt"

	"github.com/habbas99/dexory/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase(host, port, user, password, dbName string) (*Database, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create database session, error: %w", err)
	}

	return &Database{DB: db}, nil
}

func (db *Database) Migrate() error {
	err := db.DB.AutoMigrate(
		&models.BulkScanRecord{},
		&models.Scan{},
		&models.ReportRecord{},
		&models.ComparisonData{},
		&models.ExportReportRecord{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migration in database, error: %w", err)
	}

	return nil
}

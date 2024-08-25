package db

import (
	"fmt"
	"github.com/habbas99/dexory/internal/models"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase(host, port, user, password, dbName string) (*Database, error) {
	err := createDatabase(host, port, user, password, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to create database, error: %w", err)
	}

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

func (db *Database) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB from GORM DB for closing, error: %w", err)
	}
	return sqlDB.Close()
}

func createDatabase(host, port, user, password, dbName string) error {
	// connect to the default database to create database if it does not exist
	defaultDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable", host, port, user, password)
	defaultDB, err := gorm.Open(postgres.Open(defaultDSN), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to default database, error: %w", err)
	}

	// check if the database exists
	var exists bool
	checkDbExistsQuery := fmt.Sprintf("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = '%s')", dbName)
	row := defaultDB.Raw(checkDbExistsQuery).Row()
	if err := row.Scan(&exists); err != nil {
		return fmt.Errorf("failed to check if database exists, error: %w", err)
	}

	// create the database only if it does not exist
	if !exists {
		createDbQuery := fmt.Sprintf("CREATE DATABASE %s", dbName)
		if err := defaultDB.Exec(createDbQuery).Error; err != nil {
			return fmt.Errorf("failed to create database, error: %w", err)
		}

		log.WithFields(log.Fields{
			"database_name": dbName,
		}).Info("database create successfully")
	} else {
		log.WithFields(log.Fields{
			"database_name": dbName,
		}).Info("database already exists")
	}

	return nil
}

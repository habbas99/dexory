package repositories

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/habbas99/dexory/internal/models"
	"gorm.io/gorm"
)

type BulkScanRecordRepository struct {
	DB *gorm.DB
}

func NewBulkScanRecordRepository(db *gorm.DB) *BulkScanRecordRepository {
	return &BulkScanRecordRepository{
		DB: db,
	}
}

func (bs *BulkScanRecordRepository) GetAll() ([]models.BulkScanRecord, error) {
	var bulkScanRecords []models.BulkScanRecord

	result := bs.DB.Find(&bulkScanRecords)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all bulk scan records, error: %w", result.Error)
	}

	if len(bulkScanRecords) == 0 {
		return []models.BulkScanRecord{}, nil
	}

	return bulkScanRecords, nil
}

func (bs *BulkScanRecordRepository) Create(filePath string) (*models.BulkScanRecord, error) {
	bulkScanRecord := models.BulkScanRecord{
		FileName: filepath.Base(filePath),
		FilePath: filePath,
		Status:   models.Pending,
	}

	result := bs.DB.Create(&bulkScanRecord)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create bulk scan record, error: %w", result.Error)
	}

	return &bulkScanRecord, nil
}

func (bs *BulkScanRecordRepository) Get(bulkScanRecordID uint) (*models.BulkScanRecord, error) {
	var bulkScanRecord models.BulkScanRecord
	result := bs.DB.First(&bulkScanRecord, bulkScanRecordID)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fnd bulk scan record by id=%d, error: %w", bulkScanRecordID, result.Error)
	}

	return &bulkScanRecord, nil
}

func (bs *BulkScanRecordRepository) GetByFileName(fileName string) (*models.BulkScanRecord, error) {
	var bulkScanRecord models.BulkScanRecord
	result := bs.DB.Where(&models.BulkScanRecord{FileName: fileName}).First(&bulkScanRecord)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("bulk scan record not found for filename=%s", fileName)
		}

		return nil, fmt.Errorf("failed to retrieve bulk scan record, error: %w", result.Error)
	}

	return &bulkScanRecord, nil
}

func (bs *BulkScanRecordRepository) Update(bulkScanRecord *models.BulkScanRecord) error {
	result := bs.DB.Save(bulkScanRecord)
	if result.Error != nil {
		return fmt.Errorf("failed to update bulk scan record with id=%d, error: %w", bulkScanRecord.ID, result.Error)
	}

	return nil
}

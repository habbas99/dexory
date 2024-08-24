package repositories

import (
	"fmt"
	"path/filepath"

	"github.com/habbas99/dexory/internal/models"
	"gorm.io/gorm"
)

type ReportRecordRepository struct {
	DB *gorm.DB
}

func NewReportRecordRepository(db *gorm.DB) *ReportRecordRepository {
	return &ReportRecordRepository{
		DB: db,
	}
}

func (rr *ReportRecordRepository) GetAll() ([]models.ReportRecord, error) {
	var reportRecords []models.ReportRecord

	result := rr.DB.Preload("BulkScanRecord").Find(&reportRecords)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all report records, error: %w", result.Error)
	}

	return reportRecords, nil
}

func (rr *ReportRecordRepository) Create(bulkScanRecord models.BulkScanRecord, referenceFilePath string) (*models.ReportRecord, error) {
	reportRecord := models.ReportRecord{
		BulkScanRecord:    bulkScanRecord,
		ReferenceFileName: filepath.Base(referenceFilePath),
		ReferenceFilePath: referenceFilePath,
		Status:            models.Pending,
	}

	result := rr.DB.Create(&reportRecord)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create report record, error: %w", result.Error)
	}

	return &reportRecord, nil
}

func (rr *ReportRecordRepository) Get(reportRecordID uint) (*models.ReportRecord, error) {
	var reportRecord models.ReportRecord
	result := rr.DB.Preload("BulkScanRecord").First(&reportRecord, reportRecordID)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fnd report record by id=%d, error: %w", reportRecordID, result.Error)
	}

	return &reportRecord, nil
}

func (rr *ReportRecordRepository) Update(reportRecord *models.ReportRecord) error {
	result := rr.DB.Save(reportRecord)
	if result.Error != nil {
		return fmt.Errorf("failed to update report record with id=%d, error: %w", reportRecord.ID, result.Error)
	}

	return nil
}

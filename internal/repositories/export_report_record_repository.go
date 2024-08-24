package repositories

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/habbas99/dexory/internal/models"
	"gorm.io/gorm"
)

type ExportReportRecordRepository struct {
	DB *gorm.DB
}

func NewExportReportRecordRepository(db *gorm.DB) *ExportReportRecordRepository {
	return &ExportReportRecordRepository{
		DB: db,
	}
}

func (er *ExportReportRecordRepository) GetAll(reportRecordID uint) ([]models.ExportReportRecord, error) {
	var exportReportRecords []models.ExportReportRecord

	result := er.DB.Where(&models.ExportReportRecord{ReportRecordID: reportRecordID}).Find(&exportReportRecords)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all export report records for report record id=%d, error: %w", reportRecordID, result.Error)
	}

	if len(exportReportRecords) == 0 {
		return []models.ExportReportRecord{}, nil
	}

	return exportReportRecords, nil
}

func (er *ExportReportRecordRepository) Get(exportReportRecordID uint) (*models.ExportReportRecord, error) {
	var exportReportRecord models.ExportReportRecord
	result := er.DB.First(&exportReportRecord, exportReportRecordID)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fnd export report record by id=%d, error: %w", exportReportRecordID, result.Error)
	}

	return &exportReportRecord, nil
}

func (er *ExportReportRecordRepository) GetByReportType(reportRecordID uint, reportType string) (*models.ExportReportRecord, error) {
	var exportReportRecord models.ExportReportRecord
	result := er.DB.Where(&models.ExportReportRecord{ReportRecordID: reportRecordID, ReportType: models.ExportReportType(reportType)}).First(&exportReportRecord)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to retrieve export report record, error: %w", result.Error)
	}

	return &exportReportRecord, nil
}

func (er *ExportReportRecordRepository) Create(reportRecordID uint, filePath, reportType string) (*models.ExportReportRecord, error) {
	exportReportRecord := models.ExportReportRecord{
		ReportType:     models.ExportReportType(reportType),
		FileName:       filepath.Base(filePath),
		FilePath:       filePath,
		Status:         models.Pending,
		ReportRecordID: reportRecordID,
	}

	result := er.DB.Create(&exportReportRecord)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create export report record, error: %w", result.Error)
	}

	return &exportReportRecord, nil
}

func (er *ExportReportRecordRepository) Update(exportReportRecord *models.ExportReportRecord) error {
	result := er.DB.Save(exportReportRecord)
	if result.Error != nil {
		return fmt.Errorf("failed to update export report record with id=%d, error: %w", exportReportRecord.ID, result.Error)
	}

	return nil
}

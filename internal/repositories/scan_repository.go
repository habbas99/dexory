package repositories

import (
	"errors"
	"fmt"
	"github.com/habbas99/dexory/internal/models"
	"gorm.io/gorm"
	"log"
)

type ScanRepository struct {
	DB *gorm.DB
}

type FileScanData struct {
	Name     string   `json:"name"`
	Scanned  bool     `json:"scanned"`
	Occupied bool     `json:"occupied"`
	Barcodes []string `json:"detected_barcodes"`
}

func NewScanRepository(db *gorm.DB) *ScanRepository {
	return &ScanRepository{
		DB: db,
	}
}

func (s *ScanRepository) CreateAll(scans []models.Scan) error {
	if scans == nil {
		return fmt.Errorf("scans cannot be nil")
	}

	result := s.DB.Create(scans)
	if result.Error != nil {
		return fmt.Errorf("failed to create scans, error: %w", result.Error)
	}

	return nil
}

func (s *ScanRepository) Get(bulkScanRecordID uint, location string) (*models.Scan, error) {
	var scan models.Scan
	result := s.DB.Where(&models.Scan{
		BulkScanRecordID: bulkScanRecordID,
		Location:         location,
	}).First(&scan)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Println("scan record not found")
			return nil, fmt.Errorf("scan not found for bulk scan record id=%d and location=%s", bulkScanRecordID, location)
		}

		return nil, fmt.Errorf("failed to retrieve scan, error: %w", result.Error)
	}

	return &scan, nil
}

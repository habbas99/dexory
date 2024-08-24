package repositories

import (
	"fmt"
	"github.com/habbas99/dexory/internal/models"
	"gorm.io/gorm"
)

type ComparisonDataRepository struct {
	DB *gorm.DB
}

func NewComparisonDataRepository(db *gorm.DB) *ComparisonDataRepository {
	return &ComparisonDataRepository{
		DB: db,
	}
}

func (rr *ComparisonDataRepository) GetAllPaginated(reportRecordID uint, limit int, offset int) ([]models.ComparisonData, error) {
	var comparisonDataList []models.ComparisonData

	result := rr.DB.Where(&models.ComparisonData{ReportRecordID: reportRecordID}).Limit(limit).Offset(offset).Find(&comparisonDataList)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get paginated comparison data, error: %w", result.Error)
	}

	return comparisonDataList, nil
}

func (cd *ComparisonDataRepository) Create(comparisonData *models.ComparisonData) error {
	if comparisonData == nil {
		return fmt.Errorf("comparison data cannot be nil")
	}

	result := cd.DB.Create(comparisonData)
	if result.Error != nil {
		return fmt.Errorf("failed to create comparison data, error: %w", result.Error)
	}

	return nil
}

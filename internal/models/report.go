package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type ScanComparisonOutcome string

const (
	LocationEmptyAsExpected                 ScanComparisonOutcome = "The location was empty, as expected"
	LocationEmptyButNotExpected             ScanComparisonOutcome = "The location was empty, but it should have been occupied"
	LocationOccupiedWithCorrectItems        ScanComparisonOutcome = "The location was occupied by the expected items"
	LocationOccupiedWithWrongItems          ScanComparisonOutcome = "The location was occupied by the wrong items"
	LocationOccupiedButExpectedEmpty        ScanComparisonOutcome = "The location was occupied by an item, but should have been empty"
	LocationOccupiedButBarcodeNotIdentified ScanComparisonOutcome = "The location was occupied, but no barcode could be identified"
)

type ExportReportType string

const (
	ExportReportJson ExportReportType = "json"
	ExportReportCsv  ExportReportType = "csv"
)

type ReportRecord struct {
	gorm.Model
	BulkScanRecordID  uint           `gorm:"index"`
	BulkScanRecord    BulkScanRecord `gorm:"foreignKey:BulkScanRecordID;references:ID"`
	ReferenceFileName string
	ReferenceFilePath string
	Status            Status
}

type ComparisonData struct {
	Location         string
	Scanned          bool
	Occupied         bool
	ActualBarcodes   pq.StringArray `gorm:"type:text[]"`
	ExpectedBarcodes pq.StringArray `gorm:"type:text[]"`
	Result           ScanComparisonOutcome
	ReportRecordID   uint
	ReportRecord     ReportRecord `gorm:"foreignKey:ReportRecordID;references:ID"`
}

type ExportReportRecord struct {
	gorm.Model
	ReportType     ExportReportType
	FileName       string
	FilePath       string
	Status         Status
	ReportRecordID uint
	ReportRecord   ReportRecord `gorm:"foreignKey:ReportRecordID;references:ID"`
}

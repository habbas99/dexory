package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type BulkScanRecord struct {
	gorm.Model
	FileName string
	FilePath string
	Status   Status
}

type Scan struct {
	gorm.Model
	Location         string
	Scanned          bool
	Occupied         bool
	Barcodes         pq.StringArray `gorm:"type:text[]"`
	BulkScanRecordID uint
	BulkScanRecord   BulkScanRecord `gorm:"foreignKey:BulkScanRecordID;references:ID"`
}

package scan

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/habbas99/dexory/internal/models"
)

type ScanService struct {
	bulkScanRecordClient bulkScanRecordClient
	scanClient           scanClient
	batchSize            int
}

type fileScanData struct {
	Name     string   `json:"name"`
	Scanned  bool     `json:"scanned"`
	Occupied bool     `json:"occupied"`
	Barcodes []string `json:"detected_barcodes"`
}

type scanClient interface {
	CreateAll(scans []models.Scan) error
}

type bulkScanRecordClient interface {
	Update(bulkScanRecord *models.BulkScanRecord) error
}

func NewScanService(bulkScanRecordClient bulkScanRecordClient, scanClient scanClient, batchSize int) *ScanService {
	return &ScanService{
		bulkScanRecordClient: bulkScanRecordClient,
		scanClient:           scanClient,
		batchSize:            batchSize,
	}
}

func (s *ScanService) ProcessFile(bulkScanRecord *models.BulkScanRecord) {
	log.WithFields(log.Fields{
		"bulk_scan_record_id": bulkScanRecord.ID,
		"file_name":           bulkScanRecord.FileName,
		"file_path":           bulkScanRecord.FilePath,
	}).Info("starting to process bulk scan file")

	s.updateBulkScanRecord(bulkScanRecord, models.Processing)

	filePath := bulkScanRecord.FilePath
	file, err := os.Open(filePath)
	if err != nil {
		s.updateBulkScanRecordWithStatusFailed(bulkScanRecord, fmt.Sprintf("failed to open file=%s", filePath), err)
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	// read the opening bracket of the array
	_, err = decoder.Token()
	if err != nil {
		s.updateBulkScanRecordWithStatusFailed(bulkScanRecord, fmt.Sprintf("failed to read starting array bracket in json file=%s", filePath), err)
		return
	}

	log.Printf("starting batch process for bulk scan record id=%d and scan file=%s", bulkScanRecord.ID, filePath)

	var batch []models.Scan

	// parse the JSON file in batches
	for decoder.More() {
		var fileScanData fileScanData

		// decode each object in the array
		if err := decoder.Decode(&fileScanData); err != nil {
			s.updateBulkScanRecordWithStatusFailed(bulkScanRecord, fmt.Sprintf("failed to decode json data from file=%s", filePath), err)
			return
		}

		scan := models.Scan{
			Location:         fileScanData.Name,
			Scanned:          fileScanData.Scanned,
			Occupied:         fileScanData.Occupied,
			Barcodes:         fileScanData.Barcodes,
			BulkScanRecordID: bulkScanRecord.ID,
		}

		batch = append(batch, scan)
		if len(batch) == s.batchSize {
			err := s.createScans(batch)
			if err != nil {
				s.updateBulkScanRecordWithStatusFailed(bulkScanRecord, "failed to create scans in database", err)
				return
			}
			batch = batch[:0] // reset batch
		}
	}

	// save any remaining scans in the last batch
	if len(batch) > 0 {
		err := s.createScans(batch)
		if err != nil {
			s.updateBulkScanRecordWithStatusFailed(bulkScanRecord, "failed to create remaining scans in database", err)
			return
		}
	}

	// read the closing bracket of the array
	_, err = decoder.Token()
	if err != nil {
		s.updateBulkScanRecordWithStatusFailed(bulkScanRecord, fmt.Sprintf("failed reading closing array bracket in json file=%s", filePath), err)
		return
	}

	s.updateBulkScanRecord(bulkScanRecord, models.Completed)

	log.WithFields(log.Fields{
		"bulk_scan_record_id": bulkScanRecord.ID,
		"file_name":           bulkScanRecord.FileName,
		"file_path":           bulkScanRecord.FilePath,
	}).Info("finished processing of bulk scan file")
}

func (s *ScanService) updateBulkScanRecordWithStatusFailed(bulkScanRecord *models.BulkScanRecord, message string, err error) {
	log.Printf("Error: %s: %v", message, err)
	s.updateBulkScanRecord(bulkScanRecord, models.Failed)
}

func (s *ScanService) updateBulkScanRecord(bulkScanRecord *models.BulkScanRecord, status models.Status) {
	bulkScanRecord.Status = status
	err := s.bulkScanRecordClient.Update(bulkScanRecord)
	if err != nil {
		log.Printf("Error: failed to update bulk scan record: %d to status: %s: %v", bulkScanRecord.ID, bulkScanRecord.Status, err)
	}
}

func (s *ScanService) createScans(batch []models.Scan) error {
	log.Printf("creating %d scans in database", len(batch))
	return s.scanClient.CreateAll(batch)
}

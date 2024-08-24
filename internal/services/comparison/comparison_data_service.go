package comparison

import (
	"encoding/csv"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"

	"github.com/habbas99/dexory/internal"
	"github.com/habbas99/dexory/internal/models"
)

type ComparisonDataService struct {
	scanClient           scanClient
	comparisonDataClient comparisonDataClient
	reportRecordClient   reportRecordClient
}

type Record struct {
	Location string
	Barcode  string
}

type scanClient interface {
	Get(bulkScanRecordID uint, location string) (*models.Scan, error)
}

type comparisonDataClient interface {
	Create(comparisonData *models.ComparisonData) error
}

type reportRecordClient interface {
	Update(reportRecord *models.ReportRecord) error
}

func NewComparisonDataService(scanClient scanClient, comparisonDataClient comparisonDataClient, reportRecordClient reportRecordClient) *ComparisonDataService {
	return &ComparisonDataService{
		scanClient:           scanClient,
		comparisonDataClient: comparisonDataClient,
		reportRecordClient:   reportRecordClient,
	}
}

func (rg *ComparisonDataService) GenerateComparisonDataForReport(reportRecord *models.ReportRecord) {
	log.WithFields(log.Fields{
		"report_record_id":    reportRecord,
		"reference_file_name": reportRecord.ReferenceFileName,
		"reference_file_path": reportRecord.ReferenceFilePath,
	}).Info("starting process to create comparison data for report record")

	rg.updateReportRecord(reportRecord, models.Processing)

	file, err := os.Open(reportRecord.ReferenceFilePath)
	if err != nil {
		rg.updateReportRecordWithStatusFailed(reportRecord, fmt.Sprintf("failed opening reference file=%s", reportRecord.ReferenceFilePath), err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		rg.updateReportRecordWithStatusFailed(reportRecord, fmt.Sprintf("failed reading csv headers from reference file=%s", reportRecord.ReferenceFilePath), err)
		return
	}

	if !strings.EqualFold(headers[0], "location") || !strings.EqualFold(headers[1], "item") {
		rg.updateReportRecordWithStatusFailed(reportRecord, fmt.Sprintf("reference file=%s contains wrong headers=%s", reportRecord.ReferenceFilePath, headers), nil)
		return
	}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			rg.updateReportRecordWithStatusFailed(reportRecord, fmt.Sprintf("failed reading row from reference file=%s", reportRecord.ReferenceFilePath), err)
			return
		}

		_, err = rg.createComparisonData(reportRecord.BulkScanRecord.ID, reportRecord.ID, row[0], row[1])
		if err != nil {
			rg.updateReportRecordWithStatusFailed(reportRecord, "failed to generate comparison data", err)
			return
		}
	}

	rg.updateReportRecord(reportRecord, models.Completed)

	log.WithFields(log.Fields{
		"report_record_id":    reportRecord,
		"reference_file_name": reportRecord.ReferenceFileName,
		"reference_file_path": reportRecord.ReferenceFilePath,
	}).Info("finished process to create comparison data for report record")
}

func (rg *ComparisonDataService) createComparisonData(bulkScanRecordID, reportRecordID uint, location, barcode string) (*models.ComparisonData, error) {
	scan, err := rg.scanClient.Get(bulkScanRecordID, location)
	if err != nil {
		return nil, fmt.Errorf("failed to get scan with bulk scan record id=%d and location=%s, error: %w", bulkScanRecordID, location, err)
	}
	if scan == nil {
		return nil, fmt.Errorf("scan not found with bulk scan record id=%d and location=%s", bulkScanRecordID, location)
	}

	outcome, err := rg.getComparisonResult(scan, barcode)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate comparison outcome for location=%s", location)
	}

	expectedBarcodes := []string{}
	if barcode != "" {
		expectedBarcodes = []string{barcode}
	}

	comparisonData := models.ComparisonData{
		Location:         location,
		Scanned:          scan.Scanned,
		Occupied:         scan.Occupied,
		ActualBarcodes:   scan.Barcodes,
		ExpectedBarcodes: expectedBarcodes,
		Result:           outcome,
		ReportRecordID:   reportRecordID,
	}

	err = rg.comparisonDataClient.Create(&comparisonData)
	if err != nil {
		return nil, fmt.Errorf("failed to create comparison data for location=%s, error: %w ", location, err)

	}

	return &comparisonData, nil
}

func (rg *ComparisonDataService) getComparisonResult(scan *models.Scan, barcode string) (models.ScanComparisonOutcome, error) {
	/*
	 compare scanned data when empty
	*/
	if !scan.Occupied && barcode == "" {
		return models.LocationEmptyAsExpected, nil
	}

	if !scan.Occupied && barcode != "" {
		return models.LocationEmptyButNotExpected, nil
	}

	/*
	 compare scanned data that shows location is occupied
	*/
	if scan.Occupied && len(scan.Barcodes) == 0 {
		return models.LocationOccupiedButBarcodeNotIdentified, nil
	}

	if scan.Occupied && barcode == "" {
		return models.LocationOccupiedButExpectedEmpty, nil
	}

	if len(scan.Barcodes) > 1 { // seems like a special case
		return models.LocationOccupiedWithWrongItems, nil
	}

	if len(scan.Barcodes) == 1 && scan.Barcodes[0] == barcode {
		return models.LocationOccupiedWithCorrectItems, nil
	}

	if len(scan.Barcodes) == 1 && scan.Barcodes[0] != barcode {
		return models.LocationOccupiedWithWrongItems, nil
	}

	return "", internal.ErrComparisonCaseNotSupported
}

func (rg *ComparisonDataService) updateReportRecordWithStatusFailed(reportRecord *models.ReportRecord, message string, err error) {
	log.Errorf("%s: %v", message, err)
	rg.updateReportRecord(reportRecord, models.Failed)
}

func (rg *ComparisonDataService) updateReportRecord(reportRecord *models.ReportRecord, status models.Status) {
	reportRecord.Status = status
	rg.reportRecordClient.Update(reportRecord)
}

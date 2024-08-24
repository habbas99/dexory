package export

import (
	"encoding/json"
	"fmt"
	"github.com/habbas99/dexory/internal/models"
	log "github.com/sirupsen/logrus"
	"os"
)

type jsonExportedComparisonData struct {
	Location         string   `json:"location"`
	Scanned          bool     `json:"scanned"`
	Occupied         bool     `json:"occupied"`
	ActualBarcodes   []string `json:"actualBarcodes"`
	ExpectedBarcodes []string `json:"expectedBarcodes"`
	Result           string   `json:"result"`
}

type exportReportRecordClient interface {
	Update(exportReportRecord *models.ExportReportRecord) error
}

type comparisonDataClient interface {
	GetAllPaginated(reportRecordID uint, limit int, offset int) ([]models.ComparisonData, error)
}

type ExportReportService struct {
	exportReportRecordClient exportReportRecordClient
	comparisonDataClient     comparisonDataClient
}

func NewExportReportService(exportReportRecordClient exportReportRecordClient, comparisonDataClient comparisonDataClient) *ExportReportService {
	return &ExportReportService{
		exportReportRecordClient: exportReportRecordClient,
		comparisonDataClient:     comparisonDataClient,
	}
}

func (er *ExportReportService) ExportReport(exportReportRecord *models.ExportReportRecord) {
	log.WithFields(log.Fields{
		"export_report_record_id": exportReportRecord.ReportRecordID,
		"file_name":               exportReportRecord.FileName,
		"file_path":               exportReportRecord.FilePath,
	}).Info("starting process to export report record")

	er.updateExportReportRecord(exportReportRecord, models.Processing)

	file, err := os.OpenFile(exportReportRecord.FilePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		er.updateExportReportRecordWithStatusFailed(exportReportRecord, fmt.Sprintf("failed opening export report file=%s", exportReportRecord.FilePath), err)
		return
	}
	defer file.Close()

	err = er.writeArrayStartingBracket(file)
	if err != nil {
		er.updateExportReportRecordWithStatusFailed(exportReportRecord, fmt.Sprintf("failed to write starting array bracket to export report file=%s", file.Name()), err)
		return
	}

	limit := 50
	offset := 0
	firstObject := true
	for {
		comparisonDataList, err := er.comparisonDataClient.GetAllPaginated(exportReportRecord.ID, limit, offset)
		if err != nil {
			er.updateExportReportRecordWithStatusFailed(exportReportRecord, fmt.Sprintf("failed to get comparison data for report record id=%d", exportReportRecord.ReportRecordID), err)
			return
		}

		if len(comparisonDataList) == 0 {
			break
		}

		for _, comparisonData := range comparisonDataList {
			// write a comma before each object, except the first one
			if !firstObject {
				err = er.writeStringToReportFile(",\n", file)
				if err != nil {
					er.updateExportReportRecordWithStatusFailed(exportReportRecord, fmt.Sprintf("failed to write comma to export report file=%s", file.Name()), err)
					return
				}
			} else {
				firstObject = false
			}

			data := jsonExportedComparisonData{
				Location:         comparisonData.Location,
				Scanned:          comparisonData.Scanned,
				Occupied:         comparisonData.Occupied,
				ActualBarcodes:   comparisonData.ActualBarcodes,
				ExpectedBarcodes: comparisonData.ExpectedBarcodes,
				Result:           string(comparisonData.Result),
			}

			jsonData, err := json.MarshalIndent(data, "  ", "  ")
			if err != nil {
				er.updateExportReportRecordWithStatusFailed(exportReportRecord, fmt.Sprintf("failed to write comparison data to export report file=%s", file.Name()), err)
				return
			}

			err = er.writeBytesToReportFile(jsonData, file)
			if err != nil {
				er.updateExportReportRecordWithStatusFailed(exportReportRecord, fmt.Sprintf("failed to write comparison data json to export report file=%s", file.Name()), err)
				return
			}
		}

		// ensure data is flushed to disk
		err = file.Sync()
		if err != nil {
			er.updateExportReportRecordWithStatusFailed(exportReportRecord, fmt.Sprintf("failed to sync data to disk for export report file=%s", file.Name()), err)
			return
		}

		offset += len(comparisonDataList) // move to the next batch
	}

	err = er.writeArrayClosingBracket(file)
	if err != nil {
		er.updateExportReportRecordWithStatusFailed(exportReportRecord, fmt.Sprintf("failed to write ending array bracket to export report file=%s", file.Name()), err)
		return
	}

	er.updateExportReportRecord(exportReportRecord, models.Completed)

	log.WithFields(log.Fields{
		"export_report_record_id": exportReportRecord.ReportRecordID,
		"file_name":               exportReportRecord.FileName,
		"file_path":               exportReportRecord.FilePath,
	}).Info("finished process to export report record")
}

func (er *ExportReportService) writeArrayStartingBracket(file *os.File) error {
	return er.writeStringToReportFile("[\n", file)
}

func (er *ExportReportService) writeArrayClosingBracket(file *os.File) error {
	return er.writeStringToReportFile("\n]", file)
}

func (er *ExportReportService) writeStringToReportFile(str string, file *os.File) error {
	return er.writeBytesToReportFile([]byte(str), file)
}

func (er *ExportReportService) writeBytesToReportFile(bytes []byte, file *os.File) error {
	_, err := file.Write(bytes)
	return err
}

func (er *ExportReportService) updateExportReportRecordWithStatusFailed(exportReportRecord *models.ExportReportRecord, message string, err error) {
	log.Errorf("%s: %v", message, err)
	er.updateExportReportRecord(exportReportRecord, models.Failed)
}

func (er *ExportReportService) updateExportReportRecord(exportReportRecord *models.ExportReportRecord, status models.Status) {
	exportReportRecord.Status = status
	err := er.exportReportRecordClient.Update(exportReportRecord)
	if err != nil {
		log.WithFields(log.Fields{
			"export_report_record_id": exportReportRecord.ID,
			"status":                  exportReportRecord.Status,
		}).Errorf("failed to update export report record status, error: %v", err)
	}
}

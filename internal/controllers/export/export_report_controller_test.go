package export

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mockexportreportcontroller "github.com/habbas99/dexory/generated/controllers/export"
	"github.com/habbas99/dexory/internal/models"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type ExportReportControllerTestSuite struct {
	suite.Suite
	mockFileStorageClient         *mockexportreportcontroller.MockfileStorageClient
	mockExportReportRecordClient  *mockexportreportcontroller.MockexportReportRecordClient
	mockExportReportServiceClient *mockexportreportcontroller.MockexportReportServiceClient
	exportReportController        *ExportReportController
	ctrl                          *gomock.Controller
}

func TestExportReportControllerTestSuite(t *testing.T) {
	suite.Run(t, new(ExportReportControllerTestSuite))
}

func (suite *ExportReportControllerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	suite.ctrl = gomock.NewController(suite.T())
	suite.mockFileStorageClient = mockexportreportcontroller.NewMockfileStorageClient(suite.ctrl)
	suite.mockExportReportRecordClient = mockexportreportcontroller.NewMockexportReportRecordClient(suite.ctrl)
	suite.mockExportReportServiceClient = mockexportreportcontroller.NewMockexportReportServiceClient(suite.ctrl)

	tempDir, err := os.MkdirTemp("", "exports")
	if err != nil {
		fmt.Println("Error creating temp directory:", err)
		return
	}

	suite.exportReportController = NewExportReportController(
		tempDir, suite.mockFileStorageClient, suite.mockExportReportRecordClient, suite.mockExportReportServiceClient,
	)
}

func (suite *ExportReportControllerTestSuite) TearDownTest() {
	os.RemoveAll(suite.exportReportController.dirPath)
	suite.ctrl.Finish()
}

func (suite *ExportReportControllerTestSuite) TestGetExportReportRecords() {
	// Given
	reportRecordID := uint(1)
	exportReportRecord := models.ExportReportRecord{
		ReportRecordID: uint(1),
		ReportType:     models.ExportReportJson,
		FileName:       "report.json",
		Status:         models.Completed,
	}
	exportReportRecord.ID = uint(1)
	exportReportRecords := []models.ExportReportRecord{exportReportRecord}

	suite.mockExportReportRecordClient.EXPECT().GetAll(reportRecordID).Return(exportReportRecords, nil).Times(1)

	router := gin.Default()
	router.GET("/inventory-comparison-reports/:id/exports", suite.exportReportController.GetExportReportRecords)

	// When
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/inventory-comparison-reports/1/exports", nil)
	router.ServeHTTP(recorder, request)

	// Then
	suite.Equal(http.StatusOK, recorder.Code)
	suite.JSONEq(`[{
		"id":1,
		"fileName":"report.json",
		"status":"completed"
	}]`, recorder.Body.String())
}

func (suite *ExportReportControllerTestSuite) TestCreateExportReportRecord() {
	// Given
	reportRecordID := uint(1)
	reportType := string(models.ExportReportJson)
	requestBody := fmt.Sprintf(`{"reportRecordId": %d, "reportType": "%s"}`, reportRecordID, reportType)

	suite.mockExportReportRecordClient.EXPECT().GetByReportType(reportRecordID, reportType).Return(nil, nil).Times(1)

	testFileName := "report_1.json"
	tempFile, err := os.CreateTemp("", testFileName)
	suite.Require().NoError(err)
	defer os.Remove(tempFile.Name())
	suite.mockFileStorageClient.EXPECT().CreateFile(gomock.Any(), gomock.Any()).Return(tempFile, nil).Times(1)

	exportReportRecord := &models.ExportReportRecord{
		ReportRecordID: reportRecordID,
		ReportType:     models.ExportReportJson,
		FilePath:       tempFile.Name(),
		Status:         models.Pending,
	}
	exportReportRecord.ID = uint(1)
	suite.mockExportReportRecordClient.EXPECT().Create(reportRecordID, tempFile.Name(), reportType).Return(exportReportRecord, nil).Times(1)

	var wg sync.WaitGroup
	wg.Add(1)
	suite.mockExportReportServiceClient.EXPECT().ExportReport(exportReportRecord).Do(func(_ *models.ExportReportRecord) {
		wg.Done() // mark as done when the method is called
	}).Times(1)

	router := gin.Default()
	router.POST("/export-report-records", suite.exportReportController.CreateExportReportRecord)

	// When
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/export-report-records", strings.NewReader(requestBody))
	router.ServeHTTP(recorder, request)

	wg.Wait() // wait for the asynchronous call to finish

	// Then
	suite.Equal(http.StatusOK, recorder.Code)
	suite.JSONEq(`{"id": 1}`, recorder.Body.String())
}

func (suite *ExportReportControllerTestSuite) TestCreateExportReportWithUnsupportedReportType() {
	// Given
	reportRecordID := uint(1)
	reportType := "csv"
	requestBody := fmt.Sprintf(`{"reportRecordId": %d, "reportType": "%s"}`, reportRecordID, reportType)

	router := gin.Default()
	router.POST("/export-report-records", suite.exportReportController.CreateExportReportRecord)

	// When
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/export-report-records", strings.NewReader(requestBody))
	router.ServeHTTP(recorder, request)

	// Then
	suite.Equal(http.StatusInternalServerError, recorder.Code)
	suite.JSONEq(`{"error": "report type not supported"}`, recorder.Body.String())
}

func (suite *ExportReportControllerTestSuite) TestDownloadReport() {
	// Given
	exportReportRecordID := uint(1)

	tempFile, err := os.CreateTemp("", "report_*.json")
	suite.Require().NoError(err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(`[{"key":"value"}]`)
	suite.Require().NoError(err)
	tempFile.Close()

	exportReportRecord := &models.ExportReportRecord{
		ReportRecordID: exportReportRecordID,
		ReportType:     models.ExportReportJson,
		FilePath:       tempFile.Name(),
		Status:         models.Completed,
	}
	exportReportRecord.ID = exportReportRecordID
	suite.mockExportReportRecordClient.EXPECT().Get(exportReportRecordID).Return(exportReportRecord, nil).Times(1)

	router := gin.Default()
	router.GET("/export-report-records/:id/download", suite.exportReportController.DownloadReport)

	// When
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", fmt.Sprintf("/export-report-records/%d/download", exportReportRecordID), nil)
	router.ServeHTTP(recorder, request)

	// Then
	suite.Equal(http.StatusOK, recorder.Code)
	suite.Equal(fmt.Sprintf("attachment; filename=%s", filepath.Base(tempFile.Name())), recorder.Header().Get("Content-Disposition"))
	suite.Equal(`[{"key":"value"}]`, recorder.Body.String())
}

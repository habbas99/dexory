package report

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/habbas99/dexory/generated/controllers/report"
	"github.com/habbas99/dexory/internal/models"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

type ReportRecordControllerTestSuite struct {
	suite.Suite
	dirPath                         string
	mockFileStorageClient           *mockreportrecordcontroller.MockfileStorageClient
	mockBulkScanRecordClient        *mockreportrecordcontroller.MockbulkScanRecordClient
	mockReportRecordClient          *mockreportrecordcontroller.MockreportRecordClient
	mockComparisonDataClient        *mockreportrecordcontroller.MockcomparisonDataClient
	mockComparisonDataServiceClient *mockreportrecordcontroller.MockcomparisonDataServiceClient
	reportRecordController          *ReportRecordController
	ctrl                            *gomock.Controller
}

func TestReportRecordControllerTestSuite(t *testing.T) {
	suite.Run(t, new(ReportRecordControllerTestSuite))
}

func (suite *ReportRecordControllerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	suite.ctrl = gomock.NewController(suite.T())
	suite.mockFileStorageClient = mockreportrecordcontroller.NewMockfileStorageClient(suite.ctrl)
	suite.mockBulkScanRecordClient = mockreportrecordcontroller.NewMockbulkScanRecordClient(suite.ctrl)
	suite.mockReportRecordClient = mockreportrecordcontroller.NewMockreportRecordClient(suite.ctrl)
	suite.mockComparisonDataClient = mockreportrecordcontroller.NewMockcomparisonDataClient(suite.ctrl)
	suite.mockComparisonDataServiceClient = mockreportrecordcontroller.NewMockcomparisonDataServiceClient(suite.ctrl)

	tempDir, err := os.MkdirTemp("", "comparison-reports")
	if err != nil {
		fmt.Println("Error creating temp directory:", err)
		return
	}

	suite.dirPath = tempDir

	suite.reportRecordController = NewReportRecordController(
		tempDir,
		suite.mockFileStorageClient,
		suite.mockBulkScanRecordClient,
		suite.mockReportRecordClient,
		suite.mockComparisonDataClient,
		suite.mockComparisonDataServiceClient,
	)
}

func (suite *ReportRecordControllerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *ReportRecordControllerTestSuite) TestGetReportRecords() {
	// Given
	bulkScanRecord := models.BulkScanRecord{
		FileName: "scans_001.json",
		Status:   models.Completed,
	}
	bulkScanRecord.ID = uint(1)
	reportRecord := models.ReportRecord{
		ReferenceFileName: "scans.csv",
		Status:            "completed",
		BulkScanRecord:    bulkScanRecord,
	}
	reportRecord.ID = uint(1)
	reportRecord.CreatedAt = time.Date(2024, 8, 22, 13, 0, 0, 0, time.UTC)
	reportRecord.UpdatedAt = time.Date(2024, 8, 22, 13, 5, 0, 0, time.UTC)

	ReportRecordList := []models.ReportRecord{reportRecord}

	suite.mockReportRecordClient.EXPECT().GetAll().Return(ReportRecordList, nil)

	router := gin.Default()
	router.GET("/inventory-comparison-reports", suite.reportRecordController.GetAllReportRecords)

	// When
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/inventory-comparison-reports", nil)
	router.ServeHTTP(recorder, request)

	log.Println(recorder.Body.String())

	// Then
	suite.Equal(http.StatusOK, recorder.Code)
	suite.JSONEq(`[{
		"id":1,
		"bulkScanFileName":"scans_001.json",
		"referenceFileName":"scans.csv",
		"status":"completed",
		"createdAt":"2024-08-22T13:00:00Z",
		"updatedAt":"2024-08-22T13:05:00Z"
	}]`, recorder.Body.String())
}

func (suite *ReportRecordControllerTestSuite) TestCreateReportRecord() {
	// Given
	bulkScanFileName := "scans_001.json"
	uploadedFileName := "scans.csv"
	fileContent := "content does not matter"

	bulkScanRecord := &models.BulkScanRecord{
		FileName: bulkScanFileName,
		Status:   models.Completed,
	}
	bulkScanRecord.ID = uint(1)

	reportRecord := &models.ReportRecord{
		ReferenceFileName: uploadedFileName,
		Status:            "processing",
	}
	reportRecord.ID = uint(1)

	suite.mockBulkScanRecordClient.EXPECT().GetByFileName(bulkScanFileName).Return(bulkScanRecord, nil).Times(1)
	suite.mockReportRecordClient.EXPECT().Create(*bulkScanRecord, gomock.Any()).Return(reportRecord, nil).Times(1)

	tempFile, err := os.CreateTemp("", uploadedFileName)
	suite.Require().NoError(err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte(fileContent))
	suite.Require().NoError(err)
	suite.Require().NoError(tempFile.Close())

	suite.mockFileStorageClient.EXPECT().SaveFile(suite.dirPath, uploadedFileName, gomock.Any()).Return(tempFile, nil).Times(1)

	var wg sync.WaitGroup
	wg.Add(1)
	suite.mockComparisonDataServiceClient.EXPECT().GenerateComparisonDataForReport(reportRecord).Do(func(_ *models.ReportRecord) {
		wg.Done() // mark as done when the method is called
	}).Times(1)

	router := gin.Default()
	router.POST("/inventory-comparison-reports", suite.reportRecordController.CreateReportRecord)

	// create a multipart form to simulate file upload
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("bulkScanFileName", bulkScanFileName)
	part, _ := writer.CreateFormFile("csvFile", uploadedFileName)
	part.Write([]byte(fileContent))
	writer.Close()

	// When
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/inventory-comparison-reports", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(recorder, request)

	wg.Wait() // wait for the asynchronous call to finish

	// Then
	suite.Equal(http.StatusOK, recorder.Code)
	suite.JSONEq(`{"id": 1}`, recorder.Body.String())
}

func (suite *ReportRecordControllerTestSuite) TestGetReport() {
	// Given
	reportRecordID := uint(1)
	bulkScanRecord := models.BulkScanRecord{
		FileName: "scans_001.json",
		Status:   models.Completed,
	}
	bulkScanRecord.ID = uint(1)

	reportRecord := models.ReportRecord{
		ReferenceFileName: "scans.csv",
		Status:            models.Completed,
		BulkScanRecord:    bulkScanRecord,
	}
	reportRecord.ID = uint(1)
	reportRecord.CreatedAt = time.Date(2024, 8, 22, 13, 0, 0, 0, time.UTC)
	reportRecord.UpdatedAt = time.Date(2024, 8, 22, 13, 5, 0, 0, time.UTC)

	suite.mockReportRecordClient.EXPECT().Get(reportRecordID).Return(&reportRecord, nil).Times(1)

	router := gin.Default()
	router.GET("/inventory-comparison-reports/:id", suite.reportRecordController.GetReport)

	// When
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", fmt.Sprintf("/inventory-comparison-reports/%d", reportRecordID), nil)
	router.ServeHTTP(recorder, request)

	log.Println(recorder.Body.String())

	// Then
	suite.Equal(http.StatusOK, recorder.Code)
	suite.JSONEq(`{
		"id": 1,
		"bulkScanFileName": "scans_001.json",
		"referenceFileName": "scans.csv",
		"status": "completed",
		"createdAt": "2024-08-22T13:00:00Z",
		"updatedAt": "2024-08-22T13:05:00Z"
	}`, recorder.Body.String())
}

func (suite *ReportRecordControllerTestSuite) TestGetComparisonData() {
	// Given
	reportID := uint(1)
	comparisonData := models.ComparisonData{
		ReportRecordID:   uint(1),
		Location:         "Location1",
		Scanned:          true,
		Occupied:         true,
		ActualBarcodes:   []string{"Barcode1"},
		ExpectedBarcodes: []string{"Barcode1"},
		Result:           models.LocationOccupiedWithCorrectItems,
	}
	comparisonDataList := []models.ComparisonData{comparisonData}

	suite.mockComparisonDataClient.EXPECT().GetAllPaginated(reportID, gomock.Any(), gomock.Any()).Return(comparisonDataList, nil).Times(1)

	router := gin.Default()
	router.GET("/inventory-comparison-reports/:id/data", suite.reportRecordController.GetComparisonData)

	// When
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/inventory-comparison-reports/1/data", nil)
	router.ServeHTTP(recorder, request)

	// Then
	suite.Equal(http.StatusOK, recorder.Code)
	suite.JSONEq(`[{
		"location":"Location1",
		"scanned":true,
		"occupied":true,
		"actualBarcodes":["Barcode1"],
		"expectedBarcodes":["Barcode1"],
		"result":"The location was occupied by the expected items"
	}]`, recorder.Body.String())
}

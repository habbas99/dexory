package scan

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mockscancontroller "github.com/habbas99/dexory/generated/controllers/scan"
	"github.com/habbas99/dexory/internal/models"
	"github.com/stretchr/testify/suite"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

type ScanControllerTestSuite struct {
	suite.Suite
	mockFileStorageClient    *mockscancontroller.MockFileStorageClient
	mockBulkScanRecordClient *mockscancontroller.MockBulkScanRecordClient
	mockScanServiceClient    *mockscancontroller.MockScanServiceClient
	scanController           *ScanController
	ctrl                     *gomock.Controller
}

func TestScanControllerTestSuite(t *testing.T) {
	suite.Run(t, new(ScanControllerTestSuite))
}

func (suite *ScanControllerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	suite.ctrl = gomock.NewController(suite.T())

	suite.mockFileStorageClient = mockscancontroller.NewMockFileStorageClient(suite.ctrl)
	suite.mockBulkScanRecordClient = mockscancontroller.NewMockBulkScanRecordClient(suite.ctrl)
	suite.mockScanServiceClient = mockscancontroller.NewMockScanServiceClient(suite.ctrl)

	tempDir, err := os.MkdirTemp("", "scans")
	if err != nil {
		fmt.Println("Error creating temp directory:", err)
		return
	}

	suite.scanController = NewScanController(
		tempDir, suite.mockFileStorageClient, suite.mockBulkScanRecordClient, suite.mockScanServiceClient,
	)
}

func (suite *ScanControllerTestSuite) TearDownTest() {
	os.RemoveAll(suite.scanController.dirPath)
	suite.ctrl.Finish()
}

func (suite *ScanControllerTestSuite) TestGetBulkScanRecords() {
	// Given
	bulkScanRecord := models.BulkScanRecord{
		FileName: "scans_001.json",
		Status:   models.Completed,
	}
	bulkScanRecord.ID = uint(1)

	suite.mockBulkScanRecordClient.EXPECT().GetAll().Return([]models.BulkScanRecord{bulkScanRecord}, nil).Times(1)

	router := gin.Default()
	router.GET("/bulk-scan-records", suite.scanController.GetBulkScanRecords)

	// When
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/bulk-scan-records", nil)
	router.ServeHTTP(recorder, request)

	// Then
	suite.Equal(http.StatusOK, recorder.Code)
	suite.JSONEq(`[{
		"id":1,
		"fileName":"scans_001.json",
		"status":"completed"
	}]`, recorder.Body.String())
}

func (suite *ScanControllerTestSuite) TestUploadBulkScanFile() {
	// Given
	testFileName := "scans_002.json"
	testFileContent := `{
		"name": "test_location",
		"scanned": true,
		"occupied": false, 
		"detected_barcodes": ["barcode1"]
	}`

	tempFile, err := os.CreateTemp("", testFileName)
	suite.Require().NoError(err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte(testFileContent))
	suite.Require().NoError(err)
	suite.Require().NoError(tempFile.Close())

	suite.mockFileStorageClient.EXPECT().SaveFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(tempFile, nil).Times(1)

	bulkScanRecord := models.BulkScanRecord{FilePath: tempFile.Name(), Status: models.Pending}
	bulkScanRecord.ID = uint(1)
	suite.mockBulkScanRecordClient.EXPECT().Create(tempFile.Name()).Return(&bulkScanRecord, nil).Times(1)

	var wg sync.WaitGroup
	wg.Add(1)
	suite.mockScanServiceClient.EXPECT().ProcessFile(&bulkScanRecord).Do(func(_ *models.BulkScanRecord) {
		wg.Done() // mark as done when the method is called
	}).Times(1)

	router := gin.Default()
	router.POST("/upload-bulk-scan-file", suite.scanController.UploadBulkScanFile)

	// When
	recorder := httptest.NewRecorder()
	file, _ := os.Open(tempFile.Name())
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	suite.Require().NoError(err)
	_, err = io.Copy(part, file)
	suite.Require().NoError(err)
	writer.Close()

	request, err := http.NewRequest("POST", "/upload-bulk-scan-file", body)
	suite.Require().NoError(err)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	router.ServeHTTP(recorder, request)

	wg.Wait() // wait for the asynchronous call to finish

	// Then
	suite.Equal(http.StatusOK, recorder.Code)
	suite.JSONEq(`{"id": 1}`, recorder.Body.String())
}

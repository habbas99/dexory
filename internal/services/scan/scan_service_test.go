package scan

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/habbas99/dexory/generated/services/scan"
	"github.com/habbas99/dexory/internal/models"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type ScanServiceTestSuite struct {
	suite.Suite
	MockScanClient           *mockscanservice.MockscanClient
	MockBulkScanRecordClient *mockscanservice.MockbulkScanRecordClient
	ScanService              *ScanService
	ctrl                     *gomock.Controller
}

func TestScanServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ScanServiceTestSuite))
}

func (suite *ScanServiceTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.MockScanClient = mockscanservice.NewMockscanClient(suite.ctrl)
	suite.MockBulkScanRecordClient = mockscanservice.NewMockbulkScanRecordClient(suite.ctrl)

	suite.ScanService = NewScanService(suite.MockBulkScanRecordClient, suite.MockScanClient, 10)
}

func (suite *ScanServiceTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *ScanServiceTestSuite) TestProcessFileInBatchesSuccess() {
	// Given
	service := NewScanService(suite.MockBulkScanRecordClient, suite.MockScanClient, 1)

	mockFileContent := `[
		{"name": "Location1", "scanned": true, "occupied": true, "detected_barcodes": ["Barcode1", "Barcode2"]},
		{"name": "Location2", "scanned": true, "occupied": true, "detected_barcodes": ["Barcode3"]},
		{"name": "Location3", "scanned": true, "occupied": false, "detected_barcodes": []}
	]`
	mockFile := suite.createMockJSONFile(mockFileContent)
	defer os.Remove(mockFile.Name())

	bulkScanRecord := &models.BulkScanRecord{
		FilePath: mockFile.Name(),
		Status:   models.Pending,
	}
	bulkScanRecord.ID = uint(1)

	suite.MockBulkScanRecordClient.EXPECT().Update(bulkScanRecord).Return(nil).Times(2)
	suite.MockScanClient.EXPECT().CreateAll(gomock.Any()).Return(nil).Times(3)

	// When
	service.ProcessFile(bulkScanRecord)

	// Then
	suite.Equal("completed", string(bulkScanRecord.Status))
}

func (suite *ScanServiceTestSuite) TestProcessFileInOneBatchSuccess() {
	// Given
	mockFileContent := `[
		{"name": "Location1", "scanned": true, "occupied": true, "detected_barcodes": ["Barcode1", "Barcode2"]},
		{"name": "Location2", "scanned": true, "occupied": true, "detected_barcodes": ["Barcode3"]}
	]`
	mockFile := suite.createMockJSONFile(mockFileContent)
	defer os.Remove(mockFile.Name())

	bulkScanRecord := &models.BulkScanRecord{
		FilePath: mockFile.Name(),
		Status:   models.Pending,
	}
	bulkScanRecord.ID = uint(1)

	suite.MockBulkScanRecordClient.EXPECT().Update(bulkScanRecord).Return(nil).Times(2)
	suite.MockScanClient.EXPECT().CreateAll(gomock.Any()).Return(nil).Times(1)

	// When
	suite.ScanService.ProcessFile(bulkScanRecord)

	// Then
	suite.Equal("completed", string(bulkScanRecord.Status))
}

func (suite *ScanServiceTestSuite) TestProcessFileFailedToOpenJsonFile() {
	// Given
	bulkScanRecord := &models.BulkScanRecord{
		FilePath: "nonexistent.json",
		Status:   models.Pending,
	}
	bulkScanRecord.ID = uint(1)

	suite.MockBulkScanRecordClient.EXPECT().Update(bulkScanRecord).Return(nil).Times(2)

	// When
	suite.ScanService.ProcessFile(bulkScanRecord)

	// Then
	suite.Equal(models.Failed, bulkScanRecord.Status)
}

func (suite *ScanServiceTestSuite) TestProcessFileJsonDecodeFailure() {
	// Given
	bulkScanRecord := &models.BulkScanRecord{
		FilePath: "invalid.json",
		Status:   models.Pending,
	}
	bulkScanRecord.ID = uint(1)

	mockFileContent := "invalid json content"
	mockFile := suite.createMockJSONFile(mockFileContent)
	defer os.Remove(mockFile.Name())

	suite.MockBulkScanRecordClient.EXPECT().Update(bulkScanRecord).Return(nil).Times(2)

	// When
	suite.ScanService.ProcessFile(bulkScanRecord)

	// Then
	suite.Equal("failed", string(bulkScanRecord.Status))
}

func (suite *ScanServiceTestSuite) TestProcessFileCreateScansFailed() {
	// Given
	mockFileContent := `[
		{"name": "Location1", "scanned": true, "occupied": true, "detected_barcodes": ["Barcode1"]}
	]`
	mockFile := suite.createMockJSONFile(mockFileContent)
	defer os.Remove(mockFile.Name())

	bulkScanRecord := &models.BulkScanRecord{
		FilePath: mockFile.Name(),
		Status:   models.Pending,
	}
	bulkScanRecord.ID = uint(1)

	suite.MockBulkScanRecordClient.EXPECT().Update(bulkScanRecord).Return(nil).Times(2)
	suite.MockScanClient.EXPECT().CreateAll(gomock.Any()).Return(fmt.Errorf("database error")).Times(1)

	// When
	suite.ScanService.ProcessFile(bulkScanRecord)

	// Then
	suite.Equal("failed", string(bulkScanRecord.Status))
}

func (suite *ScanServiceTestSuite) createMockJSONFile(content string) *os.File {
	file, err := os.CreateTemp("", "test*.json")
	suite.Require().NoError(err)

	_, err = file.WriteString(content)
	suite.Require().NoError(err)

	err = file.Close()
	suite.Require().NoError(err)

	return file
}

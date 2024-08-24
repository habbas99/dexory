package comparison

import (
	"fmt"
	mockcomparisondataservice "github.com/habbas99/dexory/generated/services/report"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/habbas99/dexory/internal/models"
	"github.com/stretchr/testify/suite"
)

type ComparisonDataServiceTestSuite struct {
	suite.Suite
	MockScanClient           *mockcomparisondataservice.MockScanClient
	MockComparisonDataClient *mockcomparisondataservice.MockComparisonDataClient
	MockReportRecordClient   *mockcomparisondataservice.MockReportRecordClient
	ComparisonDataService    *ComparisonDataService
	ctrl                     *gomock.Controller
}

func TestComparisonDataServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ComparisonDataServiceTestSuite))
}

func (suite *ComparisonDataServiceTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.MockScanClient = mockcomparisondataservice.NewMockScanClient(suite.ctrl)
	suite.MockComparisonDataClient = mockcomparisondataservice.NewMockComparisonDataClient(suite.ctrl)
	suite.MockReportRecordClient = mockcomparisondataservice.NewMockReportRecordClient(suite.ctrl)

	suite.ComparisonDataService = NewComparisonDataService(suite.MockScanClient, suite.MockComparisonDataClient, suite.MockReportRecordClient)
}

func (suite *ComparisonDataServiceTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *ComparisonDataServiceTestSuite) TestGenerateComparisonDataForReport() {
	// Given
	mockFile := suite.createMockCSVFile([]string{
		"Location,Item",
		"Location1,Barcode1",
		"Location2,Barcode2",
	})
	defer os.Remove(mockFile.Name())

	bulkScanRecord := models.BulkScanRecord{}
	bulkScanRecord.ID = uint(1)

	reportRecord := &models.ReportRecord{
		BulkScanRecord:    bulkScanRecord,
		ReferenceFilePath: mockFile.Name(),
	}

	suite.MockReportRecordClient.EXPECT().Update(reportRecord).Return(nil).Times(2)

	suite.MockScanClient.EXPECT().Get(uint(1), "Location1").Return(&models.Scan{
		Scanned:  true,
		Occupied: true,
		Barcodes: []string{"Barcode1"},
	}, nil)

	suite.MockScanClient.EXPECT().Get(uint(1), "Location2").Return(&models.Scan{
		Scanned:  true,
		Occupied: true,
		Barcodes: []string{"Barcode2"},
	}, nil)

	suite.MockComparisonDataClient.EXPECT().Create(gomock.Any()).Return(nil).Times(2)

	// When
	suite.ComparisonDataService.GenerateComparisonDataForReport(reportRecord)

	// Then
	suite.NotNil(reportRecord)
	suite.Equal("completed", string(reportRecord.Status))
}

func (suite *ComparisonDataServiceTestSuite) TestGenerateComparisonDataForReportFileDoesNotExist() {
	// Given
	bulkScanRecord := models.BulkScanRecord{}
	bulkScanRecord.ID = uint(1)

	reportRecord := &models.ReportRecord{
		BulkScanRecord:    bulkScanRecord,
		ReferenceFilePath: "nonexistent.csv",
	}

	suite.MockReportRecordClient.EXPECT().Update(reportRecord).Return(nil).Times(2)

	// When
	suite.ComparisonDataService.GenerateComparisonDataForReport(reportRecord)

	// Then
	suite.NotNil(reportRecord)
	suite.Equal("failed", string(reportRecord.Status))
}

func (suite *ComparisonDataServiceTestSuite) TestGenerateComparisonDataForReportFileHasInvalidHeaders() {
	// Given
	mockFile := suite.createMockCSVFile([]string{
		"Location,code",
		"Location1,Barcode1",
	})
	defer os.Remove(mockFile.Name())

	bulkScanRecord := models.BulkScanRecord{}
	bulkScanRecord.ID = uint(1)

	reportRecord := &models.ReportRecord{
		BulkScanRecord:    bulkScanRecord,
		ReferenceFilePath: mockFile.Name(),
	}

	suite.MockReportRecordClient.EXPECT().Update(reportRecord).Return(nil).Times(2)

	// When
	suite.ComparisonDataService.GenerateComparisonDataForReport(reportRecord)

	// Then
	suite.NotNil(reportRecord)
	suite.Equal("failed", string(reportRecord.Status))
}

func (suite *ComparisonDataServiceTestSuite) TestScannedLocationOccupiedMatched() {
	// Given
	bulkScanRecordID := uint(1)
	reportRecordID := uint(2)

	location := "Location1"
	barcode := "Barcode1"

	scan := &models.Scan{
		Scanned:  true,
		Occupied: true,
		Barcodes: []string{"Barcode1"},
	}

	suite.MockScanClient.EXPECT().Get(bulkScanRecordID, location).Return(scan, nil)
	suite.MockComparisonDataClient.EXPECT().Create(gomock.Any()).Return(nil)

	// When
	comparisonData, err := suite.ComparisonDataService.createComparisonData(bulkScanRecordID, reportRecordID, location, barcode)

	// Then
	suite.Require().NoError(err)
	suite.NotNil(comparisonData)
	suite.Equal(uint(2), comparisonData.ReportRecordID)
	suite.Equal("Location1", comparisonData.Location)
	suite.True(comparisonData.Scanned)
	suite.True(comparisonData.Occupied)
	suite.EqualValues([]string{"Barcode1"}, comparisonData.ActualBarcodes)
	suite.EqualValues([]string{"Barcode1"}, comparisonData.ExpectedBarcodes)
	suite.Equal("The location was occupied by the expected items", string(comparisonData.Result))
}

func (suite *ComparisonDataServiceTestSuite) TestScannedLocationOccupiedWithBarcodeNotIdentified() {
	// Given
	bulkScanRecordID := uint(1)
	reportRecordID := uint(2)

	location := "Location1"
	barcode := "Barcode1"

	scan := &models.Scan{
		Scanned:  true,
		Occupied: true,
		Barcodes: []string{},
	}

	suite.MockScanClient.EXPECT().Get(bulkScanRecordID, location).Return(scan, nil)
	suite.MockComparisonDataClient.EXPECT().Create(gomock.Any()).Return(nil)

	// When
	comparisonData, err := suite.ComparisonDataService.createComparisonData(bulkScanRecordID, reportRecordID, location, barcode)

	// Then
	suite.Require().NoError(err)
	suite.NotNil(comparisonData)
	suite.Equal(uint(2), comparisonData.ReportRecordID)
	suite.Equal("Location1", comparisonData.Location)
	suite.True(comparisonData.Scanned)
	suite.True(comparisonData.Occupied)
	suite.EqualValues([]string{}, comparisonData.ActualBarcodes)
	suite.EqualValues([]string{"Barcode1"}, comparisonData.ExpectedBarcodes)
	suite.Equal("The location was occupied, but no barcode could be identified", string(comparisonData.Result))
}

func (suite *ComparisonDataServiceTestSuite) TestScannedLocationOccupiedMisMatch() {
	// Given
	bulkScanRecordID := uint(1)
	reportRecordID := uint(2)

	location := "Location1"
	barcode := "Barcode2"

	scan := &models.Scan{
		Scanned:  true,
		Occupied: true,
		Barcodes: []string{"Barcode1"},
	}

	suite.MockScanClient.EXPECT().Get(bulkScanRecordID, location).Return(scan, nil)
	suite.MockComparisonDataClient.EXPECT().Create(gomock.Any()).Return(nil)

	// When
	comparisonData, err := suite.ComparisonDataService.createComparisonData(bulkScanRecordID, reportRecordID, location, barcode)

	// Then
	suite.Require().NoError(err)
	suite.NotNil(comparisonData)
	suite.Equal(uint(2), comparisonData.ReportRecordID)
	suite.Equal("Location1", comparisonData.Location)
	suite.True(comparisonData.Scanned)
	suite.True(comparisonData.Occupied)
	suite.EqualValues([]string{"Barcode1"}, comparisonData.ActualBarcodes)
	suite.EqualValues([]string{"Barcode2"}, comparisonData.ExpectedBarcodes)
	suite.Equal("The location was occupied by the wrong items", string(comparisonData.Result))
}

func (suite *ComparisonDataServiceTestSuite) TestScannedLocationOccupiedByMultipleItems() {
	// Given
	bulkScanRecordID := uint(1)
	reportRecordID := uint(2)

	location := "Location1"
	barcode := "Barcode2"

	scan := &models.Scan{
		Scanned:  true,
		Occupied: true,
		Barcodes: []string{"Barcode1", "Barcode2"},
	}

	suite.MockScanClient.EXPECT().Get(bulkScanRecordID, location).Return(scan, nil)
	suite.MockComparisonDataClient.EXPECT().Create(gomock.Any()).Return(nil)

	// When
	comparisonData, err := suite.ComparisonDataService.createComparisonData(bulkScanRecordID, reportRecordID, location, barcode)

	// Then
	suite.Require().NoError(err)
	suite.NotNil(comparisonData)
	suite.Equal(uint(2), comparisonData.ReportRecordID)
	suite.Equal("Location1", comparisonData.Location)
	suite.True(comparisonData.Scanned)
	suite.True(comparisonData.Occupied)
	suite.EqualValues([]string{"Barcode1", "Barcode2"}, comparisonData.ActualBarcodes)
	suite.EqualValues([]string{"Barcode2"}, comparisonData.ExpectedBarcodes)
	suite.Equal("The location was occupied by the wrong items", string(comparisonData.Result))
}

func (suite *ComparisonDataServiceTestSuite) TestScannedLocationOccupiedButExpectedEmpty() {
	// Given
	bulkScanRecordID := uint(1)
	reportRecordID := uint(2)

	location := "Location1"
	barcode := ""

	scan := &models.Scan{
		Scanned:  true,
		Occupied: true,
		Barcodes: []string{"Barcode1"},
	}

	suite.MockScanClient.EXPECT().Get(bulkScanRecordID, location).Return(scan, nil)
	suite.MockComparisonDataClient.EXPECT().Create(gomock.Any()).Return(nil)

	// When
	comparisonData, err := suite.ComparisonDataService.createComparisonData(bulkScanRecordID, reportRecordID, location, barcode)

	// Then
	suite.Require().NoError(err)
	suite.NotNil(comparisonData)
	suite.Equal(uint(2), comparisonData.ReportRecordID)
	suite.Equal("Location1", comparisonData.Location)
	suite.True(comparisonData.Scanned)
	suite.True(comparisonData.Occupied)
	suite.EqualValues([]string{"Barcode1"}, comparisonData.ActualBarcodes)
	suite.EqualValues([]string{}, comparisonData.ExpectedBarcodes)
	suite.Equal("The location was occupied by an item, but should have been empty", string(comparisonData.Result))
}

func (suite *ComparisonDataServiceTestSuite) TestScannedLocationEmptyMatch() {
	// Given
	bulkScanRecordID := uint(1)
	reportRecordID := uint(2)

	location := "Location1"
	barcode := ""

	scan := &models.Scan{
		Scanned:  true,
		Occupied: false,
		Barcodes: []string{},
	}

	suite.MockScanClient.EXPECT().Get(bulkScanRecordID, location).Return(scan, nil)
	suite.MockComparisonDataClient.EXPECT().Create(gomock.Any()).Return(nil)

	// When
	comparisonData, err := suite.ComparisonDataService.createComparisonData(bulkScanRecordID, reportRecordID, location, barcode)

	// Then
	suite.Require().NoError(err)
	suite.NotNil(comparisonData)
	suite.Equal(uint(2), comparisonData.ReportRecordID)
	suite.Equal("Location1", comparisonData.Location)
	suite.True(comparisonData.Scanned)
	suite.False(comparisonData.Occupied)
	suite.EqualValues([]string{}, comparisonData.ActualBarcodes)
	suite.EqualValues([]string{}, comparisonData.ExpectedBarcodes)
	suite.Equal("The location was empty, as expected", string(comparisonData.Result))
}

func (suite *ComparisonDataServiceTestSuite) TestScannedLocationEmptyButNotExpected() {
	// Given
	bulkScanRecordID := uint(1)
	reportRecordID := uint(2)

	location := "Location1"
	barcode := "Barcode1"

	scan := &models.Scan{
		Scanned:  true,
		Occupied: false,
		Barcodes: []string{},
	}

	suite.MockScanClient.EXPECT().Get(bulkScanRecordID, location).Return(scan, nil)
	suite.MockComparisonDataClient.EXPECT().Create(gomock.Any()).Return(nil)

	// When
	comparisonData, err := suite.ComparisonDataService.createComparisonData(bulkScanRecordID, reportRecordID, location, barcode)

	// Then
	suite.Require().NoError(err)
	suite.NotNil(comparisonData)
	suite.Equal(uint(2), comparisonData.ReportRecordID)
	suite.Equal("Location1", comparisonData.Location)
	suite.True(comparisonData.Scanned)
	suite.False(comparisonData.Occupied)
	suite.EqualValues([]string{}, comparisonData.ActualBarcodes)
	suite.EqualValues([]string{"Barcode1"}, comparisonData.ExpectedBarcodes)
	suite.Equal("The location was empty, but it should have been occupied", string(comparisonData.Result))
}

func (suite *ComparisonDataServiceTestSuite) TestFailToGetScan() {
	// Given
	bulkScanRecordID := uint(1)
	reportRecordID := uint(2)

	location := "Location1"
	barcode := "Barcode1"

	suite.MockScanClient.EXPECT().Get(bulkScanRecordID, location).Return(nil, fmt.Errorf("error message"))

	// When
	comparisonData, err := suite.ComparisonDataService.createComparisonData(bulkScanRecordID, reportRecordID, location, barcode)

	// Then
	suite.Require().Error(err)
	suite.Nil(comparisonData)
	suite.Equal("failed to get scan with bulk scan record id=1 and location=Location1, error: error message", err.Error())
}

func (suite *ComparisonDataServiceTestSuite) TestNoScanFound() {
	// Given
	bulkScanRecordID := uint(1)
	reportRecordID := uint(2)

	location := "Location1"
	barcode := "Barcode1"

	suite.MockScanClient.EXPECT().Get(bulkScanRecordID, location).Return(nil, nil)

	// When
	comparisonData, err := suite.ComparisonDataService.createComparisonData(bulkScanRecordID, reportRecordID, location, barcode)

	// Then
	suite.Require().Error(err)
	suite.Nil(comparisonData)
	suite.Equal("scan not found with bulk scan record id=1 and location=Location1", err.Error())
}

func (suite *ComparisonDataServiceTestSuite) createMockCSVFile(lines []string) *os.File {
	file, err := os.CreateTemp("", "test*.csv")
	suite.Require().NoError(err)

	for _, line := range lines {
		_, err = file.WriteString(line + "\n")
		suite.Require().NoError(err)
	}

	err = file.Close()
	suite.Require().NoError(err)

	return file
}

package export

import (
	"github.com/golang/mock/gomock"
	mockexportreportservice "github.com/habbas99/dexory/generated/services/export"
	"github.com/habbas99/dexory/internal/models"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type ExportReportServiceTestSuite struct {
	suite.Suite
	MockExportReportRecordClient *mockexportreportservice.MockexportReportRecordClient
	MockComparisonDataClient     *mockexportreportservice.MockcomparisonDataClient
	ExportReportService          *ExportReportService
	tempFilePath                 string
	ctrl                         *gomock.Controller
}

func TestExportReportServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ExportReportServiceTestSuite))
}

func (suite *ExportReportServiceTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.MockExportReportRecordClient = mockexportreportservice.NewMockexportReportRecordClient(suite.ctrl)
	suite.MockComparisonDataClient = mockexportreportservice.NewMockcomparisonDataClient(suite.ctrl)

	suite.ExportReportService = NewExportReportService(suite.MockExportReportRecordClient, suite.MockComparisonDataClient)

	// create a temporary file to simulate the export report file
	file, err := os.CreateTemp("", "export_report_*.json")
	suite.Require().NoError(err)
	suite.tempFilePath = file.Name()
	file.Close()
}

func (suite *ExportReportServiceTestSuite) TearDownTest() {
	os.Remove(suite.tempFilePath)
	suite.ctrl.Finish()
}

func (suite *ExportReportServiceTestSuite) TestExportReport() {
	// Given
	reportRecordID := uint(1)
	exportReportRecord := &models.ExportReportRecord{
		ReportType:     models.ExportReportJson,
		FilePath:       suite.tempFilePath,
		ReportRecordID: reportRecordID,
		Status:         models.Pending,
	}
	exportReportRecord.ID = uint(1)

	comparisonData := []models.ComparisonData{
		{
			ReportRecordID:   reportRecordID,
			Location:         "Location1",
			Scanned:          true,
			Occupied:         true,
			ActualBarcodes:   []string{"Barcode1"},
			ExpectedBarcodes: []string{"Barcode1"},
			Result:           models.LocationOccupiedWithCorrectItems,
		},
		{
			ReportRecordID:   reportRecordID,
			Location:         "Location2",
			Scanned:          true,
			Occupied:         true,
			ActualBarcodes:   []string{"Barcode2"},
			ExpectedBarcodes: []string{"Barcode2"},
			Result:           models.LocationOccupiedWithCorrectItems,
		},
	}

	suite.MockComparisonDataClient.EXPECT().GetAllPaginated(reportRecordID, 50, 0).Return(comparisonData, nil).Times(1)
	suite.MockComparisonDataClient.EXPECT().GetAllPaginated(reportRecordID, 50, 2).Return([]models.ComparisonData{}, nil).Times(1)

	suite.MockExportReportRecordClient.EXPECT().Update(gomock.Any()).Times(2)

	// When
	suite.ExportReportService.ExportReport(exportReportRecord)

	// Then
	suite.Equal(models.Completed, exportReportRecord.Status)

	fileContents, err := os.ReadFile(suite.tempFilePath)
	suite.Require().NoError(err)
	suite.JSONEq(`[{
		"location":"Location1",
		"scanned":true,
		"occupied":true,
		"actualBarcodes":["Barcode1"],
		"expectedBarcodes":["Barcode1"],
		"result":"The location was occupied by the expected items"
	},{
		"location":"Location2",
		"scanned":true,
		"occupied":true,
		"actualBarcodes":["Barcode2"],
		"expectedBarcodes":["Barcode2"],
		"result":"The location was occupied by the expected items"
	}]`, string(fileContents))
}

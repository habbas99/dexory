package report

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/habbas99/dexory/internal/models"
	"github.com/habbas99/dexory/internal/utilities"
)

type reportRecordResponse struct {
	ID                uint      `json:"id"`
	BulkScanFileName  string    `json:"bulkScanFileName"`
	ReferenceFileName string    `json:"referenceFileName"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type comparisonDataResponse struct {
	Location         string   `json:"location"`
	Scanned          bool     `json:"scanned"`
	Occupied         bool     `json:"occupied"`
	ActualBarcodes   []string `json:"actualBarcodes"`
	ExpectedBarcodes []string `json:"expectedBarcodes"`
	Result           string   `json:"result"`
}

type fileStorageClient interface {
	SaveFile(dirPath, fileName string, fileContent io.Reader) (*os.File, error)
}

type reportRecordClient interface {
	GetAll() ([]models.ReportRecord, error)
	Create(bulkScanRecord models.BulkScanRecord, referenceFilePath string) (*models.ReportRecord, error)
	Get(reportRecordID uint) (*models.ReportRecord, error)
}

type bulkScanRecordClient interface {
	GetByFileName(fileName string) (*models.BulkScanRecord, error)
}

type comparisonDataClient interface {
	GetAllPaginated(reportRecordID uint, limit int, offset int) ([]models.ComparisonData, error)
}

type comparisonDataServiceClient interface {
	GenerateComparisonDataForReport(reportRecord *models.ReportRecord)
}

type ReportRecordController struct {
	dirPath                     string
	fileStorageClient           fileStorageClient
	bulkScanRecordClient        bulkScanRecordClient
	reportRecordClient          reportRecordClient
	comparisonDataClient        comparisonDataClient
	comparisonDataServiceClient comparisonDataServiceClient
}

func NewReportRecordController(
	dirPath string,
	fileStorageClient fileStorageClient,
	BulkScanRecordClient bulkScanRecordClient,
	reportRecordClient reportRecordClient,
	comparisonDataClient comparisonDataClient,
	comparisonDataServiceClient comparisonDataServiceClient,
) *ReportRecordController {
	return &ReportRecordController{
		dirPath:                     dirPath,
		fileStorageClient:           fileStorageClient,
		bulkScanRecordClient:        BulkScanRecordClient,
		reportRecordClient:          reportRecordClient,
		comparisonDataClient:        comparisonDataClient,
		comparisonDataServiceClient: comparisonDataServiceClient,
	}
}

func (rr *ReportRecordController) GetAllReportRecords(c *gin.Context) {
	log.Info("received request to get all report records")

	reportRecords, err := rr.reportRecordClient.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get report records from database"})
		return
	}

	reportResponses := []reportRecordResponse{}
	for _, reportRecord := range reportRecords {
		reportResponse := reportRecordResponse{
			ID:                reportRecord.ID,
			BulkScanFileName:  reportRecord.BulkScanRecord.FileName,
			ReferenceFileName: reportRecord.ReferenceFileName,
			Status:            string(reportRecord.Status),
			CreatedAt:         reportRecord.CreatedAt,
			UpdatedAt:         reportRecord.UpdatedAt,
		}
		reportResponses = append(reportResponses, reportResponse)
	}

	c.JSON(http.StatusOK, reportResponses)
}

func (rr *ReportRecordController) CreateReportRecord(c *gin.Context) {
	log.Info("received request to create report record")

	bulkScanFileName := c.PostForm("bulkScanFileName")
	fileHeader, err := c.FormFile("csvFile")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file is received"})
		return
	}

	receivedFile, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open the csv file"})
		return
	}
	defer receivedFile.Close()

	bulkScanRecord, err := rr.bulkScanRecordClient.GetByFileName(bulkScanFileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find bulk scan record"})
		return
	}

	if bulkScanRecord.Status != models.Completed {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bulk scan record has not been processed"})
		return
	}

	savedFile, err := rr.fileStorageClient.SaveFile(rr.dirPath, fileHeader.Filename, receivedFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save the csv file"})
		return
	}

	reportRecord, err := rr.reportRecordClient.Create(*bulkScanRecord, savedFile.Name())

	log.WithFields(log.Fields{
		"report_record_id":    reportRecord.ID,
		"bulk_scan_file_name": bulkScanFileName,
		"uploaded_file_name":  fileHeader.Filename,
	}).Info("trigger generate comparison data for report")

	// start a go routine to generate comparison report data
	go rr.comparisonDataServiceClient.GenerateComparisonDataForReport(reportRecord)

	c.JSON(http.StatusOK, gin.H{"id": reportRecord.ID})
}

func (rr *ReportRecordController) GetReport(c *gin.Context) {
	id := c.Param("id")

	log.WithFields(log.Fields{
		"report_record_id": id,
	}).Info("received request to get report record")

	reportRecordId, err := utilities.ToUint(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid report id"})
		return
	}

	reportRecord, err := rr.reportRecordClient.Get(reportRecordId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get report from database"})
		return
	}

	reportRecordResponse := reportRecordResponse{
		ID:                reportRecord.ID,
		BulkScanFileName:  reportRecord.BulkScanRecord.FileName,
		ReferenceFileName: reportRecord.ReferenceFileName,
		CreatedAt:         reportRecord.CreatedAt,
		UpdatedAt:         reportRecord.UpdatedAt,
		Status:            string(reportRecord.Status),
	}

	c.JSON(http.StatusOK, reportRecordResponse)
}

func (rr *ReportRecordController) GetComparisonData(c *gin.Context) {
	id := c.Param("id")

	log.WithFields(log.Fields{
		"report_record_id": id,
	}).Info("received request to get comparison data for report")

	reportRecordId, err := utilities.ToUint(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid report id"})
		return
	}

	comparisonDataList, err := rr.comparisonDataClient.GetAllPaginated(reportRecordId, 10000, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get comparison data for report from database"})
		return
	}

	comparisonDataResponses := []comparisonDataResponse{}
	for _, comparisonData := range comparisonDataList {
		comparisonDataResponse := comparisonDataResponse{
			Location:         comparisonData.Location,
			Scanned:          comparisonData.Scanned,
			Occupied:         comparisonData.Occupied,
			ActualBarcodes:   comparisonData.ActualBarcodes,
			ExpectedBarcodes: comparisonData.ExpectedBarcodes,
			Result:           string(comparisonData.Result),
		}
		comparisonDataResponses = append(comparisonDataResponses, comparisonDataResponse)
	}

	c.JSON(http.StatusOK, comparisonDataResponses)
}

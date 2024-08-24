package export

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/habbas99/dexory/internal/models"
	"github.com/habbas99/dexory/internal/utilities"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
)

type exportReportRecordRequest struct {
	ReportRecordID uint   `json:"reportRecordId"`
	ReportType     string `json:"reportType"`
}

type exportReportRecordResponse struct {
	ID       uint   `json:"id"`
	FileName string `json:"fileName"`
	Status   string `json:"status"`
}

type fileStorageClient interface {
	CreateFile(dirPath, fileName string) (*os.File, error)
}

type exportReportRecordClient interface {
	GetAll(reportRecordID uint) ([]models.ExportReportRecord, error)
	Create(reportRecordID uint, filePath, reportType string) (*models.ExportReportRecord, error)
	Get(exportReportRecordID uint) (*models.ExportReportRecord, error)
	GetByReportType(reportRecordID uint, reportType string) (*models.ExportReportRecord, error)
}

type exportReportServiceClient interface {
	ExportReport(exportReportRecord *models.ExportReportRecord)
}

type ExportReportController struct {
	dirPath                   string
	fileStorageClient         fileStorageClient
	exportReportRecordClient  exportReportRecordClient
	exportReportServiceClient exportReportServiceClient
}

func NewExportReportController(
	dirPath string,
	fileStorageClient fileStorageClient,
	exportReportRecordClient exportReportRecordClient,
	exportReportServiceClient exportReportServiceClient,
) *ExportReportController {
	return &ExportReportController{
		dirPath:                   dirPath,
		fileStorageClient:         fileStorageClient,
		exportReportRecordClient:  exportReportRecordClient,
		exportReportServiceClient: exportReportServiceClient,
	}
}

func (er *ExportReportController) GetExportReportRecords(c *gin.Context) {
	id := c.Param("id")
	log.WithFields(log.Fields{
		"report_record_id": id,
	}).Info("received request to get all exports for report record")

	reportRecordID, err := utilities.ToUint(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid report record id"})
		return
	}

	exportReportRecords, err := er.exportReportRecordClient.GetAll(reportRecordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get export report records from database"})
		return
	}

	exportReportRecordResponses := []exportReportRecordResponse{}
	for _, exportReportRecord := range exportReportRecords {
		response := exportReportRecordResponse{
			ID:       exportReportRecord.ID,
			FileName: exportReportRecord.FileName,
			Status:   string(exportReportRecord.Status),
		}
		exportReportRecordResponses = append(exportReportRecordResponses, response)
	}

	c.JSON(http.StatusOK, exportReportRecordResponses)
}

func (er *ExportReportController) CreateExportReportRecord(c *gin.Context) {
	log.Info("received request to create export report")

	var exportReportRecordReq exportReportRecordRequest
	if err := c.ShouldBindJSON(&exportReportRecordReq); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	reportRecordID := exportReportRecordReq.ReportRecordID
	reportType := exportReportRecordReq.ReportType

	log.WithFields(log.Fields{
		"report_record_id":   reportRecordID,
		"export_report_type": reportType,
	}).Info("received request to export report")

	if reportType != string(models.ExportReportJson) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "report type not supported"})
		return
	}

	exportReportRecord, err := er.exportReportRecordClient.GetByReportType(reportRecordID, reportType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find export report record"})
		return
	}

	if exportReportRecord != nil && exportReportRecord.Status != models.Failed {
		c.JSON(http.StatusOK, gin.H{"id": exportReportRecord.ID})
		return
	}

	savedFile, err := er.fileStorageClient.CreateFile(er.dirPath, fmt.Sprintf("report_%d.%s", reportRecordID, reportType))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create report file"})
		return
	}

	exportReportRecord, err = er.exportReportRecordClient.Create(reportRecordID, savedFile.Name(), reportType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create export report record"})
		return
	}

	// start a go routine to export report
	go er.exportReportServiceClient.ExportReport(exportReportRecord)

	c.JSON(http.StatusOK, gin.H{"id": exportReportRecord.ID})
}

func (er *ExportReportController) DownloadReport(c *gin.Context) {
	id := c.Param("id")

	log.WithFields(log.Fields{
		"export_report_record": id,
	}).Info("received request to download exported report file")

	exportReportRecordID, err := utilities.ToUint(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid export report record id"})
		return
	}

	exportReportRecord, err := er.exportReportRecordClient.Get(exportReportRecordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find export report record"})
		return
	}

	if exportReportRecord.Status != models.Completed {
		c.JSON(http.StatusAccepted, gin.H{"error": "report is not available for download"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(exportReportRecord.FilePath))

	c.File(exportReportRecord.FilePath)
}

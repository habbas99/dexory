package scan

import (
	"github.com/gin-gonic/gin"
	"github.com/habbas99/dexory/internal/models"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
)

type bulkScanRecordResponse struct {
	ID       uint   `json:"id"`
	FileName string `json:"fileName"`
	Status   string `json:"status"`
}

type fileStorageClient interface {
	SaveFile(dirPath, fileName string, fileContent io.Reader) (*os.File, error)
}

type bulkScanRecordClient interface {
	GetAll() ([]models.BulkScanRecord, error)
	Create(filePath string) (*models.BulkScanRecord, error)
}

type scanServiceClient interface {
	ProcessFile(bulkScanRecord *models.BulkScanRecord)
}

type ScanController struct {
	dirPath              string
	fileStorageClient    fileStorageClient
	bulkScanRecordClient bulkScanRecordClient
	scanServiceClient    scanServiceClient
}

func NewScanController(
	dirPath string,
	fileStorageClient fileStorageClient,
	bulkScanRecordClient bulkScanRecordClient,
	scanServiceClient scanServiceClient,
) *ScanController {
	return &ScanController{
		dirPath:              dirPath,
		fileStorageClient:    fileStorageClient,
		bulkScanRecordClient: bulkScanRecordClient,
		scanServiceClient:    scanServiceClient,
	}
}

func (sc *ScanController) GetBulkScanRecords(c *gin.Context) {
	log.Info("received request to get bulk scan records")

	bulkScanRecords, err := sc.bulkScanRecordClient.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get bulk scan records from database"})
		return
	}

	bulkScanRecordResponses := []bulkScanRecordResponse{}
	for _, bulkScanRecord := range bulkScanRecords {
		bulkScanRecordResponse := bulkScanRecordResponse{
			ID:       bulkScanRecord.ID,
			FileName: bulkScanRecord.FileName,
			Status:   string(bulkScanRecord.Status),
		}
		bulkScanRecordResponses = append(bulkScanRecordResponses, bulkScanRecordResponse)
	}

	c.JSON(http.StatusOK, bulkScanRecordResponses)
}

func (sc *ScanController) UploadBulkScanFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file not received"})
		return
	}

	log.WithFields(log.Fields{
		"filename": fileHeader.Filename,
	}).Info("received upload bulk scan file from robot")

	// open the file for reading
	receivedFile, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open the uploaded file"})
		return
	}
	defer receivedFile.Close()

	savedFile, err := sc.fileStorageClient.SaveFile(sc.dirPath, fileHeader.Filename, receivedFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save the file"})
		return
	}

	bulkScanRecord, err := sc.bulkScanRecordClient.Create(savedFile.Name())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start file processing"})
		return
	}

	// start a go routine to parse the JSON file
	go sc.scanServiceClient.ProcessFile(bulkScanRecord)

	c.JSON(http.StatusOK, gin.H{"id": bulkScanRecord.ID})
}

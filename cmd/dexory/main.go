package main

import (
	exportcontroller "github.com/habbas99/dexory/internal/controllers/export"
	"github.com/habbas99/dexory/internal/controllers/report"
	scancontroller "github.com/habbas99/dexory/internal/controllers/scan"
	"github.com/habbas99/dexory/internal/repositories"
	"github.com/habbas99/dexory/internal/services/comparison"
	exportservice "github.com/habbas99/dexory/internal/services/export"
	"github.com/habbas99/dexory/internal/services/file"
	scanservice "github.com/habbas99/dexory/internal/services/scan"
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/habbas99/dexory/internal/db"
	"github.com/joho/godotenv"
)

func init() {
	// set log output to standard output
	log.SetOutput(os.Stdout)

	// set log format with colors
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	// set log level to debug
	log.SetLevel(log.DebugLevel)
}

func main() {
	log.Info("starting server")

	// load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load .env file")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// initialize the database
	database, err := db.NewDatabase(host, port, username, password, dbName)
	if err != nil {
		log.Fatalf("failed making database connection, error: %v", err)
	}

	// Run the migrations
	err = database.Migrate()
	if err != nil {
		log.Fatalf("failed running database migration, error: %v", err)
	}

	bulkScanRecordRepository := repositories.NewBulkScanRecordRepository(database.DB)
	scanRepository := repositories.NewScanRepository(database.DB)
	reportRecordRepository := repositories.NewReportRecordRepository(database.DB)
	comparisonDataRepository := repositories.NewComparisonDataRepository(database.DB)
	exportReportRecordRepository := repositories.NewExportReportRecordRepository(database.DB)

	fileStorageService := file.NewFileStorageService()
	scanService := scanservice.NewScanService(bulkScanRecordRepository, scanRepository, 50)

	comparisonDataService := comparison.NewComparisonDataService(
		scanRepository, comparisonDataRepository, reportRecordRepository,
	)

	exportReportService := exportservice.NewExportReportService(exportReportRecordRepository, comparisonDataRepository)

	scanController := scancontroller.NewScanController(
		"./bulk-uploaded-scans",
		fileStorageService,
		bulkScanRecordRepository,
		scanService,
	)

	reportRecordController := report.NewReportRecordController(
		"./comparison-files",
		fileStorageService,
		bulkScanRecordRepository,
		reportRecordRepository,
		comparisonDataRepository,
		comparisonDataService,
	)

	exportReportController := exportcontroller.NewExportReportController(
		"./exported-reports", fileStorageService, exportReportRecordRepository, exportReportService,
	)

	// Setup Gin router
	router := gin.Default()

	// Check if we are in production mode
	env := os.Getenv("ENVIRONMENT")
	if env == "production" {
		// Serve static files in production
		router.Static("/static", "./frontend/build/static")
		router.StaticFile("/", "./frontend/build/index.html")
		router.NoRoute(func(c *gin.Context) {
			c.File("./frontend/build/index.html")
		})
	} else {
		// In development, redirect root to the React development server
		router.GET("/", func(c *gin.Context) {
			c.Redirect(307, "http://localhost:3000/")
		})

		// proxy API requests to the backend
		router.NoRoute(func(c *gin.Context) {
			c.JSON(404, gin.H{"error": "not found"})
		})
	}

	// routes
	router.GET("/bulk-scan-records", scanController.GetBulkScanRecords)
	router.POST("/upload-bulk-scan-file", scanController.UploadBulkScanFile)
	router.GET("/inventory-comparison-reports", reportRecordController.GetAllReportRecords)
	router.POST("/inventory-comparison-reports", reportRecordController.CreateReportRecord)
	router.GET("/inventory-comparison-reports/:id", reportRecordController.GetReport)
	router.GET("/inventory-comparison-reports/:id/data", reportRecordController.GetComparisonData)
	router.GET("/inventory-comparison-reports/:id/exports", exportReportController.GetExportReportRecords)
	router.POST("/export-report-records", exportReportController.CreateExportReportRecord)
	router.GET("/export-report-records/:id/download", exportReportController.DownloadReport)

	log.Info("server initialized")

	// Run the server
	router.Run(":8080")
}

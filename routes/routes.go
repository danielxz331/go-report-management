package routes

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"go-report-management/handlers"
	"go-report-management/services"
	"gorm.io/gorm"
	"sync"
)

func SetupRoutes(router *gin.Engine, db *sql.DB, dbormi *gorm.DB, reportQueue chan int, blockSize int) {
	router.POST("/login", func(c *gin.Context) { services.Login(c, dbormi) })
	router.POST("/refresh-token", func(c *gin.Context) { services.RefreshToken(c) })

	authorized := router.Group("/")
	authorized.Use(services.AuthenticateJWT())
	{
		authorized.GET("/report/:id", func(c *gin.Context) {
			handlers.GetReportDataPaginatedHandler(c, db)
		})

		authorized.GET("/report/:id/excel", func(c *gin.Context) {
			handlers.GenerateExcelReportHandler(c, db, reportQueue, blockSize)
		})

		authorized.POST("/reports", func(c *gin.Context) { handlers.CreateReportHandler(c, dbormi) })
		authorized.GET("/reports/:id", func(c *gin.Context) { handlers.GetReportByIDHandler(c, dbormi) })
		authorized.PUT("/reports/:id", func(c *gin.Context) { handlers.UpdateReportHandler(c, dbormi) })
		authorized.DELETE("/reports/:id", func(c *gin.Context) { handlers.DeleteReportHandler(c, dbormi) })
		authorized.GET("/reports", func(c *gin.Context) { handlers.ListReportsHandler(c, dbormi) })
	}
}

func ProcessReports(db *sql.DB, reportQueue chan int, semaphore chan struct{}, wg *sync.WaitGroup, blockSize int, filters map[string]string) {
	for reportID := range reportQueue {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(id int) {
			defer wg.Done()
			defer func() { <-semaphore }()
			services.GenerateReport(db, id, blockSize, filters)
		}(reportID)
	}
}

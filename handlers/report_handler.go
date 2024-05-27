package handlers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"go-report-management/cruds"
	"go-report-management/services"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

func CreateReportHandler(c *gin.Context, db *gorm.DB) {
	cruds.CreateReport(c, db)
}

func GetReportByIDHandler(c *gin.Context, db *gorm.DB) {
	cruds.GetReport(c, db)
}

func UpdateReportHandler(c *gin.Context, db *gorm.DB) {
	cruds.UpdateReport(c, db)
}

func DeleteReportHandler(c *gin.Context, db *gorm.DB) {
	cruds.DeleteReport(c, db)
}

func ListReportsHandler(c *gin.Context, db *gorm.DB) {
	cruds.ListReports(c, db)
}

func GenerateExcelReportHandler(c *gin.Context, db *sql.DB, reportQueue chan int, blockSize int) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}

	clientID := c.Param("clientid")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing clientID"})
		return
	}

	filters := extractFilters(c)
	reportQueue <- id
	go func() {
		services.GenerateReport(db, id, blockSize, filters, clientID)
	}()
	c.JSON(http.StatusAccepted, gin.H{"message": "Excel report generation in progress"})
}

func GetReportDataPaginatedHandler(c *gin.Context, db *sql.DB) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	filters := extractFilters(c)
	results, err := services.GetReportDataPaginated(db, id, limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":     page,
		"pageSize": limit,
		"results":  results,
	})
}

func extractFilters(c *gin.Context) map[string]string {
	filters := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if key != "id" && key != "page" && key != "limit" && key != "clientid" {
			filters[key] = values[0]
		}
	}
	return filters
}

package cruds

import (
	"github.com/gin-gonic/gin"
	"go-report-management/structs"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
)

func CreateReport(c *gin.Context, db *gorm.DB) {
	var report structs.SysMetaRpt
	if err := c.ShouldBindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Create(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, report)
}

func GetReport(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")
	var report structs.SysMetaRpt
	if err := db.First(&report, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}
	c.JSON(http.StatusOK, report)
}

func UpdateReport(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")
	var report structs.SysMetaRpt
	if err := db.First(&report, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	var update structs.SysMetaRpt
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.Model(&report).Updates(update)
	c.JSON(http.StatusOK, report)
}

func DeleteReport(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")
	if err := db.Delete(&structs.SysMetaRpt{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Report deleted successfully",
	})
}

func ListReports(c *gin.Context, db *gorm.DB) {
	pageStr := strings.TrimSpace(c.DefaultQuery("page", "1"))
	pageSizeStr := strings.TrimSpace(c.DefaultQuery("pageSize", "10"))

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	var reports []structs.SysMetaRpt
	offset := (page - 1) * pageSize

	result := db.Offset(offset).Limit(pageSize).Find(&reports)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":     page,
		"pageSize": pageSize,
		"results":  reports,
	})
}

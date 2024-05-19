package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-report-management/cruds"
	"go-report-management/database"
	"go-report-management/services"
	"log"
	"net/http"
	"strconv"
)

func main() {
	dataSourceName := "root:X0AhfRCK8GMeHfx2@tcp(db:3306)/umbrella-claro?charset=utf8&parseTime=True&loc=Local"
	db, err := database.InitconnectionSQL(dataSourceName)

	dbormi, err := database.InitconnectionGORM(dataSourceName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	router := gin.Default()

	router.POST("/login", func(c *gin.Context) { services.Login(c, dbormi) })

	authorized := router.Group("/")
	authorized.Use(services.AuthenticateJWT())
	{
		authorized.GET("/report/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
				return
			}

			query, whereClause, err := database.GetQueryByID(db, id)

			fmt.Println("ID: ", id)

			results, err := database.ExecuteQuery(db, query, whereClause)

			c.JSON(http.StatusOK, results)
		})

		authorized.POST("/reports", func(c *gin.Context) { cruds.CreateReport(c, dbormi) })
		authorized.GET("/reports/:id", func(c *gin.Context) { cruds.GetReport(c, dbormi) })
		authorized.PUT("/reports/:id", func(c *gin.Context) { cruds.UpdateReport(c, dbormi) })
		authorized.DELETE("/reports/:id", func(c *gin.Context) { cruds.DeleteReport(c, dbormi) })
		authorized.GET("/reports", func(c *gin.Context) { cruds.ListReports(c, dbormi) })
	}
	router.Run(":8080")
}

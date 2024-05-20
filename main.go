package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go-report-management/cruds"
	"go-report-management/database"
	"go-report-management/services"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	reportQueue = make(chan int, 10)     // Cola para manejar las solicitudes de informes
	semaphore   = make(chan struct{}, 2) // Limitar el número de goroutines concurrentes
	wg          sync.WaitGroup           // WaitGroup para esperar la finalización de todas las goroutines
)

const (
	blockSize = 250000 // Tamaño del bloque para cada consulta
)

func main() {
	db, err := database.InitconnectionSQL()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Configurar el pool de conexiones
	db.SetMaxOpenConns(20) // Número máximo de conexiones abiertas
	db.SetMaxIdleConns(10) // Número máximo de conexiones inactivas
	db.SetConnMaxLifetime(time.Minute * 3)

	dbormi, err := database.InitconnectionGORM()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	router := gin.Default()

	config := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}
	router.Use(cors.New(config))

	router.POST("/login", func(c *gin.Context) { services.Login(c, dbormi) })
	router.POST("/refresh-token", func(c *gin.Context) { services.RefreshToken(c) })

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
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting query by ID"})
				return
			}

			results, err := database.ExecuteQuery(db, query, whereClause, 0, blockSize)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "error executing query"})
				return
			}

			c.JSON(http.StatusOK, results)
		})

		authorized.GET("/report/:id/excel", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
				return
			}

			reportQueue <- id
			c.JSON(http.StatusAccepted, gin.H{"message": "Excel report generation in progress"})
		})

		authorized.POST("/reports", func(c *gin.Context) { cruds.CreateReport(c, dbormi) })
		authorized.GET("/reports/:id", func(c *gin.Context) { cruds.GetReport(c, dbormi) })
		authorized.PUT("/reports/:id", func(c *gin.Context) { cruds.UpdateReport(c, dbormi) })
		authorized.DELETE("/reports/:id", func(c *gin.Context) { cruds.DeleteReport(c, dbormi) })
		authorized.GET("/reports", func(c *gin.Context) { cruds.ListReports(c, dbormi) })
	}

	go processReports(db)

	router.Run(":8080")
	wg.Wait()
}

func processReports(db *sql.DB) {
	for reportID := range reportQueue {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore
		go func(id int) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			query, whereClause, err := database.GetQueryByID(db, id)
			if err != nil {
				log.Printf("error getting query by ID: %v", err)
				return
			}

			totalRows, err := getTotalRows(db, query, whereClause)
			if err != nil {
				log.Printf("error getting total rows: %v", err)
				return
			}

			chunks := totalRows / blockSize
			if totalRows%blockSize != 0 {
				chunks++
			}

			resultsChan := make(chan []map[string]interface{}, chunks)
			var wgChunks sync.WaitGroup

			// Crear el archivo Excel y escribir encabezados
			f := excelize.NewFile()
			headersWritten := false
			var headers []string
			sheetIndex := 1
			rowIndex := 1
			sheetName := fmt.Sprintf("Sheet%d", sheetIndex)
			f.NewSheet(sheetName)

			go func() {
				for results := range resultsChan {
					if !headersWritten {
						headers = make([]string, 0, len(results[0]))
						for header := range results[0] {
							headers = append(headers, header)
						}
						writeHeaders(f, headers, sheetName)
						headersWritten = true
					}

					writeResults(f, headers, results, &rowIndex, &sheetIndex, &sheetName)
				}
				filename, err := saveExcelFile(f, id)
				if err != nil {
					log.Printf("error saving Excel report: %v", err)
				} else {
					log.Printf("Excel file created successfully: %s", filename)
				}
			}()

			for i := 0; i < chunks; i++ {
				offset := i * blockSize
				wgChunks.Add(1)
				go func(offset int) {
					defer wgChunks.Done()
					results, err := database.ExecuteQuery(db, query, whereClause, offset, blockSize)
					if err != nil {
						log.Printf("error executing query block: %v", err)
						return
					}
					resultsChan <- results
				}(offset)
			}

			wgChunks.Wait()
			close(resultsChan)
		}(reportID)
	}
}

func getTotalRows(db *sql.DB, query, whereClause string) (int, error) {
	var totalRows int
	row := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM (%s WHERE %s) AS count_query", query, whereClause))
	err := row.Scan(&totalRows)
	return totalRows, err
}

func writeHeaders(f *excelize.File, headers []string, sheetName string) {
	for i, header := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheetName, col+"1", header)
	}
}

func writeResults(f *excelize.File, headers []string, results []map[string]interface{}, rowIndex, sheetIndex *int, sheetName *string) {
	for _, result := range results {
		*rowIndex++
		for colIdx, header := range headers {
			col := string(rune('A' + colIdx))
			cell := fmt.Sprintf("%s%d", col, *rowIndex)
			f.SetCellValue(*sheetName, cell, result[header])
		}
		if *rowIndex >= 1048576 {
			*sheetIndex++
			*rowIndex = 1
			*sheetName = fmt.Sprintf("Sheet%d", *sheetIndex)
			f.NewSheet(*sheetName)
			writeHeaders(f, headers, *sheetName)
		}
	}
	// Liberar la memoria del bloque de resultados una vez procesado
	results = nil
}

func saveExcelFile(f *excelize.File, reportID int) (string, error) {
	filename := fmt.Sprintf("report_%d.xlsx", reportID)
	filepath := filepath.Join("reports", filename)

	if err := os.MkdirAll("reports", os.ModePerm); err != nil {
		return "", err
	}

	if err := f.SaveAs(filepath); err != nil {
		return "", err
	}
	return filename, nil
}

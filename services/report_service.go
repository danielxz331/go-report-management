package services

import (
	"database/sql"
	"fmt"
	"github.com/xuri/excelize/v2"
	"go-report-management/utils"
	"log"
	"sync"
)

func GenerateReport(db *sql.DB, reportID int, blockSize int) {
	query, whereClause, err := GetQueryByID(db, reportID)
	if err != nil {
		log.Printf("error getting query by ID: %v", err)
		return
	}

	totalRows, err := GetTotalRows(db, query, whereClause)
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
				utils.WriteHeaders(f, headers, sheetName)
				headersWritten = true
			}

			utils.WriteResults(f, headers, results, &rowIndex, &sheetIndex, &sheetName)
		}
		filename, err := utils.SaveExcelFile(f, reportID)
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
			results, err := ExecuteQuery(db, query, whereClause, offset, blockSize)
			if err != nil {
				log.Printf("error executing query block: %v", err)
				return
			}
			resultsChan <- results
		}(offset)
	}

	wgChunks.Wait()
	close(resultsChan)
}

func GetReportDataPaginated(db *sql.DB, reportID, limit, offset int) ([]map[string]interface{}, error) {
	query, whereClause, err := GetQueryByID(db, reportID)
	if err != nil {
		return nil, fmt.Errorf("error getting query by ID: %v", err)
	}

	results, err := ExecuteQuery(db, query, whereClause, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}

	return results, nil
}

func GetQueryByID(db *sql.DB, id int) (string, string, error) {
	var query, whereClause string
	err := db.QueryRow("SELECT query, _where FROM sys_meta_rpt WHERE id = ?", id).Scan(&query, &whereClause)
	if err != nil {
		log.Printf("Error fetching query by ID: %v\n", err)
		return "", "", err
	}
	return query, whereClause, nil
}

func ExecuteQuery(db *sql.DB, query, whereClause string, offset, limit int) ([]map[string]interface{}, error) {
	paginatedQuery := fmt.Sprintf("%s WHERE %s LIMIT %d OFFSET %d", query, whereClause, limit, offset)
	rows, err := db.Query(paginatedQuery)
	if err != nil {
		log.Printf("Error executing query: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		log.Printf("Error getting columns: %v\n", err)
		return nil, err
	}

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			log.Printf("Error scanning row: %v\n", err)
			return nil, err
		}

		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			processedValue, err := utils.ProcessValue(*val)
			if err != nil {
				log.Printf("Error processing value: %v\n", err)
				return nil, err
			}
			m[colName] = processedValue
		}
		results = append(results, m)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error in rows: %v\n", err)
		return nil, err
	}

	return results, nil
}

func GetTotalRows(db *sql.DB, query, whereClause string) (int, error) {
	var totalRows int
	row := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM (%s WHERE %s) AS count_query", query, whereClause))
	err := row.Scan(&totalRows)
	return totalRows, err
}
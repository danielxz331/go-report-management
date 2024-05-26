package utils

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
	"path/filepath"
)

// WriteHeaders writes the headers to the excel file.
func WriteHeaders(f *excelize.File, headers []string, sheetName string) {
	for i, header := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheetName, col+"1", header)
	}
}

// WriteResults writes the results to the excel file.
func WriteResults(f *excelize.File, headers []string, results []map[string]interface{}, rowIndex, sheetIndex *int, sheetName *string) {
	for _, result := range results {
		*rowIndex++
		for colIdx, header := range headers {
			col := string(rune('A' + colIdx))
			cell := fmt.Sprintf("%s%d", col, *rowIndex)
			value, err := ProcessValue(result[header])
			if err != nil {
				value = fmt.Sprintf("error: %v", err)
			}
			f.SetCellValue(*sheetName, cell, value)
		}
		if *rowIndex >= 1048576 {
			*sheetIndex++
			*rowIndex = 1
			*sheetName = fmt.Sprintf("Sheet%d", *sheetIndex)
			f.NewSheet(*sheetName)
			WriteHeaders(f, headers, *sheetName)
		}
	}
	results = nil
}

// SaveExcelFile saves the excel file and uploads it to DigitalOcean Spaces.
func SaveExcelFile(f *excelize.File, reportID int) (string, error) {
	filename := fmt.Sprintf("report_%d.xlsx", reportID)
	localFilePath := filepath.Join("reports", filename)

	if err := os.MkdirAll("reports", os.ModePerm); err != nil {
		return "", err
	}

	if err := f.SaveAs(localFilePath); err != nil {
		return "", err
	}

	// Upload the file to DigitalOcean Space
	if err := UploadFileToSpace(localFilePath, "reports/"+filename); err != nil {
		return "", fmt.Errorf("error uploading Excel report to space: %v", err)
	}

	// Optionally delete the local file after uploading
	if err := os.Remove(localFilePath); err != nil {
		log.Printf("error deleting local file: %v", err)
	}

	return filename, nil
}

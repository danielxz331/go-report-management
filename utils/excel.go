package utils

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
	"path/filepath"
)

func WriteHeaders(f *excelize.File, headers []string, sheetName string) {
	for i, header := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheetName, col+"1", header)
	}
}

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

func SaveExcelFile(f *excelize.File, reportID int) (string, error) {
	uuid := uuid.New()
	filename := fmt.Sprintf("report_%d_%s.xlsx", reportID, uuid.String())
	localFilePath := filepath.Join("reports", filename)

	if err := os.MkdirAll("reports", os.ModePerm); err != nil {
		return "", err
	}

	if err := f.SaveAs(localFilePath); err != nil {
		return "", err
	}

	if err := UploadFileToSpace(localFilePath, "reports/"+filename); err != nil {
		return "", fmt.Errorf("error uploading Excel report to space: %v", err)
	}

	if err := os.Remove(localFilePath); err != nil {
		log.Printf("error deleting local file: %v", err)
	}

	return filename, nil
}

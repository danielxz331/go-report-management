package utils

import (
	"fmt"
	"github.com/xuri/excelize/v2"
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

package database

import (
	"database/sql"
	"fmt"
	"go-report-management/utils"
	"log"
)

func GetQueryByID(db *sql.DB, id int) (string, string, error) {
	var query, whereClause string
	err := db.QueryRow("SELECT query, _where FROM sys_meta_rpt WHERE id = ?", id).Scan(&query, &whereClause)
	if err != nil {
		log.Printf("Error fetching query by ID: %v\n", err)
		return "", "", err
	}
	return query, whereClause, nil
}

func ExecuteQuery(db *sql.DB, query, whereClause string) ([]map[string]interface{}, error) {
	if whereClause != "" {
		query += " WHERE " + whereClause + " LIMIT 10"
	}

	rows, err := db.Query(query)
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
			processedValue, err := utils.ProcessValue(val)
			fmt.Println(err)
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

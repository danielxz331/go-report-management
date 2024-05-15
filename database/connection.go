package database

import (
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitconnectionSQL(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	fmt.Println("Successfully connected to the database.")
	return db, nil
}

func InitconnectionGORM(dataSourceName string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	fmt.Println("Successfully connected to the database.")
	return db, nil
}

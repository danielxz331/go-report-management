package database

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
)

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func getDBConfig() (string, string, string, string, string, error) {
	host, hostExists := os.LookupEnv("MYSQL_HOST_U2")
	port, portExists := os.LookupEnv("MYSQL_PORT_U2")
	user, userExists := os.LookupEnv("MYSQL_USER_U2")
	password, passwordExists := os.LookupEnv("MYSQL_PASSWORD_U2")
	database, databaseExists := os.LookupEnv("MYSQL_DB_U2")

	if !hostExists || !portExists || !userExists || !passwordExists || !databaseExists {
		return "", "", "", "", "", fmt.Errorf("one or more environment variables are missing")
	}

	return host, port, user, password, database, nil
}

func InitconnectionSQL() (*sql.DB, error) {
	loadEnv()

	host, port, user, password, database, err := getDBConfig()
	if err != nil {
		return nil, err
	}

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", user, password, host, port, database)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return db, nil
}

func InitconnectionGORM() (*gorm.DB, error) {
	loadEnv()

	host, port, user, password, database, err := getDBConfig()
	if err != nil {
		return nil, err
	}

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", user, password, host, port, database)
	db, err := gorm.Open(mysql.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	return db, nil
}

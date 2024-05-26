package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go-report-management/database"
	"go-report-management/routes"
	"log"
	"sync"
	"time"
)

var (
	reportQueue = make(chan int, 10)
	semaphore   = make(chan struct{}, 2)
	wg          sync.WaitGroup
)

const (
	blockSize = 250000
)

func main() {
	db, err := database.InitconnectionSQL()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
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

	routes.SetupRoutes(router, db, dbormi, reportQueue, blockSize)

	// Filtros vacíos como ejemplo; ajusta según sea necesario
	filters := map[string]string{}

	go routes.ProcessReports(db, reportQueue, semaphore, &wg, blockSize, filters)

	router.Run(":8080")
	wg.Wait()
}

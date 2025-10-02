package server

import (
	"PVZ/internal/config"
	"PVZ/pkg/database"
	"PVZ/pkg/logger"
	"PVZ/pkg/metrics"
	"log"
)

func main() {
	logger.Init()
	metrics.Init()

	dsn := config.BuildDSNFromEnv()

	db, err := database.InitDB(dsn)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	err = db.Close()
	if err != nil {
		log.Printf("Failed to close database connection: %v", err)
		return
	}

}

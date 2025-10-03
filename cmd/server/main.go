package server

import (
	"PVZ/internal/config"
	"PVZ/internal/repository"
	"PVZ/internal/service"
	"PVZ/internal/transport/http/routers"
	"PVZ/pkg/database"
	"PVZ/pkg/logger"
	"PVZ/pkg/metrics"
	"log"
)

func main() {
	metrics.RegisterMetrics()
	logger.Init()

	cfg := config.Load()

	dsn := config.BuildDSNFromEnv()

	db, err := database.InitDB(dsn)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	userRepo := repository.NewUserRepo(db)
	pvzRepo := repository.NewPVZRepo(db)
	receptionRepo := repository.NewReceptionRepo(db)
	productRepo := repository.NewProductRepo(db)

	jwtKey := []byte(cfg.JWTSecret)
	userService := service.NewUserService(userRepo, jwtKey)
	pvzService := service.NewPVZService(pvzRepo)
	receptionService := service.NewReceptionService(receptionRepo)
	productService := service.NewProductService(productRepo, receptionRepo)

	r := routers.SetupRouter(
		receptionService,
		pvzService,
		productService,
		userService,
		jwtKey,
		logger.Log,
	)
	
	logger.Log.Printf("Starting server on :%s\n", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		logger.Log.Fatalf("server failed: %v", err)
	}

}

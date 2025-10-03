package main

import (
	"PVZ/internal/config"
	"PVZ/internal/repository"
	"PVZ/internal/service"
	"PVZ/internal/transport/http/routers"
	"PVZ/pkg/database"
	"PVZ/pkg/logger"
	"PVZ/pkg/metrics"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	metrics.RegisterMetrics()
	logger.Init()

	cfg := config.Load()

	dsn := config.BuildDSNFromEnv()

	db, err := database.InitDB(dsn)
	if err != nil {
		logger.Log.Fatalf("failed to init db: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Log.Printf("Failed to close database: %v", err)
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

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Log.Printf("Starting server on :%s\n", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatalf("server failed: %v", err)
		}
	}()

	<-quit
	logger.Log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Log.Println("Server exited gracefully")
}

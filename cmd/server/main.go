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
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	metrics.RegisterMetrics()
	logger.Init()
	log := logger.Log

	cfg := config.Load()

	dsn := config.BuildDSNFromEnv()

	db, err := database.InitDB(dsn)
	if err != nil {
		slog.Error("failed to init db: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database: %v", err)
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
		log,
	)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("Starting server on :%s\n", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed: %v", err)
		}
	}()

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Warn("Server forced to shutdown: %v", err)
	}

}

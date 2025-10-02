package main

import (
	"log"
	"net/http"
	"os"

	"PVZ/controllers"
	"PVZ/database"
	"PVZ/metrics"
	"PVZ/middleware"

	"github.com/gorilla/mux"
)

// setupLogger creates the logger
func setupLogger() *log.Logger {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	defer file.Close()

	logger := log.New(file, "APP: ", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Println("Logger initialized")
	return logger
}

var jwtKey = []byte("my_secret_key") // JWT signature key

func main() {
	metrics.RegisterMetrics()

	logger := setupLogger()

	db, err := database.InitDB("sqlite.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
		logger.Println("Failed to initialize database")
	}
	defer db.Close()

	router := mux.NewRouter()

	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware(jwtKey, logger, "employee", "moderator"))

	protected.HandleFunc("/pvz", controllers.CreatePVZ(db, logger)).Methods("POST")
	protected.HandleFunc("/pvz", controllers.GetPVZList(db, logger)).Methods("GET")
	protected.HandleFunc("/pvz/close_last_reception", controllers.CloseLastReception(db, logger)).Methods("POST")

	protected.HandleFunc("/receptions", controllers.CreateReception(db, logger)).Methods("POST")
	protected.HandleFunc("/receptions/close", controllers.CloseLastReception(db, logger)).Methods("POST")

	protected.HandleFunc("/products", controllers.CreateProduct(db, logger)).Methods("POST")

	router.HandleFunc("/dummyLogin", controllers.DummyLogin(jwtKey, logger)).Methods("POST")
	router.HandleFunc("/register", controllers.Register(db, logger)).Methods("POST")
	router.HandleFunc("/login", controllers.Login(db, jwtKey, logger)).Methods("POST")

	metricsRouter := mux.NewRouter()
	metricsRouter.Handle("/metrics", metrics.MetricsHandler())

	metricsRouter.Use(metrics.MetricsMiddleware)

	go func() {
		logger.Println("Prometheus server started on port 9000")
		log.Fatal(http.ListenAndServe(":9000", router))
	}()

	logger.Println("Server started on :8080")
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

package controllers

import (
	"PVZ/database"
	"PVZ/metrics"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// CreateReception creates a reception and adds it to the database
func CreateReception(db database.DBInterface, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Println("Received request to create a reception")

		userRole, ok := r.Context().Value("userRole").(string)
		if !ok || userRole != "employee" {
			logger.Printf("Access denied for role: %s", userRole)
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		var req struct {
			PvzId string `json:"pvzId"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			logger.Printf("Invalid request body: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.PvzId == "" {
			logger.Println("Invalid pvzId: empty value")
			http.Error(w, "Invalid pvzId", http.StatusBadRequest)
			return
		}

		var pvzExists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pvz WHERE id = ?)", req.PvzId).Scan(&pvzExists)
		if err != nil {
			logger.Printf("Failed to check PVZ existence: %v", err)
			http.Error(w, "Failed to check PVZ existence", http.StatusInternalServerError)
			return
		}
		if !pvzExists {
			logger.Printf("PVZ with ID %s does not exist", req.PvzId)
			http.Error(w, "PVZ does not exist", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			logger.Printf("Failed to start transaction: %v", err)
			http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var activeReceptionCount int
		err = tx.QueryRow(`
				SELECT COUNT(*)
				FROM receptions
				WHERE pvz_id = ? AND status = ?
			`, req.PvzId, "in_progress").Scan(&activeReceptionCount)
		if err != nil {
			logger.Printf("Failed to check active receptions: %v", err)
			http.Error(w, "Failed to check active receptions", http.StatusInternalServerError)
			return
		}
		if activeReceptionCount > 0 {
			logger.Printf("There is already an active reception for PVZ ID: %s", req.PvzId)
			http.Error(w, "There is already an active reception", http.StatusBadRequest)
			return
		}

		receptionID := GenerateUUID()
		dateTime := time.Now().Format(time.RFC3339)

		_, err = tx.Exec(`
				INSERT INTO receptions (id, date_time, pvz_id, status)
				VALUES (?, ?, ?, ?)
			`, receptionID, dateTime, req.PvzId, "in_progress")
		if err != nil {
			logger.Printf("Failed to create reception with ID %s: %v", receptionID, err)
			http.Error(w, "Failed to create reception: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			logger.Printf("Failed to commit transaction: %v", err)
			http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		logger.Printf("Reception created successfully with ID: %s", receptionID)

		metrics.ReceptionCreated.Inc()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":       receptionID,
			"dateTime": dateTime,
			"pvzId":    req.PvzId,
			"status":   "in_progress",
		})
	}
}

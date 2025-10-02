package controllers

import (
	"PVZ/database"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// GenerateUUID creates UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// CreateProduct processes the creation of a new product and adds it to the database
func CreateProduct(db database.DBInterface, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Println("Received request to add a product")

		userRole, ok := r.Context().Value("userRole").(string)
		if !ok || userRole != "employee" {
			logger.Printf("Access denied for role: %s", userRole)
			http.Error(w, "{\"error\": \"Access denied\"}", http.StatusForbidden)
			return
		}

		var req struct {
			Type  string `json:"type"`
			PvzId string `json:"pvzId"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			logger.Printf("Invalid request body: %v", err)
			http.Error(w, "{\"error\": \"Invalid request body\"}", http.StatusBadRequest)
			return
		}

		validTypes := map[string]bool{"электроника": true, "одежда": true, "обувь": true}
		if !validTypes[req.Type] {
			logger.Printf("Invalid product type: %s", req.Type)
			http.Error(w, "{\"error\": \"Invalid product type\"}", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			logger.Printf("Failed to start transaction: %v", err)
			http.Error(w, "{\"error\": \"Failed to start transaction\"}", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var receptionID string
		var productIDsJSON string
		err = tx.QueryRow(`
            SELECT id, product_ids FROM receptions
            WHERE pvz_id = ? AND status = ?
        `, req.PvzId, "in_progress").Scan(&receptionID, &productIDsJSON)
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Printf("No active reception found for PVZ ID: %s", req.PvzId)
				http.Error(w, "{\"error\": \"No active reception found\"}", http.StatusBadRequest)
				return
			}
			logger.Printf("Failed to query active reception: %v", err)
			http.Error(w, "{\"error\": \"Database error\"}", http.StatusInternalServerError)
			return
		}

		productID := GenerateUUID()
		dateTime := time.Now().Format(time.RFC3339)

		_, err = tx.Exec(`
            INSERT INTO products (id, date_time, type)
            VALUES (?, ?, ?)
        `, productID, dateTime, req.Type)
		if err != nil {
			logger.Printf("Failed to create product: %v", err)
			http.Error(w, "{\"error\": \"Failed to create product\"}", http.StatusInternalServerError)
			return
		}

		var productIDs []string
		if err := json.Unmarshal([]byte(productIDsJSON), &productIDs); err != nil {
			logger.Printf("Failed to parse product IDs: %v", err)
			http.Error(w, "{\"error\": \"Failed to parse product IDs\"}", http.StatusInternalServerError)
			return
		}
		productIDs = append(productIDs, productID)

		updatedProductIDsJSON, err := json.Marshal(productIDs)
		if err != nil {
			logger.Printf("Failed to serialize product IDs: %v", err)
			http.Error(w, "{\"error\": \"Failed to serialize product IDs\"}", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec("UPDATE receptions SET product_ids = ? WHERE id = ?", string(updatedProductIDsJSON), receptionID)
		if err != nil {
			logger.Printf("Failed to update reception: %v", err)
			http.Error(w, "{\"error\": \"Failed to update reception\"}", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			logger.Printf("Failed to commit transaction: %v", err)
			http.Error(w, "{\"error\": \"Failed to commit transaction\"}", http.StatusInternalServerError)
			return
		}

		logger.Printf("Product created successfully: %s", productID)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"id":       productID,
			"dateTime": dateTime,
			"type":     req.Type,
		})
	}
}

package controllers

import (
	"PVZ/database"
	"PVZ/metrics"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

// CreatePVZ creates new PVZ
func CreatePVZ(db database.DBInterface, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Println("Received request to create a PVZ")

		userRole, ok := r.Context().Value("userRole").(string)
		if !ok || userRole != "moderator" {
			logger.Printf("Access denied for role: %s", userRole)
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		var req struct {
			City string `json:"city"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			logger.Printf("Invalid request body: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		validCities := map[string]bool{
			"Москва":          true,
			"Санкт-Петербург": true,
			"Казань":          true,
		}
		if !validCities[req.City] {
			logger.Printf("Invalid city requested: %s", req.City)
			http.Error(w, "Invalid city. Allowed cities: Москва, Санкт-Петербург, Казань", http.StatusBadRequest)
			return
		}

		pvzID := GenerateUUID()
		registrationDate := time.Now().Format(time.RFC3339)

		result, err := db.Exec(`
            INSERT INTO pvz (id, city, registration_date)
            VALUES (?, ?, ?)
        `, pvzID, req.City, registrationDate)
		if err != nil {
			logger.Printf("Failed to create PVZ with ID %s: %v", pvzID, err)
			http.Error(w, "Failed to create PVZ: "+err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			logger.Printf("Failed to confirm PVZ creation for ID: %s", pvzID)
			http.Error(w, "Failed to create PVZ", http.StatusInternalServerError)
			return
		}

		logger.Printf("PVZ created successfully with ID: %s", pvzID)

		metrics.PVZCreated.Inc()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":               pvzID,
			"city":             req.City,
			"registrationDate": registrationDate,
		})
	}
}

// GetPVZList returns the list of PVZs filtered by date and pagination
func GetPVZList(db database.DBInterface, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userRole, ok := r.Context().Value("userRole").(string)
		logger.Printf("Extracted userRole from context: %v (type: %T)", userRole, userRole)

		if !ok {
			logger.Println("Failed to extract userRole from context")
			http.Error(w, "{\"error\": \"Access denied\"}", http.StatusForbidden)
			return
		}

		if userRole != "employee" && userRole != "moderator" {
			logger.Printf("Access denied for role: %s", userRole)
			http.Error(w, "{\"error\": \"Access denied\"}", http.StatusForbidden)
			return
		}

		logger.Println("Received request to fetch PVZ list")

		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil || page < 1 {
			page = 1
		}

		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit < 1 || limit > 30 {
			limit = 10
		}

		offset := (page - 1) * limit

		startDate := r.URL.Query().Get("startDate")
		endDate := r.URL.Query().Get("endDate")

		query := `
            SELECT id, registration_date, city
            FROM pvz
            WHERE (COALESCE(?, '') = '' OR registration_date >= ?)
              AND (COALESCE(?, '') = '' OR registration_date <= ?)
            LIMIT ? OFFSET ?
        `
		rows, err := db.Query(query, startDate, startDate, endDate, endDate, limit, offset)
		if err != nil {
			logger.Printf("Failed to fetch PVZ list: %v", err)
			http.Error(w, "Failed to fetch PVZ list", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var pvzList []map[string]any
		for rows.Next() {
			var pvz struct {
				ID               string `json:"id"`
				RegistrationDate string `json:"registrationDate"`
				City             string `json:"city"`
			}
			if err := rows.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City); err != nil {
				logger.Printf("Failed to scan PVZ data: %v", err)
				http.Error(w, "Failed to scan PVZ data", http.StatusInternalServerError)
				return
			}
			pvzList = append(pvzList, map[string]interface{}{
				"id":               pvz.ID,
				"registrationDate": pvz.RegistrationDate,
				"city":             pvz.City,
			})
		}

		logger.Printf("Successfully fetched PVZ list with %d items", len(pvzList))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pvzList)
	}
}

// DeleteLastProduct finds the last open reception for the specified PVZ and deletes the last product
func DeleteLastProduct(db database.DBInterface, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Println("Received request to delete the last product")

		userRole, ok := r.Context().Value("userRole").(string)
		if !ok || userRole != "employee" {
			logger.Printf("Access denied for role: %s", userRole)
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		var req struct {
			PvzID string `json:"pvzId"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			logger.Printf("Invalid request body: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if req.PvzID == "" {
			logger.Println("Invalid pvzId: empty value")
			http.Error(w, "Invalid pvzId", http.StatusBadRequest)
			return
		}

		var receptionID string
		var productIDsJSON string
		err = db.QueryRow(`
            SELECT id, product_ids FROM receptions
            WHERE pvz_id = ? AND status = 'in_progress'
            ORDER BY date_time DESC
            LIMIT 1
        `, req.PvzID).Scan(&receptionID, &productIDsJSON)
		if err != nil {
			logger.Printf("No active reception found for PVZ ID %s: %v", req.PvzID, err)
			http.Error(w, "No active reception found", http.StatusBadRequest)
			return
		}

		var productIDs []string
		if err := json.Unmarshal([]byte(productIDsJSON), &productIDs); err != nil {
			logger.Printf("Failed to parse product IDs for reception ID %s: %v", receptionID, err)
			http.Error(w, "Failed to parse product IDs", http.StatusInternalServerError)
			return
		}

		if len(productIDs) == 0 {
			logger.Printf("No products found in reception ID %s", receptionID)
			http.Error(w, "No products found in the current reception", http.StatusBadRequest)
			return
		}

		lastProductID := productIDs[len(productIDs)-1]
		productIDs = productIDs[:len(productIDs)-1]

		_, err = db.Exec("DELETE FROM products WHERE id = ?", lastProductID)
		if err != nil {
			logger.Printf("Failed to delete product ID %s: %v", lastProductID, err)
			http.Error(w, "Failed to delete product", http.StatusInternalServerError)
			return
		}

		updatedProductIDsJSON, err := json.Marshal(productIDs)
		if err != nil {
			logger.Printf("Failed to serialize updated product IDs for reception ID %s: %v", receptionID, err)
			http.Error(w, "Failed to serialize product IDs", http.StatusInternalServerError)
			return
		}
		_, err = db.Exec("UPDATE receptions SET product_ids = ? WHERE id = ?", string(updatedProductIDsJSON), receptionID)
		if err != nil {
			logger.Printf("Failed to update reception ID %s after deleting product ID %s: %v", receptionID, lastProductID, err)
			http.Error(w, "Failed to update reception after product deletion", http.StatusInternalServerError)
			return
		}

		logger.Printf("Product deleted successfully from reception ID %s", receptionID)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Product deleted"})
	}
}

// CloseLastReception closes the last open reception based on PvzId
func CloseLastReception(db database.DBInterface, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Println("Received request to close the last reception")

		userRole, ok := r.Context().Value("userRole").(string)
		if !ok || userRole != "employee" {
			logger.Printf("Access denied for role: %s", userRole)
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		var req struct {
			PvzID string `json:"pvzId"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			logger.Printf("Invalid request body: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.PvzID == "" {
			logger.Println("Invalid pvzId: empty value")
			http.Error(w, "Invalid pvzId", http.StatusBadRequest)
			return
		}

		var receptionID string
		err = db.QueryRow(`
            SELECT id FROM receptions
            WHERE pvz_id = ? AND status = 'in_progress'
            ORDER BY date_time DESC
            LIMIT 1
        `, req.PvzID).Scan(&receptionID)
		if err != nil {
			logger.Printf("No active reception found for PVZ ID %s: %v", req.PvzID, err)
			http.Error(w, "No active reception found", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("UPDATE receptions SET status = 'close' WHERE id = ?", receptionID)
		if err != nil {
			logger.Printf("Failed to close reception ID %s: %v", receptionID, err)
			http.Error(w, "Failed to close reception", http.StatusInternalServerError)
			return
		}

		logger.Printf("Reception closed successfully with ID %s", receptionID)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Reception closed"})
	}
}

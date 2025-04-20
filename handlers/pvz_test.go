package handlers

import (
	"PVZ/database"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreatePVZ(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialisation of the test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Create a query with correct data
	payload := map[string]string{"city": "Москва"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/pvz", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	// Add a role to the context
	ctx := req.Context()
	ctx = context.WithValue(ctx, "userRole", "moderator")
	req = req.WithContext(ctx)

	// Call the handler
	handler := CreatePVZ(db, logger)
	handler.ServeHTTP(rec, req)

	// Check status code
	assert.Equal(t, http.StatusCreated, rec.Code, "Expected status code 201 for valid PVZ creation")

	// Check the response body
	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response body")
	assert.NotEmpty(t, response["id"], "Expected non-empty PVZ ID in response")
}

func TestGetPVZList(t *testing.T) {
	// Initialisation of the test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Delete old tables (if they exist)
	_, err = db.Exec(`DROP TABLE IF EXISTS pvz;`)
	if err != nil {
		t.Fatalf("Failed to drop tables: %v", err)
	}

	// Create tables
	_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS pvz (
				id TEXT PRIMARY KEY,
				city TEXT NOT NULL,
				registration_date TEXT NOT NULL
			);
		`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Preparing test data
	pvzID1 := GenerateUUID()
	city1 := "Москва"
	registrationDate1 := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)

	pvzID2 := GenerateUUID()
	city2 := "Санкт-Петербург"
	registrationDate2 := time.Now().Format(time.RFC3339)

	_, err = db.Exec(`
			INSERT INTO pvz (id, city, registration_date)
			VALUES (?, ?, ?), (?, ?, ?)
		`, pvzID1, city1, registrationDate1, pvzID2, city2, registrationDate2)
	if err != nil {
		t.Fatalf("Failed to insert test PVZ data: %v", err)
	}

	// Initialising the logger
	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Test 1: Getting the list of PVZs without filtering
	req := httptest.NewRequest("GET", "/pvz?page=1&limit=10", nil)
	rec := httptest.NewRecorder()

	ctx := req.Context()
	ctx = context.WithValue(ctx, "userRole", "employee")
	req = req.WithContext(ctx)

	handler := GetPVZList(db, logger)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200 for successful request")

	var response []map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response body")
	assert.NotEmpty(t, response, "Expected non-empty PVZ list")

	// Test 2: Filtering by date
	startDate := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	endDate := time.Now().Format(time.RFC3339)

	req = httptest.NewRequest("GET", "/pvz?startDate="+startDate+"&endDate="+endDate+"&page=1&limit=10", nil)
	rec = httptest.NewRecorder()

	ctx = req.Context()
	ctx = context.WithValue(ctx, "userRole", "employee")
	req = req.WithContext(ctx)

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200 for filtered request")

	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response body")
	assert.NotEmpty(t, response, "Expected non-empty PVZ list after filtering")

	for _, item := range response {
		date := item["registrationDate"].(string)
		assert.True(t, date >= startDate && date <= endDate, "Expected registration date to be within the filter range")
	}
}

func TestDeleteLastProduct(t *testing.T) {
	// Initialisation of the test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Delete old tables (if they exist)
	_, err = db.Exec(`
			DROP TABLE IF EXISTS receptions;
			DROP TABLE IF EXISTS products;
		`)
	if err != nil {
		t.Fatalf("Failed to drop tables: %v", err)
	}

	// Create tables
	_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS receptions (
				id TEXT PRIMARY KEY,
				pvz_id TEXT NOT NULL,
				status TEXT NOT NULL,
				product_ids TEXT,
				date_time TEXT NOT NULL
			);
			CREATE TABLE IF NOT EXISTS products (
				id TEXT PRIMARY KEY,
				type TEXT NOT NULL,
				reception_id TEXT NOT NULL
			);
		`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Data preparation
	receptionID := "reception-1"
	pvzID := "pvz-1"
	productIDs := []string{"product-1", "product-2"}

	productIDsJSON, _ := json.Marshal(productIDs)

	_, err = db.Exec(`
			INSERT INTO receptions (id, pvz_id, status, product_ids, date_time)
			VALUES (?, ?, ?, ?, ?)
		`, receptionID, pvzID, "in_progress", string(productIDsJSON), time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to insert test reception: %v", err)
	}

	for _, productID := range productIDs {
		_, err := db.Exec(`
				INSERT INTO products (id, type, reception_id)
				VALUES (?, ?, ?)
			`, productID, "электроника", receptionID)
		if err != nil {
			t.Fatalf("Failed to insert test product: %v", err)
		}
	}

	// Initialising the logger
	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Test 1: Successful deletion of the last product
	payload := map[string]string{"pvzId": pvzID}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/pvz/delete_last_product", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	// Add a role to the context
	ctx := req.Context()
	ctx = context.WithValue(ctx, "userRole", "employee")
	req = req.WithContext(ctx)

	// Call the handler
	handler := DeleteLastProduct(db, logger)
	handler.ServeHTTP(rec, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200 for successful deletion")

	// Check the response body
	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response body")
	assert.Equal(t, "Product deleted", response["message"], "Expected success message")

	// Check that the product is deleted from the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM products WHERE id = ?", "product-2").Scan(&count)
	assert.NoError(t, err, "Failed to query database")
	assert.Equal(t, 0, count, "Expected product to be deleted from the database")

	// Test 2: No active reception
	payload = map[string]string{"pvzId": "non-existent-pvz"}
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest("POST", "/pvz/delete_last_product", bytes.NewReader(body))
	rec = httptest.NewRecorder()

	ctx = req.Context()
	ctx = context.WithValue(ctx, "userRole", "employee")
	req = req.WithContext(ctx)

	handler.ServeHTTP(rec, req)

	// Check status code
	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for no active reception")

	// Test 3: No products to delete
	_, err = db.Exec("DELETE FROM products")
	if err != nil {
		t.Fatalf("Failed to delete test products: %v", err)
	}

	// Update the product_ids field in the receptions table
	_, err = db.Exec("UPDATE receptions SET product_ids = '[]' WHERE id = ?", receptionID)
	if err != nil {
		t.Fatalf("Failed to update product_ids in receptions: %v", err)
	}

	payload = map[string]string{"pvzId": pvzID}
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest("POST", "/pvz/delete_last_product", bytes.NewReader(body))
	rec = httptest.NewRecorder()

	ctx = req.Context()
	ctx = context.WithValue(ctx, "userRole", "employee")
	req = req.WithContext(ctx)

	handler.ServeHTTP(rec, req)

	// Check status code
	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for no products in reception")
}

func TestCloseLastReception(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialisation of the test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Create table receptions
	_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS receptions (
				id TEXT PRIMARY KEY,
				pvz_id TEXT NOT NULL,
				status TEXT NOT NULL,
				date_time TEXT NOT NULL
			);
		`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Preparing test data
	receptionID := GenerateUUID()
	pvzID := GenerateUUID()
	dateTime := time.Now().Format(time.RFC3339)
	_, err = db.Exec(`
			INSERT INTO receptions (id, pvz_id, status, date_time)
			VALUES (?, ?, ?, ?)
		`, receptionID, pvzID, "in_progress", dateTime)
	if err != nil {
		t.Fatalf("Failed to insert test reception: %v", err)
	}

	// Test 1: Successful closing of reception
	payload := map[string]string{"pvzId": pvzID}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/pvz/close_last_reception", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	// Add a role to the context
	ctx := req.Context()
	ctx = context.WithValue(ctx, "userRole", "employee")
	req = req.WithContext(ctx)

	// Call the handler
	handler := CloseLastReception(db, logger)
	handler.ServeHTTP(rec, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200 for successful closure")

	// Check the response body
	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response body")
	assert.Equal(t, "Reception closed", response["message"], "Expected success message")

	// Check that the reception is closed in the database
	var status string
	err = db.QueryRow("SELECT status FROM receptions WHERE id = ?", receptionID).Scan(&status)
	assert.NoError(t, err, "Failed to query database")
	assert.Equal(t, "close", status, "Expected reception to be closed in the database")

	// Test 2: No active reception
	payload = map[string]string{"pvzId": "non-existent-pvz"}
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest("POST", "/pvz/close_last_reception", bytes.NewReader(body))
	rec = httptest.NewRecorder()

	ctx = req.Context()
	ctx = context.WithValue(ctx, "userRole", "employee")
	req = req.WithContext(ctx)

	handler.ServeHTTP(rec, req)

	// Check status code
	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for no active reception")

	// Test 3: Insufficient access rights
	payload = map[string]string{"pvzId": pvzID}
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest("POST", "/pvz/close_last_reception", bytes.NewReader(body))
	rec = httptest.NewRecorder()

	ctx = req.Context()
	ctx = context.WithValue(ctx, "userRole", "moderator")
	req = req.WithContext(ctx)

	handler.ServeHTTP(rec, req)

	// Check status code
	assert.Equal(t, http.StatusForbidden, rec.Code, "Expected status code 403 for insufficient permissions")
}

func TestGetPVZList_BoundaryCases(t *testing.T) {
	// Initialisation of the test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Create table pvz
	_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS pvz (
				id TEXT PRIMARY KEY,
				city TEXT NOT NULL,
				registration_date TEXT NOT NULL
			);
		`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Preparing test data
	pvzID := GenerateUUID()
	city := "Москва"
	registrationDate := time.Now().Format(time.RFC3339)
	_, err = db.Exec(`
			INSERT INTO pvz (id, city, registration_date)
			VALUES (?, ?, ?)
		`, pvzID, city, registrationDate)
	if err != nil {
		t.Fatalf("Failed to insert test PVZ data: %v", err)
	}

	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Test 1: Minimum values for page and limit
	req := httptest.NewRequest("GET", "/pvz?page=1&limit=1", nil)
	rec := httptest.NewRecorder()
	ctx := req.Context()
	ctx = context.WithValue(ctx, "userRole", "employee")
	req = req.WithContext(ctx)

	handler := GetPVZList(db, logger)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200 for valid pagination")
	var response []map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response body")
	assert.Len(t, response, 1, "Expected exactly one item in the response")

	// Test 2: Maximum value for limit
	req = httptest.NewRequest("GET", "/pvz?page=1&limit=30", nil)
	rec = httptest.NewRecorder()
	ctx = req.Context()
	ctx = context.WithValue(ctx, "userRole", "employee")
	req = req.WithContext(ctx)

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200 for valid pagination")
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response body")
	assert.Len(t, response, 1, "Expected exactly one item in the response")

	// Test 3: Incorrect values for page and limit
	req = httptest.NewRequest("GET", "/pvz?page=0&limit=0", nil)
	rec = httptest.NewRecorder()
	ctx = req.Context()
	ctx = context.WithValue(ctx, "userRole", "employee")
	req = req.WithContext(ctx)

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200 with default pagination")
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response body")
	assert.Len(t, response, 1, "Expected exactly one item in the response")
}

func TestInvalidDataHandling(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialisation of the test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Test 1: Inadmissible city when creating a PVZ
	payload := map[string]string{"city": "НевалидныйГород"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/pvz", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	ctx := req.Context()
	ctx = context.WithValue(ctx, "userRole", "moderator")
	req = req.WithContext(ctx)

	handler := CreatePVZ(db, logger)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for invalid city")
}

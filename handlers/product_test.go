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

func TestCreateProduct(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialize test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Create tables
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS receptions (
            id TEXT PRIMARY KEY,
            date_time TEXT NOT NULL,
            pvz_id TEXT NOT NULL,
            status TEXT NOT NULL,
            product_ids TEXT NOT NULL
        );
        CREATE TABLE IF NOT EXISTS products (
            id TEXT PRIMARY KEY,
            date_time TEXT NOT NULL,
            type TEXT NOT NULL
        );
    `)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Add an active reception
	receptionID := GenerateUUID()
	pvzID := GenerateUUID()
	_, err = db.Exec(`
        INSERT INTO receptions (id, date_time, pvz_id, status, product_ids)
        VALUES (?, ?, ?, ?, ?)
    `, receptionID, time.Now().Format(time.RFC3339), pvzID, "in_progress", "[]")
	if err != nil {
		t.Fatalf("Failed to insert test reception: %v", err)
	}

	// Test 1: Successful product addition
	payload := map[string]string{"type": "электроника", "pvzId": pvzID}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/products", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	// Add role to context
	ctx := req.Context()
	ctx = context.WithValue(ctx, "userRole", "employee")
	req = req.WithContext(ctx)

	// Call the handler
	handler := CreateProduct(db, logger)
	handler.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusCreated, rec.Code, "Expected status code 201 for valid product addition")

	// Check the response body
	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response body")
	assert.NotEmpty(t, response["id"], "Expected non-empty product ID in response")
	assert.Equal(t, "электроника", response["type"], "Expected correct type in response")
}

func TestCreateProduct_Errors(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialize test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Create tables
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS receptions (
            id TEXT PRIMARY KEY,
            date_time TEXT NOT NULL,
            pvz_id TEXT NOT NULL,
            status TEXT NOT NULL,
            product_ids TEXT NOT NULL
        );
        CREATE TABLE IF NOT EXISTS products (
            id TEXT PRIMARY KEY,
            date_time TEXT NOT NULL,
            type TEXT NOT NULL
        );
    `)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Test 1: Missing type or pvzId
	payload := map[string]string{"pvzId": "pvz-1"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/products", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	ctx := req.Context()
	ctx = context.WithValue(ctx, "userRole", "employee")
	req = req.WithContext(ctx)

	handler := CreateProduct(db, logger)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for missing type")

	// Test 2: Non-existent reception
	payload = map[string]string{"type": "электроника", "pvzId": "non-existent-pvz"}
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest("POST", "/products", bytes.NewReader(body))
	rec = httptest.NewRecorder()

	ctx = req.Context()
	ctx = context.WithValue(ctx, "userRole", "employee")
	req = req.WithContext(ctx)

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for non-existent reception")
}

package controllers

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

func TestCreateReception(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialisation of the test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Сreate tables
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS pvz (
            id TEXT PRIMARY KEY,
            city TEXT NOT NULL,
            registration_date TEXT NOT NULL
        );
        CREATE TABLE IF NOT EXISTS receptions (
            id TEXT PRIMARY KEY,
            date_time TEXT NOT NULL,
            pvz_id TEXT NOT NULL,
            status TEXT NOT NULL
        );
    `)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Add a test PVZ
	pvzID := GenerateUUID()
	_, err = db.Exec(`
        INSERT INTO pvz (id, city, registration_date)
        VALUES (?, ?, ?)
    `, pvzID, "Москва", time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to insert test PVZ: %v", err)
	}

	// Test 1: Successful establishment of reception
	payload := map[string]string{"pvzId": pvzID}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/receptions", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	// Add a role to the context
	ctx := req.Context()
	ctx = context.WithValue(ctx, "userRole", "employee")
	req = req.WithContext(ctx)

	// Call the handler
	handler := CreateReception(db, logger)
	handler.ServeHTTP(rec, req)

	// Check status code
	assert.Equal(t, http.StatusCreated, rec.Code, "Expected status code 201 for valid reception creation")

	// Check the response body
	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response body")
	assert.NotEmpty(t, response["id"], "Expected non-empty reception ID in response")
	assert.Equal(t, pvzID, response["pvzId"], "Expected correct pvzId in response")
}

func TestCreateReception_Errors(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialisation of the test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Сreate tables
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS pvz (
            id TEXT PRIMARY KEY,
            city TEXT NOT NULL,
            registration_date TEXT NOT NULL
        );
        CREATE TABLE IF NOT EXISTS receptions (
            id TEXT PRIMARY KEY,
            date_time TEXT NOT NULL,
            pvz_id TEXT NOT NULL,
            status TEXT NOT NULL
        );
    `)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Test 1: No pvzId
	payload := map[string]string{}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/receptions", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	ctx := req.Context()
	ctx = context.WithValue(ctx, "userRole", "employee")
	req = req.WithContext(ctx)

	handler := CreateReception(db, logger)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for missing pvzId")

	// Test 2: Non-existent PVZ
	payload = map[string]string{"pvzId": "non-existent-pvz"}
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest("POST", "/receptions", bytes.NewReader(body))
	rec = httptest.NewRecorder()

	ctx = req.Context()
	ctx = context.WithValue(ctx, "userRole", "employee")
	req = req.WithContext(ctx)

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for non-existent PVZ")
}

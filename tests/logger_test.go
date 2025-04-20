package tests

import (
	"PVZ/database"
	"PVZ/handlers"
	"bytes"
	"context"
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogging(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialisation of the test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Test 1: Logging error on invalid JSON
	var logBuffer bytes.Buffer
	logger = log.New(&logBuffer, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	req := httptest.NewRequest("POST", "/pvz", bytes.NewReader([]byte("invalid-json")))
	rec := httptest.NewRecorder()

	ctx := req.Context()
	ctx = context.WithValue(ctx, "userRole", "moderator")
	req = req.WithContext(ctx)

	handler := handlers.CreatePVZ(db, logger)
	handler.ServeHTTP(rec, req)

	assert.Contains(t, logBuffer.String(), "Invalid request body", "Expected log message for invalid request body")
}

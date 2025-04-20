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

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestDummyLogin(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Test 1: Successful authorization with a valid role
	payload := map[string]string{"role": "employee"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler := DummyLogin([]byte("my_secret_key"), logger)
	handler.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200 for valid role")

	// Check the response body
	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response body")
	assert.NotEmpty(t, response["token"], "Expected non-empty token in response")

	// Check the token format
	token := response["token"]
	_, err = jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("my_secret_key"), nil
	})
	assert.NoError(t, err, "Failed to parse JWT token")

	// Test 2: Invalid user role
	payload = map[string]string{"role": "admin"} // Invalid role
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest("POST", "/dummyLogin", bytes.NewReader(body))
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for invalid role")

	// Test 3: Missing role in the request
	payload = map[string]string{} // Role not specified
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest("POST", "/dummyLogin", bytes.NewReader(body))
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for missing role")

	// Test 4: Invalid request format
	req = httptest.NewRequest("POST", "/dummyLogin", bytes.NewReader([]byte("invalid-json")))
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for invalid JSON")

	// Test 5: Logging errors
	var logBuffer bytes.Buffer
	logger = log.New(&logBuffer, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	handler = DummyLogin([]byte("my_secret_key"), logger)
	req = httptest.NewRequest("POST", "/dummyLogin", bytes.NewReader([]byte("invalid-json")))
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check that the logger recorded the error
	assert.Contains(t, logBuffer.String(), "Invalid request body", "Expected log message for invalid request body")
}

func TestRegister(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialize test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Create users table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY,
            email TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL,
            role TEXT NOT NULL
        );
    `)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Test 1: Successful registration
	payload := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
		"role":     "employee",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler := Register(db, logger)
	handler.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusCreated, rec.Code, "Expected status code 201 for valid registration")

	// Check the response body
	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response body")
	assert.NotEmpty(t, response["id"], "Expected non-empty user ID in response")

	// Check that the password is stored in hashed form
	var storedPassword string
	err = db.QueryRow("SELECT password FROM users WHERE email = ?", payload["email"]).Scan(&storedPassword)
	assert.NoError(t, err, "Failed to query database")
	assert.NotEqual(t, payload["password"], storedPassword, "Expected password to be hashed")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(payload["password"])), "Failed to verify hashed password")

	// Test 2: Registration with an invalid role
	payload = map[string]string{
		"email":    "invalid-role@example.com",
		"password": "password123",
		"role":     "admin", // Invalid role
	}
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for invalid role")

	// Test 3: Registration with an invalid email
	payload = map[string]string{
		"email":    "invalid-email",
		"password": "password123",
		"role":     "employee",
	}
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for invalid email")

	// Test 4: Registration with missing fields
	payload = map[string]string{
		"email": "missing-fields@example.com",
		// Password and role are missing
	}
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for missing fields")

	// Test 5: Duplicate registration with the same email
	payload = map[string]string{
		"email":    "test@example.com", // Email already in use
		"password": "password123",
		"role":     "employee",
	}
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for duplicate email")

	// Test 6: Invalid request format
	req = httptest.NewRequest("POST", "/register", bytes.NewReader([]byte("invalid-json")))
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for invalid JSON")

	// Test 7: Logging errors
	var logBuffer bytes.Buffer
	logger = log.New(&logBuffer, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	handler = Register(db, logger)
	req = httptest.NewRequest("POST", "/register", bytes.NewReader([]byte("invalid-json")))
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check that the logger recorded the error
	assert.Contains(t, logBuffer.String(), "Invalid request body", "Expected log message for invalid request body")
}
func TestLogin(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialize test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Create users table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY,
            email TEXT NOT NULL,
            password TEXT NOT NULL,
            role TEXT NOT NULL
        );
    `)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Hash the password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// Add a test user
	_, err = db.Exec(`
        INSERT INTO users (id, email, password, role)
        VALUES (?, ?, ?, ?)
    `, "user-1", "test@example.com", hashedPassword, "employee")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Test 1: Successful login
	payload := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	jwtKey := []byte("my_secret_key") // Add secret key
	handler := Login(db, jwtKey, logger)
	handler.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200 for valid login")

	// Check the response body
	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response body")
	assert.NotEmpty(t, response["token"], "Expected non-empty token in response")

	// Test 2: Invalid password
	payload = map[string]string{
		"email":    "test@example.com",
		"password": "wrong-password",
	}
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusUnauthorized, rec.Code, "Expected status code 401 for invalid password")

	// Test 3: Invalid email
	payload = map[string]string{
		"email":    "wrong@example.com",
		"password": "password123",
	}
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusUnauthorized, rec.Code, "Expected status code 401 for invalid email")

	// Test 5: Invalid request format
	req = httptest.NewRequest("POST", "/login", bytes.NewReader([]byte("invalid-json")))
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected status code 400 for invalid JSON")
}

func TestAuthorizationErrors(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialize test database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Test 1: Accessing `/pvz` without a role
	req := httptest.NewRequest("GET", "/pvz", nil)
	rec := httptest.NewRecorder()

	handler := GetPVZList(db, logger)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code, "Expected status code 403 for missing role")

	// Test 2: Accessing `/pvz` with an invalid role
	ctx := req.Context()
	ctx = context.WithValue(ctx, "userRole", "admin") // Invalid role
	req = req.WithContext(ctx)

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code, "Expected status code 403 for invalid role")
}

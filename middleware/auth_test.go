package middleware

import (
	"PVZ/controllers"
	"PVZ/database"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var jwtKey = []byte("my_secret_key")

func TestAuthMiddleware(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Create a test token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &UserClaims{
		Role: "employee",
	})
	tokenString, _ := token.SignedString(jwtKey)

	// Create an in-memory database
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Create a router
	router := mux.NewRouter()

	// Register the /pvz endpoint with middleware
	protected := router.PathPrefix("/").Subrouter()
	protected.Use(AuthMiddleware(jwtKey, logger, "employee"))
	protected.HandleFunc("/pvz", controllers.GetPVZList(db, logger)).Methods("GET")

	// Test 1: Valid token and role
	req := httptest.NewRequest("GET", "/pvz", nil)
	rec := httptest.NewRecorder()

	// Add a valid token to the header
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Execute the request
	protected.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200 for valid token and role")

	// Test 2: Request without a token
	req = httptest.NewRequest("GET", "/pvz", nil)
	rec = httptest.NewRecorder()

	// Execute the request without the Authorization header
	protected.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusUnauthorized, rec.Code, "Expected status code 401 for missing token")

	// Test 3: Request with an invalid token
	req = httptest.NewRequest("GET", "/pvz", nil)
	rec = httptest.NewRecorder()

	// Add an invalid token to the header
	req.Header.Set("Authorization", "Bearer invalid-token")

	// Execute the request
	protected.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusUnauthorized, rec.Code, "Expected status code 401 for invalid token")

	// Test 4: Request with insufficient permissions
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, &UserClaims{
		Role: "moderator",
	})
	tokenString, _ = token.SignedString(jwtKey)

	req = httptest.NewRequest("GET", "/pvz", nil)
	rec = httptest.NewRecorder()

	// Add a token with the moderator role
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Execute the request
	protected.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusForbidden, rec.Code, "Expected status code 403 for insufficient permissions")
}

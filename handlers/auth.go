package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// UserClaims represents the data stored in the JWT token
type UserClaims struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.StandardClaims
}

func DummyLogin(jwtKey []byte, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Role string `json:"role"`
		}

		logger.Println("Received dummy login request")

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			logger.Printf("Invalid request body: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		validRoles := map[string]bool{"employee": true, "moderator": true}
		if !validRoles[req.Role] {
			logger.Printf("Invalid role requested: %s", req.Role)
			http.Error(w, "Invalid role", http.StatusBadRequest)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"role": req.Role,
			"exp":  time.Now().Add(24 * time.Hour).Unix(),
		})

		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			logger.Printf("Failed to generate token: %v", err)
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		logger.Printf("Token generated successfully for role: %s", req.Role)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	}
}

// Register registers a new user, hashes the password and saves the user to the database
func Register(db *sql.DB, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}

		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			logger.Printf("Invalid request body: %v", err)
			http.Error(w, "{\"error\": \"Invalid request body\"}", http.StatusBadRequest)
			return
		}

		if !isValidEmail(user.Email) {
			logger.Printf("Invalid email format: %s", user.Email)
			http.Error(w, "{\"error\": \"Invalid email format\"}", http.StatusBadRequest)
			return
		}

		if len(user.Password) < 8 {
			logger.Printf("Password too short for email: %s", user.Email)
			http.Error(w, "{\"error\": \"Password must be at least 8 characters long\"}", http.StatusBadRequest)
			return
		}

		if user.Role != "employee" && user.Role != "moderator" {
			logger.Printf("Invalid role: %s", user.Role)
			http.Error(w, "{\"error\": \"Invalid role\"}", http.StatusBadRequest)
			return
		}

		var existingEmail string
		err = db.QueryRow("SELECT email FROM users WHERE email = ?", user.Email).Scan(&existingEmail)
		if err == nil {
			logger.Printf("User with email %s already exists", user.Email)
			http.Error(w, "{\"error\": \"User with this email already exists\"}", http.StatusBadRequest)
			return
		} else if err != sql.ErrNoRows {
			logger.Printf("Database error: %v", err)
			http.Error(w, "{\"error\": \"Database error\"}", http.StatusInternalServerError)
			return
		}

		userID := GenerateUUID()
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			logger.Printf("Failed to hash password: %v", err)
			http.Error(w, "{\"error\": \"Failed to hash password\"}", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("INSERT INTO users (id, email, password, role) VALUES (?, ?, ?, ?)",
			userID, user.Email, hashedPassword, user.Role)
		if err != nil {
			logger.Printf("Failed to create user: %v", err)
			http.Error(w, "{\"error\": \"Failed to create user\"}", http.StatusInternalServerError)
			return
		}

		logger.Printf("User registered successfully: %s", user.Email)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"id":    userID,
			"email": user.Email,
			"role":  user.Role,
		})
	}
}

// isValidEmail checks email
func isValidEmail(email string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`).MatchString(email)
}

// Login authorises the user and returns the token
func Login(db *sql.DB, jwtKey []byte, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var credentials struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		logger.Println("Received login request")
		err := json.NewDecoder(r.Body).Decode(&credentials)
		if err != nil {
			logger.Printf("Invalid request body: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		logger.Printf("Processing login for email: %s", credentials.Email)

		var storedPassword, role string
		err = db.QueryRow("SELECT password, role FROM users WHERE email = ?", credentials.Email).
			Scan(&storedPassword, &role)
		if err == sql.ErrNoRows {
			logger.Printf("Invalid email or password for email: %s", credentials.Email)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		} else if err != nil {
			logger.Printf("Database error during login: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(credentials.Password))
		if err != nil {
			logger.Printf("Invalid password for email: %s", credentials.Email)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"email": credentials.Email,
			"role":  role,
			"exp":   time.Now().Add(time.Hour * 24).Unix(),
		})
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			logger.Printf("Failed to generate token: %v", err)
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		logger.Printf("User logged in successfully: %s", credentials.Email)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	}
}

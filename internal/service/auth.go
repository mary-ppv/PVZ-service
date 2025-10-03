package service

import (
	"PVZ/internal/models"
	"PVZ/pkg/logger"
	"PVZ/pkg/uuid"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo   UserRepository
	jwtKey []byte
}

func NewUserService(repo UserRepository, jwtKey []byte) *UserService {
	return &UserService{repo: repo, jwtKey: jwtKey}
}

func (s *UserService) Register(email, password, role string) (*models.User, error) {
	r := models.Role(role)
	if r != models.RoleEmployee && r != models.RoleModerator {
		logger.Log.Printf("Invalid role: %s for email %s", role, email)
		return nil, errors.New("invalid role")
	}

	if len(password) < 8 {
		logger.Log.Printf("Password too short for email %s", email)
		return nil, errors.New("password too short")
	}

	existing, err := s.repo.GetByEmail(email)
	if err != nil {
		logger.Log.Printf("Failed to check existing user for email %s: %v", email, err)
		return nil, err
	}
	if existing != nil {
		logger.Log.Printf("User already exists: %s", email)
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Printf("Failed to hash password for email %s: %v", email, err)
		return nil, err
	}

	id, err := uuid.GenerateUUID7()
	if err != nil {
		logger.Log.Printf("Failed to generate UUID for email %s: %v", email, err)
		return nil, errors.New("cannot generate uuid")
	}

	user := &models.User{
		ID:        id,
		Email:     email,
		Password:  string(hashedPassword),
		Role:      r,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateUser(user); err != nil {
		logger.Log.Printf("Failed to create user %s: %v", email, err)
		return nil, err
	}

	logger.Log.Printf("User registered: %s with role %s", email, role)
	return user, nil
}

func (s *UserService) Login(email, password string) (string, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		logger.Log.Printf("Failed to get user by email %s: %v", email, err)
		return "", err
	}

	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		logger.Log.Printf("Invalid login attempt for email %s", email)
		return "", errors.New("invalid email or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	signedToken, err := token.SignedString(s.jwtKey)
	if err != nil {
		logger.Log.Printf("Failed to sign JWT for email %s: %v", email, err)
		return "", err
	}

	logger.Log.Printf("User logged in: %s", email)
	return signedToken, nil
}

func (s *UserService) DummyLogin(role string) (string, error) {
	r := models.Role(role)
	if r != models.RoleEmployee && r != models.RoleModerator {
		logger.Log.Printf("Invalid dummy login role: %s", role)
		return "", errors.New("invalid role")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": r,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})

	signedToken, err := token.SignedString(s.jwtKey)
	if err != nil {
		logger.Log.Printf("Failed to sign dummy JWT for role %s: %v", role, err)
		return "", err
	}

	logger.Log.Printf("Dummy login for role: %s", role)
	return signedToken, nil
}

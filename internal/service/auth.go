package service

import (
	"PVZ/internal/models"
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
		return nil, errors.New("invalid role")
	}

	if len(password) < 8 {
		return nil, errors.New("password too short")
	}

	existing, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:        GenerateUUID(),
		Email:     email,
		Password:  string(hashedPassword),
		Role:      r,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(email, password string) (string, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return "", err
	}

	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return "", errors.New("invalid email or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(s.jwtKey)
}

func (s *UserService) DummyLogin(role string) (string, error) {
	r := models.Role(role)
	if r != models.RoleEmployee && r != models.RoleModerator {
		return "", errors.New("invalid role")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": r,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(s.jwtKey)
}

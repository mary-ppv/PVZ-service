package service

import (
	"PVZ/internal/constants"
	"PVZ/models"
	"PVZ/pkg/auth"
	"PVZ/pkg/uuid"
	"context"
	"errors"
	"log/slog"
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

func (s *UserService) Register(ctx context.Context, email, password, role string) (*models.User, error) {
	if role != constants.RoleEmployee && role != constants.RoleModerator {
		return nil, errors.New("invalid role")
	}

	if len(password) < 8 {
		return nil, errors.New("password too short")
	}

	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if existing != nil {
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	id, err := uuid.GenerateUUID7()
	if err != nil {
		return nil, errors.New("cannot generate uuid")
	}

	user := &models.User{
		ID:        id,
		Email:     email,
		Password:  string(hashedPassword),
		Role:      role,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("user not found")
	}

	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return "", errors.New("invalid email or password")
	}

	claims := &auth.UserClaims{
		Role: string(user.Role),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtKey)
	if err != nil {
		slog.Warn("failed to sign JWT")
		return "", err
	}

	return signedToken, nil
}

func (s *UserService) DummyLogin(role string) (string, error) {
	if role != constants.RoleModerator && role != constants.RoleEmployee {
		return "", errors.New("invalid role")
	}

	claims := &auth.UserClaims{
		Role: role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.jwtKey)
	if err != nil {
		slog.Warn("Failed to sign dummy JWT for role %s: %v", role, err)
		return "", err
	}

	return signedToken, nil
}

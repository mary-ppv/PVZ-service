package repository

import (
	"PVZ/internal/models"
	"PVZ/pkg/database"
	"PVZ/pkg/logger"
	"database/sql"
	"errors"
)

type UserRepo struct {
	db *database.DB
}

func NewUserRepo(db *database.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateUser(user *models.User) error {
	_, err := r.db.Exec(
		"INSERT INTO users (id, email, password, role, created_at) VALUES ($1, $2, $3, $4, $5)",
		user.ID, user.Email, user.Password, user.Role, user.CreatedAt,
	)

	if err != nil {
		logger.Log.Printf("Failed to create user %s: %v", user.Email, err)
		return err
	}

	logger.Log.Printf("User %s created with ID %s", user.Email, user.ID)
	return nil
}

func (r *UserRepo) GetByEmail(email string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow("SELECT id, email, password, role, created_at FROM users WHERE email = $1", email).
		Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Log.Printf("User %s not found", email)
		return nil, nil
	}
	if err != nil {
		logger.Log.Printf("Failed to get user %s: %v", email, err)
		return nil, err
	}
	logger.Log.Printf("User %s retrieved", email)
	return &u, nil
}

package repository

import (
	"PVZ/internal/models"
	"PVZ/pkg/database"
	"database/sql"
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
	return err
}

func (r *UserRepo) GetByEmail(email string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow("SELECT id, email, password, role, created_at FROM users WHERE email = $1", email).
		Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

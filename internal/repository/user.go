package repository

import (
	"PVZ/internal/models"
	"database/sql"
)

type UserRepo struct {
	db *sql.DB
}
	
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateUser(user *models.User) error {
	_, err := r.db.Exec(
		"INSERT INTO users (id, email, password, role, created_at) VALUES (?, ?, ?, ?, ?)",
		user.ID, user.Email, user.Password, user.Role, user.CreatedAt,
	)
	return err
}

func (r *UserRepo) GetByEmail(email string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow("SELECT id, email, password, role, created_at FROM users WHERE email = ?", email).
		Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

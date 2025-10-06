package repository

import (
	"PVZ/models"
	"PVZ/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
)

type UserRepo struct {
	db boil.ContextExecutor
}

func NewUserRepo(db boil.ContextExecutor) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateUser(ctx context.Context, user *models.User) error {
	err := user.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		logger.Log.Printf("Failed to create user %s: %v", user.Email, err)
		return fmt.Errorf("failed to create user: %w", err)
	}

	logger.Log.Printf("User %s created with ID %s", user.Email, user.ID)
	return nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := models.Users(
		models.UserWhere.Email.EQ(email),
	).One(ctx, r.db)

	if errors.Is(err, sql.ErrNoRows) {
		logger.Log.Printf("User %s not found", email)
		return nil, nil
	}
	if err != nil {
		logger.Log.Printf("Failed to get user %s: %v", email, err)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	logger.Log.Printf("User %s retrieved", email)
	return user, nil
}

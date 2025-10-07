package repository

import (
	"PVZ/models"
	"context"
	"database/sql"
	"errors"
	"log/slog"

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
		slog.Error("Failed to create user %s: %v", user.Email, err)
		return err
	}

	return nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := models.Users(
		models.UserWhere.Email.EQ(email),
	).One(ctx, r.db)

	if errors.Is(err, sql.ErrNoRows) {
		slog.Info("User %s not found", email)
		return nil, nil
	}
	if err != nil {
		slog.Info("Failed to get user %s: %v", email, err)
		return nil, err
	}

	return user, nil
}

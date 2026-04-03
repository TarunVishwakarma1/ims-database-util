package service

import (
	"context"
	"errors"
	"fmt"
	"ims-database-util/internal/repository"
	"log/slog"
	"time"
)

type UserService interface {
	GetUserByID(ctx context.Context, id string) (*repository.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*repository.User, error) {
	if id == "" {
		return nil, errors.New("user id cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		slog.Error("GetUserByID failed", "id", id, "error", err)
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	return user, nil
}

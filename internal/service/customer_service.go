package service

import (
	"context"
	"fmt"
	"ims-database-util/internal/repository"
	"log/slog"
	"time"
)

type CustomerService interface {
	StreamCustomers(ctx context.Context, handler func([]repository.Customer) error) error
}

type customerService struct {
	repo repository.CustomerRepository
}

func NewCustomerService(repo repository.CustomerRepository) CustomerService {
	return &customerService{repo: repo}
}

func (s *customerService) StreamCustomers(
	ctx context.Context,
	handler func([]repository.Customer) error,
) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	batchSize := 10

	err := s.repo.StreamCustomers(ctx, batchSize, func(batch []repository.Customer) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		return handler(batch)
	})

	if err != nil {
		slog.Error("StreamCustomers failed", "error", err)
		return fmt.Errorf("streaming failed: %w", err)
	}

	return nil
}

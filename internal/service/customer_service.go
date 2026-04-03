package service

import (
	"context"
	"fmt"
	"ims-database-util/internal/repository"
)

type CustomerService interface {
	StreamCustomers(ctx context.Context, handler func([]repository.Customer) error) error
}

type customerService struct {
	repo repository.CustomerRepository
}

// NewCustomerService returns a CustomerService backed by the provided CustomerRepository.
func NewCustomerService(repo repository.CustomerRepository) CustomerService {
	return &customerService{repo: repo}
}

func (s *customerService) StreamCustomers(
	ctx context.Context,
	handler func([]repository.Customer) error,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	batchSize := 100

	err := s.repo.StreamCustomers(ctx, batchSize, func(batch []repository.Customer) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		return handler(batch)
	})

	if err != nil {
		return fmt.Errorf("streaming failed: %w", err)
	}

	return nil
}

package service

import (
	"context"
	"errors"
	"fmt"
	"ims-database-util/internal/repository"
	"log/slog"
	"time"
)

type ProductService interface {
	StreamProducts(ctx context.Context, handler func([]repository.Product) error) error

	GetProductByID(ctx context.Context, id string) (repository.Product, error)
	GetProductsByUserID(ctx context.Context, userID string) ([]repository.Product, error)

	CreateProduct(ctx context.Context, product repository.Product) (repository.Product, error)
	UpdateProduct(ctx context.Context, id string, product repository.Product) (repository.Product, error)
	DeleteProduct(ctx context.Context, id string) (string, error)
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) StreamProducts(
	ctx context.Context,
	handler func([]repository.Product) error,
) error {

	// ⏱️ Add timeout (important for safety)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	batchSize := 100

	err := s.repo.StreamProducts(ctx, batchSize, func(batch []repository.Product) error {
		// 🧠 Backpressure / cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		return handler(batch)
	})

	if err != nil {
		slog.Error("StreamProducts failed", "error", err)
		return fmt.Errorf("streaming failed: %w", err)
	}

	return nil
}

func (s *productService) GetProductByID(ctx context.Context, id string) (repository.Product, error) {
	if id == "" {
		return repository.Product{}, errors.New("product id cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	product, err := s.repo.GetProductById(ctx, id)
	if err != nil {
		slog.Error("GetProductByID failed", "error", err)
		return repository.Product{}, fmt.Errorf("failed to fetch product: %w", err)
	}

	return product, nil
}

func (s *productService) GetProductsByUserID(ctx context.Context, userID string) ([]repository.Product, error) {
	if userID == "" {
		return nil, errors.New("user id cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.repo.GetProductsByUserId(ctx, userID)
}

func (s *productService) CreateProduct(ctx context.Context, p repository.Product) (repository.Product, error) {
	if p.Name == "" || p.Sku == "" {
		return repository.Product{}, errors.New("invalid product data")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.repo.CreateProduct(ctx, p)
}

func (s *productService) UpdateProduct(ctx context.Context, id string, p repository.Product) (repository.Product, error) {
	if id == "" {
		return repository.Product{}, errors.New("id required")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.repo.UpdateProductById(ctx, id, p)
}

func (s *productService) DeleteProduct(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", errors.New("id required")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.repo.DeleteProductById(ctx, id)
}

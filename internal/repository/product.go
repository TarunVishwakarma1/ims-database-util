package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Product struct {
	Id          string    `json:"id"`
	Sku         string    `json:"sku"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	Status      string    `json:"status"`
	LastUpdated time.Time `json:"lastUpdated"`
	UpdatedBy   string    `json:"updatedBy"`
	AddedBy     string    `json:"addedBy"`
	UserId      string    `json:"userId"`
	AddedAt     time.Time `json:"addedAt"`
}

type ProductRepository interface {
	StreamProducts(ctx context.Context, batchSize int, handler func([]Product) error) error
	GetProductById(ctx context.Context, productId string) (Product, error)
	GetProductsByUserId(ctx context.Context, userId string) ([]Product, error)
	UpdateProductById(ctx context.Context, productId string, product Product) (Product, error)
	DeleteProductById(ctx context.Context, productId string) (string, error)
	CreateProduct(ctx context.Context, product Product) (Product, error)
}

type postgresProductRepo struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) ProductRepository {
	return &postgresProductRepo{pool: pool}
}

func (r *postgresProductRepo) StreamProducts(
	ctx context.Context,
	batchSize int,
	handler func([]Product) error,
) error {
	var lastAddedAt time.Time
	var lastId string

	for {
		batch, err := r.fetchProductBatch(ctx, batchSize, lastAddedAt, lastId)
		if err != nil {
			return err
		}

		if len(batch) == 0 {
			break
		}

		if err := handler(batch); err != nil {
			return err
		}

		lastProduct := batch[len(batch)-1]
		lastAddedAt = lastProduct.AddedAt
		lastId = lastProduct.Id
	}

	return nil
}

func (r *postgresProductRepo) fetchProductBatch(
	ctx context.Context, 
	batchSize int, 
	lastAddedAt time.Time, 
	lastId string,
) ([]Product, error) {
	query := `
		SELECT id, sku, name, category, price, stock, status,
		       last_updated, updated_by, added_by, user_id, added_at
		FROM products`
		
	var args []any

	if lastId != "" {
		query += ` WHERE (added_at, id) > ($1, $2)`
		args = append(args, lastAddedAt, lastId)
	}

	query += fmt.Sprintf(` ORDER BY added_at ASC, id ASC LIMIT $%d`, len(args)+1)
	args = append(args, batchSize)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	batch := make([]Product, 0, batchSize)

	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.Id,
			&p.Sku,
			&p.Name,
			&p.Category,
			&p.Price,
			&p.Stock,
			&p.Status,
			&p.LastUpdated,
			&p.UpdatedBy,
			&p.AddedBy,
			&p.UserId,
			&p.AddedAt,
		); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		batch = append(batch, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return batch, nil
}

func (r *postgresProductRepo) GetProductById(ctx context.Context, productId string) (Product, error) {
	query := `SELECT id, sku, name, category, price, stock, status, lastupdated, updated_by, added_by, user_id, added_at 
	FROM PRODUCTS WHERE id = $1`

	var p Product
	err := r.pool.QueryRow(ctx, query, productId).Scan(
		&p.Id,
		&p.Sku,
		&p.Name,
		&p.Category,
		&p.Price,
		&p.Stock,
		&p.Status,
		&p.LastUpdated,
		&p.UpdatedBy,
		&p.AddedBy,
		&p.UserId,
		&p.AddedAt,
	)

	if err != nil {
		return Product{}, fmt.Errorf("error getting product by id: %w", err)
	}

	return p, nil
}

func (r *postgresProductRepo) GetProductsByUserId(ctx context.Context, userId string) ([]Product, error) {
	// Assuming there's a user_id column in the PRODUCTS table based on this method name
	query := `SELECT id, sku, name, category, price, stock, status, last_updated, updated_by, added_by, user_id, added_at 
	FROM PRODUCTS WHERE user_id = $1`

	var products []Product
	rows, err := r.pool.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("error querying products by user id: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.Id,
			&p.Sku,
			&p.Name,
			&p.Category,
			&p.Price,
			&p.Stock,
			&p.Status,
			&p.LastUpdated,
			&p.UpdatedBy,
			&p.AddedBy,
			&p.UserId,
			&p.AddedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning product row: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return products, nil
}

// Ensure the updated data gets passed in by changing the signature appropriately.
func (r *postgresProductRepo) UpdateProductById(ctx context.Context, id string, product Product) (Product, error) {
	// Updating with current_timestamp. Adjust query if lastupdated is purely application logic.
	query := `UPDATE PRODUCTS 
	SET sku = $1, name = $2, category = $3, price = $4, stock = $5, status = $6, last_updated = current_timestamp, updated_by = $7
	WHERE id = $8 
	RETURNING id, sku, name, category, price, stock, status, lastupdated, updated_by, added_by, user_id, added_at`

	var p Product
	err := r.pool.QueryRow(ctx, query, product.Sku, product.Name, product.Category, product.Price, product.Stock, product.Status, product.UpdatedBy, id).Scan(
		&p.Id,
		&p.Sku,
		&p.Name,
		&p.Category,
		&p.Price,
		&p.Stock,
		&p.Status,
		&p.LastUpdated,
		&p.UpdatedBy,
		&p.AddedBy,
		&p.UserId,
		&p.AddedAt,
	)

	if err != nil {
		return Product{}, fmt.Errorf("error updating product: %w", err)
	}

	return p, nil
}

func (r *postgresProductRepo) DeleteProductById(ctx context.Context, id string) (string, error) {
	query := `DELETE FROM PRODUCTS WHERE id = $1`

	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return "", fmt.Errorf("error deleting product: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return "", fmt.Errorf("product with id %s not found", id)
	}

	return id, nil
}

func (r *postgresProductRepo) CreateProduct(ctx context.Context, product Product) (Product, error) {
	query := `INSERT INTO PRODUCTS (id, sku, name, category, price, stock, status, last_updated, updated_by, added_by, user_id, added_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) 
	RETURNING id, sku, name, category, price, stock, status, lastupdated, updated_by, added_by, user_id, added_at`

	var p Product
	err := r.pool.QueryRow(ctx, query, product.Id, product.Sku, product.Name, product.Category, product.Price, product.Stock, product.Status, product.LastUpdated, product.UpdatedBy, product.AddedBy, product.UserId, product.AddedAt).Scan(
		&p.Id,
		&p.Sku,
		&p.Name,
		&p.Category,
		&p.Price,
		&p.Stock,
		&p.Status,
		&p.LastUpdated,
		&p.UpdatedBy,
		&p.AddedBy,
		&p.UserId,
		&p.AddedAt,
	)

	if err != nil {
		return Product{}, fmt.Errorf("error creating product: %w", err)
	}

	return p, nil
}

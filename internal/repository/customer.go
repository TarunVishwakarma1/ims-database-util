package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Customer struct {
	Id        string    `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	Status    string    `json:"status"` // E.g., Active, Inactive, Banned
	UserId    string    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CustomerRepository interface {
	StreamCustomers(ctx context.Context, batchSize int, handler func([]Customer) error) error
	GetById(ctx context.Context, id string) (Customer, error)
	GetByUserId(ctx context.Context, userId string) ([]Customer, error)
	Add(ctx context.Context, customer Customer) (Customer, error)
	Update(ctx context.Context, id string, customer Customer) (Customer, error)
	Delete(ctx context.Context, id string) (string, error)
}

type postgresCustomerRepo struct {
	pool *pgxpool.Pool
}

func NewCustomerRepository(pool *pgxpool.Pool) CustomerRepository {
	return &postgresCustomerRepo{pool: pool}
}

func (r *postgresCustomerRepo) StreamCustomers(
	ctx context.Context,
	batchSize int,
	handler func([]Customer) error,
) error {
	var lastCreatedAt time.Time
	var lastId string

	for {
		batch, err := r.fetchCustomerBatch(ctx, batchSize, lastCreatedAt, lastId)
		if err != nil {
			return err
		}

		if len(batch) == 0 {
			break
		}

		if err := handler(batch); err != nil {
			return err
		}

		lastCustomer := batch[len(batch)-1]
		lastCreatedAt = lastCustomer.CreatedAt
		lastId = lastCustomer.Id
	}

	return nil
}

func (r *postgresCustomerRepo) fetchCustomerBatch(
	ctx context.Context,
	batchSize int,
	lastCreatedAt time.Time,
	lastId string,
) ([]Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, address, status, user_id, created_at, updated_at
		FROM customers`

	var args []any

	if lastId != "" {
		query += ` WHERE (created_at, id) > ($1, $2)`
		args = append(args, lastCreatedAt, lastId)
	}

	query += fmt.Sprintf(` ORDER BY created_at ASC, id ASC LIMIT $%d`, len(args)+1)
	args = append(args, batchSize)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	batch := make([]Customer, 0, batchSize)

	for rows.Next() {
		var c Customer
		if err := rows.Scan(
			&c.Id,
			&c.FirstName,
			&c.LastName,
			&c.Email,
			&c.Phone,
			&c.Address,
			&c.Status,
			&c.UserId,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		batch = append(batch, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return batch, nil
}

func (r *postgresCustomerRepo) GetById(ctx context.Context, id string) (Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, address, status, user_id, created_at, updated_at
		FROM customers
		WHERE id = $1`

	var c Customer
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.Id,
		&c.FirstName,
		&c.LastName,
		&c.Email,
		&c.Phone,
		&c.Address,
		&c.Status,
		&c.UserId,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		return Customer{}, fmt.Errorf("error getting customer by id: %w", err)
	}

	return c, nil
}

func (r *postgresCustomerRepo) GetByUserId(ctx context.Context, userId string) ([]Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, address, status, user_id, created_at, updated_at
		FROM customers
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("error querying customers by user id: %w", err)
	}
	defer rows.Close()

	var customers []Customer
	for rows.Next() {
		var c Customer
		if err := rows.Scan(
			&c.Id,
			&c.FirstName,
			&c.LastName,
			&c.Email,
			&c.Phone,
			&c.Address,
			&c.Status,
			&c.UserId,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		customers = append(customers, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error iteration: %w", err)
	}

	return customers, nil
}

func (r *postgresCustomerRepo) Add(ctx context.Context, customer Customer) (Customer, error) {
	query := `
		INSERT INTO customers (id, first_name, last_name, email, phone, address, status, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, first_name, last_name, email, phone, address, status, user_id, created_at, updated_at`

	var c Customer
	err := r.pool.QueryRow(ctx, query,
		customer.Id,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.Phone,
		customer.Address,
		customer.Status,
		customer.UserId,
		customer.CreatedAt,
		customer.UpdatedAt,
	).Scan(
		&c.Id,
		&c.FirstName,
		&c.LastName,
		&c.Email,
		&c.Phone,
		&c.Address,
		&c.Status,
		&c.UserId,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		return Customer{}, fmt.Errorf("error creating customer: %w", err)
	}

	return c, nil
}

func (r *postgresCustomerRepo) Update(ctx context.Context, id string, customer Customer) (Customer, error) {
	query := `
		UPDATE customers
		SET first_name = $1, last_name = $2, email = $3, phone = $4, address = $5, status = $6, updated_at = current_timestamp
		WHERE id = $7
		RETURNING id, first_name, last_name, email, phone, address, status, user_id, created_at, updated_at`

	var c Customer
	err := r.pool.QueryRow(ctx, query,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.Phone,
		customer.Address,
		customer.Status,
		id,
	).Scan(
		&c.Id,
		&c.FirstName,
		&c.LastName,
		&c.Email,
		&c.Phone,
		&c.Address,
		&c.Status,
		&c.UserId,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		return Customer{}, fmt.Errorf("error updating customer: %w", err)
	}

	return c, nil
}

func (r *postgresCustomerRepo) Delete(ctx context.Context, id string) (string, error) {
	query := `DELETE FROM customers WHERE id = $1`

	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return "", fmt.Errorf("error deleting customer: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return "", fmt.Errorf("customer with id %s not found", id)
	}

	return id, nil
}

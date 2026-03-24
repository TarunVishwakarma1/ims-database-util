package storage

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pgInstance *pgxpool.Pool
	pgOnce     sync.Once
)

// InitPostgres initializes and stores a singleton pgxpool.Pool using the given
// connection string, verifies connectivity, and returns the shared pool.
// Initialization is performed exactly once; subsequent calls return the same
// pool and the error (if any) from the initial attempt. The function configures
// the connection pool with sensible defaults and pings the database to ensure
// reachability before storing the pool.
func InitPostgres(context context.Context, connString string) (*pgxpool.Pool, error) {
	var err error
	pgOnce.Do(func() {
		config, parseErr := pgxpool.ParseConfig(connString)
		if parseErr != nil {
			err = fmt.Errorf("failed to parse postgres config: %w", parseErr)
			return
		}

		config.MaxConns = 25                      // prevent overwhelming the db
		config.MinConns = 5                       // keep warm connections ready
		config.MaxConnLifetime = time.Hour        // Recycle connection to prevent memory leakage
		config.MaxConnIdleTime = 30 * time.Minute // close idle connections

		pool, poolErr := pgxpool.NewWithConfig(context, config)
		if poolErr != nil {
			err = fmt.Errorf("failed to create postgres pool: %w", poolErr)
			return
		}

		if pingErr := pool.Ping(context); pingErr != nil {
			err = fmt.Errorf("postgres is not reachable: %w", pingErr)
			return
		}

		slog.Info("✅ Successfully connected to PostgreSQL")
		pgInstance = pool
	})
	return pgInstance, err
}

// GetPostgres provides access to the package-level PostgreSQL connection pool.
// It returns the initialized *pgxpool.Pool and will panic if the pool has not been initialized.
func GetPostgres() *pgxpool.Pool {
	if pgInstance == nil {
		panic("PostgreSQL pool accessed before initialization")
	}
	return pgInstance
}

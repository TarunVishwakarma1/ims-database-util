package app

import (
	"ims-database-util/internal/config"
	"ims-database-util/internal/repository"
	"ims-database-util/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// App is the central dependency container. All repositories are wired into
// services here. To integrate a new domain:
//  1. Add a new service field
//  2. Wire the repo → service in New()
//
// Handlers and gRPC servers receive App (or individual services) — they never
// touch repositories directly.
type App struct {
	Config          *config.Config
	UserService     service.UserService
	ProductService  service.ProductService
	CustomerService service.CustomerService
	// Future domains:
	// OrderService    service.OrderService
	// InventoryService service.InventoryService
}

// New builds the full application dependency graph from infrastructure handles.
func New(cfg *config.Config, pgPool *pgxpool.Pool, rdb *redis.Client) *App {
	// Repositories
	userRepo := repository.NewUserRepository(pgPool)
	productRepo := repository.NewProductRepository(pgPool)
	cusomerRepo := repository.NewCustomerRepository(pgPool)
	// _ = repository.NewSessionRepository(rdb) // wire when auth handler is ready

	// Services
	userService := service.NewUserService(userRepo)
	productService := service.NewProductService(productRepo)
	customerService := service.NewCustomerService(cusomerRepo)

	_ = rdb // acknowledge redis client; will be used when SessionService is added

	return &App{
		Config:          cfg,
		UserService:     userService,
		ProductService:  productService,
		CustomerService: customerService,
	}
}

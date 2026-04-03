package router

import (
	"ims-database-util/internal/app"
	"ims-database-util/internal/handler"
	"ims-database-util/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

// Setup creates a chi router configured with standard middleware and the application's HTTP routes.
// Setup creates and returns a configured chi router for the given application.
// 
// The router is populated with standard middleware, a health endpoint, and domain routes:
// - GET /health responds with HTTP 200 and body "OK".
// - A user group protected by HMAC middleware exposes GET /v1/user/profile.
// - A products group exposes GET /v1/products/stream.
// - A customers group exposes GET /v1/customers/stream.
// 
// The App parameter supplies services and configuration used to wire domain handlers and middleware.
func Setup(a *app.App) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	userHandler := handler.NewUserHandler(a.UserService)
	productHandler := handler.NewProductHandler(a.ProductService)
	customerHandler := handler.NewCustomerHandler(a.CustomerService)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireHMAC(a.Config.HMACSecret))
		r.Get("/v1/user/profile", userHandler.GetProfile)
	})

	r.Group(func(r chi.Router) {
		r.Get("/v1/products/stream", productHandler.StreamProducts)
	})

	r.Group(func(r chi.Router) {
		r.Get("/v1/customers/stream", customerHandler.StreamCustomers)
	})
	return r
}

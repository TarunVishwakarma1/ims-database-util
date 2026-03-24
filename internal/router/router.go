package router

import (
	"ims-database-util/internal/config"
	"ims-database-util/internal/handler"
	"ims-database-util/internal/repository"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Setup creates a chi router configured with standard middleware and the application's HTTP routes.
// The router registers request ID, real IP, logging, and recoverer middleware, exposes a public GET /health endpoint that responds "OK", and a grouped set of routes protected by HMAC using cfg.HMACSecret which includes GET /v1/user/profile handled by the user repository-backed handler.
func Setup(cfg *config.Config, userRepo repository.UserRepository) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	userHandler := handler.NewUserHandler(userRepo)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	r.Group(func(r chi.Router) {
		r.Use(handler.RequireHMAC(cfg.HMACSecret))
		r.Get("/v1/user/profile", userHandler.GetProfile)
	})

	return r
}

package server

import (
	"context"
	"ims-database-util/internal/app"
	productpb "ims-database-util/internal/gen/productpb"
	internalgrpc "ims-database-util/internal/grpc"
	"ims-database-util/internal/router"
	"log/slog"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server encapsulates the HTTP and gRPC server lifecycle.
type Server struct {
	http *http.Server
	grpc *grpc.Server
	app  *app.App
}

// New creates a Server wired to the given App.
func New(a *app.App) *Server {
	// HTTP
	mux := router.Setup(a)
	httpSrv := &http.Server{
		Addr:         ":" + a.Config.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// gRPC
	grpcSrv := grpc.NewServer()
	reflection.Register(grpcSrv)
	productpb.RegisterProductServiceServer(
		grpcSrv,
		internalgrpc.NewProductServer(a.ProductService),
	)

	return &Server{
		http: httpSrv,
		grpc: grpcSrv,
		app:  a,
	}
}

// Start launches both servers in background goroutines.
func (s *Server) Start() {
	// HTTP
	go func() {
		slog.Info("🚀 HTTP server running", "port", s.app.Config.Port)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	// gRPC
	go func() {
		addr := ":" + s.app.Config.GRPCPort
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			slog.Error("gRPC listen failed", "error", err)
			return
		}
		slog.Info("🚀 gRPC server running", "port", s.app.Config.GRPCPort)
		if err := s.grpc.Serve(lis); err != nil {
			slog.Error("gRPC server error", "error", err)
		}
	}()
}

// Stop performs graceful shutdown of both servers within the given context deadline.
func (s *Server) Stop(ctx context.Context) {
	slog.Info("Shutting down servers...")

	if err := s.http.Shutdown(ctx); err != nil {
		slog.Error("HTTP shutdown error", "error", err)
	}

	s.grpc.GracefulStop()

	slog.Info("Servers exited properly")
}

package main

import (
	"context"
	_ "fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv"
	_ "github.com/pedrovsilva/taskapi/docs"
	"github.com/pedrovsilva/taskapi/internal/handler"
	"github.com/pedrovsilva/taskapi/internal/repository"
	"github.com/pedrovsilva/taskapi/internal/service"
	_ "github.com/swaggo/http-swagger"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to create database pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	repo := repository.NewPostgresTaskRepository(pool)
	svc := service.NewTaskService(repo)
	h := handler.NewTaskHandler(svc)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1/tasks", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetByID)
			r.Patch("/", h.Update)
			r.Delete("/", h.Delete)
		})
	})

	// Swagger UI — available at /swagger/index.html
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("doc.json")))
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	shutdownCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErr := make(chan error, 1)

	go func() {
		log.Printf("server listening on :%s", port)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:

		log.Printf("server error: %v", err)
		return

	case <-shutdownCtx.Done():
		log.Println("shutdown signal received")
	}

	shutdownTimeout, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownTimeout); err != nil {
		log.Printf("graceful shutdown failed: %v", err)

		if err := server.Close(); err != nil {
			log.Printf("forced server shutdown failed: %v", err)
		}

		return
	}

	log.Println("server stopped gracefully")
}

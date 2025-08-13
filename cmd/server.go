package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pbdeuchler/assistant-server/dao/postgres"
	"github.com/pbdeuchler/assistant-server/service"
)

func Serve(ctx context.Context, cfg Config) error {
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	db, err := postgres.New(ctx, dbPool)
	if err != nil {
		return err
	}

	r := chi.NewRouter()

	// Auth endpoints (unprotected)
	authConfig := service.AuthConfig{
		GCloudClientID:     cfg.GCloudClientID,
		GCloudClientSecret: cfg.GCloudClientSecret,
		GCloudProjectID:    cfg.GCloudProjectID,
		BaseURL:            cfg.BaseURL,
	}
	r.Mount("/oauth", service.NewAuthHandlers(authConfig, db))

	// API endpoints (can be protected with JWT middleware if needed)
	// To protect routes, uncomment the following line:
	// r.Use(service.JWTMiddleware([]byte(cfg.JWTSecret)))

	r.Mount("/todos", service.NewTodos(db))
	r.Mount("/preferences", service.NewPreferences(db))
	r.Mount("/notes", service.NewNotes(db))
	r.Mount("/recipes", service.NewRecipes(db))
	r.Mount("/bootstrap", service.NewBootstrap(db))
	r.Mount("/mcp", service.NewMCPRouter(db, db, db, db))

	addr := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	log.Printf("Starting server on %s", addr)

	srv := &http.Server{Addr: addr, Handler: r}
	go func() { <-ctx.Done(); _ = srv.Shutdown(context.Background()) }()
	return srv.ListenAndServe()
}

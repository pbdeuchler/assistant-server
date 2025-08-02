package cmd

import (
	"context"
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
	r.Mount("/todos", service.NewHandlers(db))
	r.Mount("/backgrounds", service.NewBackgroundHandlers(db))
	r.Mount("/preferences", service.NewPreferencesHandlers(db))
	r.Mount("/notes", service.NewNotesHandlers(db))
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}
	go func() { <-ctx.Done(); _ = srv.Shutdown(context.Background()) }()
	return srv.ListenAndServe()
}

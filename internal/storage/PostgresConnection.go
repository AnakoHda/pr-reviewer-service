package storage

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log/slog"
	"os"
)

func NewConnection(ctx context.Context) (*sqlx.DB, error) {
	URL := os.Getenv("POSTGRES_URL")
	if URL == "" {
		URL = "postgres://authuser:authpass@postgres:5432/authdb?sslmode=disable"
	}

	slog.Info("Connecting to PostgreSQL", "url", URL)
	db, err := sqlx.Connect("postgres", URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

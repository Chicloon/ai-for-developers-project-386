package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Standard Postgres port (Hexlet CI, local Postgres). Docker Compose maps db to host 5434 — set DATABASE_URL in .env.
		dsn = "postgres://postgres:postgres@localhost:5432/call_booking?sslmode=disable"
	}
	return pgxpool.New(ctx, dsn)
}

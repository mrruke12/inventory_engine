package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)

	if err != nil {
		return nil, err
	}

	return pgxpool.NewWithConfig(ctx, config)
}

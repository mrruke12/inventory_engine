package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrConflict     = errors.New("conflict")
	ErrInvalidInput = errors.New("invalid input")
	ErrRetryable    = errors.New("retryable")
	ErrTimeout      = errors.New("timeout")
	ErrInternal     = errors.New("internal")
)

func classifyError(err error) error {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("%w: %s", ErrConflict, pgErr.Message)
		case "23503":
			return fmt.Errorf("%w: %s", ErrConflict, pgErr.Message)
		case "23502":
			return fmt.Errorf("%w: %s", ErrInvalidInput, pgErr.Message)
		case "23514":
			return fmt.Errorf("%w: %s", ErrInvalidInput, pgErr.Message)
		case "22P02", "22003", "22001":
			return fmt.Errorf("%w: %s", ErrInvalidInput, pgErr.Message)
		case "40001", "40P01":
			return fmt.Errorf("%w: %s", ErrRetryable, pgErr.Message)
		case "57014":
			return fmt.Errorf("%w: %s", ErrTimeout, pgErr.Message)
		case "42P01", "42703":
			return fmt.Errorf("%w: %s", ErrInternal, pgErr.Message)
		}
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %v", ErrTimeout, err)
	}

	return fmt.Errorf("unexpected driver error: %w", err)
}

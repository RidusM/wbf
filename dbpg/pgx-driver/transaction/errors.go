package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
)

var (
	// ErrMaxRetriesExceeded is returned when a transaction fails after exhausting all retry attempts.
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")
	// ErrTransactionTimeout is returned when a transaction exceeds its context deadline.
	ErrTransactionTimeout = errors.New("transaction timeout")
	// ErrConflictingData indicates a unique constraint violation (PostgreSQL error code 23505).
	// It is typically returned when inserting or updating data that conflicts with an existing unique row.
	ErrConflictingData = errors.New("data conflicts with existing data in unique column")
	// ErrInvalidData indicates a foreign key violation or other referential integrity error (PostgreSQL error code 23503).
	ErrInvalidData = errors.New("invalid data")
)

// HandleError wraps a raw error with contextual information and maps PostgreSQL error codes
// to semantic, user-friendly error types.
// It takes the operation name (e.g., "transfer_funds"), a step description (e.g., "execute"),
// and the original error, then returns a wrapped error suitable for logging or upstream handling.
// If the input error is nil, it returns nil.
func HandleError(operation string, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%s: timeout: %w", operation, ErrTransactionTimeout)
	}

	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("%s: canceled: %w", operation, err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "40P01":
			return fmt.Errorf("%s: deadlock: %w", operation, err)
		case "40001":
			return fmt.Errorf("%s: serialization failure: %w", operation, err)
		case "57014":
			return fmt.Errorf("%s: statement timeout: %w", operation, err)
		case "55P03":
			return fmt.Errorf("%s: lock timeout: %w", operation, err)
		case "23505":
			return fmt.Errorf(
				"%s: unique constraint violation: %w",
				operation,
				ErrConflictingData,
			)
		case "23503":
			return fmt.Errorf(
				"%s: foreign key violation: %w",
				operation,
				ErrInvalidData,
			)
		}
	}

	if errors.Is(err, ErrMaxRetriesExceeded) {
		return fmt.Errorf("%s: max retries exceeded: %w", operation, err)
	}

	if errors.Is(err, ErrTransactionTimeout) {
		return fmt.Errorf("%s: transaction timeout: %w", operation, err)
	}

	return fmt.Errorf("%s: %w", operation, err)
}

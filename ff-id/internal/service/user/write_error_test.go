package user

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ivasnev/FinFlow/ff-id/internal/service"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

// TestClassifyUserWriteError проверяет распознавание только известных ограничений, включая гонки записи.
func TestClassifyUserWriteError(t *testing.T) {
	for _, constraint := range []string{"users_phone_key", "users_email_key", "users_nickname_key", "users_pkey", "future_unique_key"} {
		t.Run(constraint, func(t *testing.T) {
			original := fmt.Errorf("write: %w", &pgconn.PgError{Code: "23505", ConstraintName: constraint})
			result := classifyUserWriteError(original)
			switch constraint {
			case "users_phone_key":
				require.ErrorIs(t, result, service.ErrPhoneConflict)
			case "future_unique_key":
				require.Same(t, original, result)
			default:
				require.ErrorIs(t, result, service.ErrIdentityConflict)
			}
		})
	}
	for _, original := range []error{errors.New("connection lost"), &pgconn.PgError{Code: "23503", ConstraintName: "users_phone_key"}} {
		require.Same(t, original, classifyUserWriteError(original))
	}
}

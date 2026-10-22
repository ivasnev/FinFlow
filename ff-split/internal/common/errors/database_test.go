package errors

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestDatabaseErrorClassification(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want bool
	}{
		{"postgres driver", &pgconn.PgError{Code: "23505"}, true},
		{"wrapped postgres driver", fmt.Errorf("repository: %w", &pgconn.PgError{Code: "23503"}), true},
		{"sql transaction", fmt.Errorf("repository: %w", sql.ErrTxDone), true},
		{"application", errors.New("application failure"), false},
		{"nil", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isDatabaseError(tc.err); got != tc.want {
				t.Fatalf("database classification=%t, want %t", got, tc.want)
			}
		})
	}
}

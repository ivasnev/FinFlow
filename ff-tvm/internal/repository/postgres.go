package repository

import (
	"context"
	"database/sql"
)

type PostgresRepository struct {
	db *sql.DB
}

func (r *PostgresRepository) GrantAccess(ctx context.Context, from, to int64) error {
	query := `
		INSERT INTO service_access (from_id, to_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, from, to)
	return err
}

func (r *PostgresRepository) RevokeAccess(ctx context.Context, from, to int64) error {
	query := `
		DELETE FROM service_access
		WHERE from_id = $1 AND to_id = $2
	`
	_, err := r.db.ExecContext(ctx, query, from, to)
	return err
}

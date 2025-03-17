package service

import (
	"context"
	"database/sql"
)

type repositoryImpl struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) ServiceRepository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) Create(ctx context.Context, service *Service) error {
	return r.db.QueryRowContext(ctx,
		"INSERT INTO services (name, public_key, private_key_hash) VALUES ($1, $2, $3) RETURNING id",
		service.Name, service.PublicKey, service.PrivateKeyHash,
	).Scan(&service.ID)
}

func (r *repositoryImpl) GetByID(ctx context.Context, id int64) (*Service, error) {
	service := &Service{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, public_key, private_key_hash FROM services WHERE id = $1",
		id,
	).Scan(&service.ID, &service.Name, &service.PublicKey, &service.PrivateKeyHash)
	if err != nil {
		return nil, ErrServiceNotFound
	}
	return service, nil
}

func (r *repositoryImpl) GetPublicKey(ctx context.Context, id int64) (string, error) {
	var publicKey string
	err := r.db.QueryRowContext(ctx,
		"SELECT public_key FROM services WHERE id = $1",
		id,
	).Scan(&publicKey)
	if err != nil {
		return "", ErrServiceNotFound
	}
	return publicKey, nil
}

func (r *repositoryImpl) GetPrivateKeyHash(ctx context.Context, id int64) (string, error) {
	var hash string
	err := r.db.QueryRowContext(ctx,
		"SELECT private_key_hash FROM services WHERE id = $1",
		id,
	).Scan(&hash)
	if err != nil {
		return "", ErrServiceNotFound
	}
	return hash, nil
}

func (r *repositoryImpl) GrantAccess(ctx context.Context, from, to int64) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO service_access (from_id, to_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		from, to,
	)
	return err
}

func (r *repositoryImpl) RevokeAccess(ctx context.Context, from, to int64) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM service_access WHERE from_id = $1 AND to_id = $2",
		from, to,
	)
	return err
}

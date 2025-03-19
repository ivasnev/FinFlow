package service

import (
	"database/sql"
)

type accessManagerImpl struct {
	db *sql.DB
}

func NewAccessManager(db *sql.DB) AccessManager {
	return &accessManagerImpl{db: db}
}

func (m *accessManagerImpl) CheckAccess(from, to int) bool {
	var exists bool
	err := m.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM service_access WHERE from_id = $1 AND to_id = $2)",
		from, to,
	).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func (m *accessManagerImpl) GrantAccess(from, to int) error {
	_, err := m.db.Exec(
		"INSERT INTO service_access (from_id, to_id) VALUES ($1, $2)",
		from, to,
	)
	return err
}

func (m *accessManagerImpl) RevokeAccess(from, to int) error {
	_, err := m.db.Exec(
		"DELETE FROM service_access WHERE from_id = $1 AND to_id = $2",
		from, to,
	)
	return err
}

package service

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCheckAccess(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	manager := NewAccessManager(db)

	// Тест 1: Доступ разрешен
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(int(1), int(2)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	hasAccess := manager.CheckAccess(1, 2)
	assert.True(t, hasAccess)

	// Тест 2: Доступ запрещен
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(int(2), int(1)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	hasAccess = manager.CheckAccess(2, 1)
	assert.False(t, hasAccess)

	// Проверяем, что все ожидаемые запросы были выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGrantAccess(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	manager := NewAccessManager(db)

	// Тест: Успешное предоставление доступа
	mock.ExpectExec("INSERT INTO service_access").
		WithArgs(int(1), int(2)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = manager.GrantAccess(1, 2)
	assert.NoError(t, err)

	// Проверяем, что все ожидаемые запросы были выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRevokeAccess(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	manager := NewAccessManager(db)

	// Тест: Успешное отзыв доступа
	mock.ExpectExec("DELETE FROM service_access").
		WithArgs(int(1), int(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = manager.RevokeAccess(1, 2)
	assert.NoError(t, err)

	// Проверяем, что все ожидаемые запросы были выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

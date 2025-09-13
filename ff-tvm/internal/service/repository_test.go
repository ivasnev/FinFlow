package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	// Тест: Успешное создание сервиса
	service := &Service{
		Name:      "test-service",
		PublicKey: "test-public-key",
	}

	mock.ExpectQuery("INSERT INTO services").
		WithArgs(service.Name, service.PublicKey, service.PrivateKeyHash).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.Create(context.Background(), service)
	assert.NoError(t, err)
	assert.Equal(t, int(1), service.ID)

	// Проверяем, что все ожидаемые запросы были выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	// Тест: Успешное получение сервиса
	mock.ExpectQuery("SELECT id, name, public_key, private_key_hash FROM services").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "public_key", "private_key_hash"}).
			AddRow(1, "test-service", "test-public-key", "test-private-key-hash"))

	service, err := repo.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, 1, service.ID)
	assert.Equal(t, "test-service", service.Name)
	assert.Equal(t, "test-public-key", service.PublicKey)

	// Тест: Сервис не найден
	mock.ExpectQuery("SELECT id, name, public_key, private_key_hash FROM services").
		WithArgs(2).
		WillReturnError(ErrServiceNotFound)

	service, err = repo.GetByID(context.Background(), 2)
	assert.Error(t, err)
	assert.Equal(t, ErrServiceNotFound, err)
	assert.Nil(t, service)

	// Проверяем, что все ожидаемые запросы были выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPublicKey(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	// Тест: Успешное получение публичного ключа
	mock.ExpectQuery("SELECT public_key FROM services").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"public_key"}).
			AddRow("test-public-key"))

	publicKey, err := repo.GetPublicKey(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "test-public-key", publicKey)

	// Тест: Сервис не найден
	mock.ExpectQuery("SELECT public_key FROM services").
		WithArgs(2).
		WillReturnError(ErrServiceNotFound)

	publicKey, err = repo.GetPublicKey(context.Background(), 2)
	assert.Error(t, err)
	assert.Equal(t, ErrServiceNotFound, err)
	assert.Empty(t, publicKey)

	// Проверяем, что все ожидаемые запросы были выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

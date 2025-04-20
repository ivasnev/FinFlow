package postgres

import (
	"fmt"

	"github.com/ivasnev/FinFlow/ff-split/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresRepository представляет собой репозиторий, реализующий все интерфейсы репозиториев для PostgreSQL
type PostgresRepository struct {
	db                  *gorm.DB
	categoryRepo        *CategoryRepository
	eventRepo           *EventRepository
	activityRepo        *ActivityRepository
	transactionRepo     *TransactionRepository
	transactionTypeRepo *TransactionTypeRepository
	iconRepo            *IconRepository
}

// NewPostgresRepository создает новый экземпляр PostgresRepository
func NewPostgresRepository(cfg *config.Config) (*PostgresRepository, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	repo := &PostgresRepository{
		db: db,
	}

	repo.categoryRepo = NewCategoryRepository(db)
	repo.eventRepo = NewEventRepository(db)
	repo.activityRepo = NewActivityRepository(db)
	repo.transactionRepo = NewTransactionRepository(db)
	repo.transactionTypeRepo = NewTransactionTypeRepository(db)
	repo.iconRepo = NewIconRepository(db)

	return repo, nil
}

// CategoryRepository возвращает репозиторий для работы с категориями
func (r *PostgresRepository) CategoryRepository() repository.CategoryRepository {
	return r.categoryRepo
}

// EventRepository возвращает репозиторий для работы с мероприятиями
func (r *PostgresRepository) EventRepository() repository.EventRepository {
	return r.eventRepo
}

// ActivityRepository возвращает репозиторий для работы с активностями
func (r *PostgresRepository) ActivityRepository() repository.ActivityRepository {
	return r.activityRepo
}

// TransactionRepository возвращает репозиторий для работы с транзакциями
func (r *PostgresRepository) TransactionRepository() repository.TransactionRepository {
	return r.transactionRepo
}

// TransactionTypeRepository возвращает репозиторий для работы с типами транзакций
func (r *PostgresRepository) TransactionTypeRepository() repository.TransactionTypeRepository {
	return r.transactionTypeRepo
}

// IconRepository возвращает репозиторий для работы с иконками
func (r *PostgresRepository) IconRepository() repository.IconRepository {
	return r.iconRepo
}

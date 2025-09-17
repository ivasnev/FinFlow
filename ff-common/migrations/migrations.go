package migrations

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/ivasnev/FinFlow/ff-common/logger"
)

func RunMigrations(db *sql.DB, migrationsPath string, log logger.FFLogger) error {
	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %v", err)
	}
	defer driver.Close()

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %v", err)
	}
	defer func(m *migrate.Migrate) {
		err, _ := m.Close()
		if err != nil {
			log.Error("Can't close migration" + err.Error())
			return
		}
	}(m)

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	log.Info("Migrations completed successfully")
	return nil
}

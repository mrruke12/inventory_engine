package postgres

import (
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(pathToMigrations, connString string) error {
	m, err := migrate.New(pathToMigrations, connString)

	if err != nil {
		return err
	}

	err = m.Up()

	if err == migrate.ErrNoChange {
		slog.Info("Migrations are up to date")

		return nil
	}

	if err != nil {
		return err
	}

	slog.Info("Migrations are ran successfully")

	return nil
}

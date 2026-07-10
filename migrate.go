package main

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
)

//go:embed sql/migrations/*.sql
var migrationFS embed.FS

func RunMigrations() error {
	connStr, err := buildConnStr()
	if err != nil {
		return err
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("open database for migration: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	source, err := iofs.New(migrationFS, "sql/migrations")
	if err != nil {
		return fmt.Errorf("load embedded migrations: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			logrus.WithError(sourceErr).Warn("failed to close migration source")
		}
		if dbErr != nil {
			logrus.WithError(dbErr).Warn("failed to close migration database")
		}
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	logrus.Info("Database migrations completed")
	return nil
}

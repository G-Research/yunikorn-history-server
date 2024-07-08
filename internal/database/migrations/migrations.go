package migrations

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/G-Research/yunikorn-history-server/internal/config"
	"github.com/G-Research/yunikorn-history-server/internal/database/postgres"
	"github.com/G-Research/yunikorn-history-server/internal/log"
)

type Migrator struct {
	migrator      *migrate.Migrate
	migrationsDir string
}

func New(cfg *config.PostgresConfig, migrationsDir string) (*Migrator, error) {
	m, err := getMigrator(cfg, migrationsDir)
	if err != nil {
		return nil, err
	}
	return &Migrator{
		migrator:      m,
		migrationsDir: migrationsDir,
	}, nil
}

func getMigrator(cfg *config.PostgresConfig, migrationsDir string) (*migrate.Migrate, error) {
	source := "file://" + migrationsDir
	connString := postgres.BuildConnectionStringFromConfig(cfg)
	return migrate.New(source, connString)
}

func (m *Migrator) Up() (applied bool, err error) {
	log.Logger.Info("running migrate up")

	err = m.migrator.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Logger.Info("no change after running up migrations")
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (m *Migrator) Down() (applied bool, err error) {
	log.Logger.Info("running migrate down")

	err = m.migrator.Down()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Logger.Info("no change after running down migrations")
			return false, nil
		}
		return false, err
	}

	return true, nil
}

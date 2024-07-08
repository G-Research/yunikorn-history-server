package commands

import (
	"github.com/G-Research/yunikorn-history-server/internal/database/migrations"
	"github.com/spf13/cobra"

	"github.com/G-Research/yunikorn-history-server/internal/config"
	"github.com/G-Research/yunikorn-history-server/internal/log"
)

// migrateCmd represents the migrate command which is used to run database migrations
var migrateCmd = &cobra.Command{
	Use:       "migrate up|down",
	Short:     "Run or destroy database migrations.",
	Long:      `Run or destroy database migrations against the configured Postgres database.`,
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{"up", "down"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate(); err != nil {
			return err
		}

		cfg, err := config.NewFromFile(ConfigFile)
		if err != nil {
			return err
		}

		log.Init(&cfg.LogConfig)

		m, err := migrations.New(&cfg.PostgresConfig, MigrationsDir)
		if err != nil {
			return err
		}

		if args[0] == "up" {
			_, err = m.Up()
		} else {
			_, err = m.Down()
		}

		return err
	},
}

func newMigrateCmd() *cobra.Command {
	migrateCmd.Flags().StringVarP(
		&MigrationsDir,
		"migrations-dir",
		"m",
		MigrationsDir,
		"path to the folder containing the database migrations",
	)
	return migrateCmd
}

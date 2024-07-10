package commands

import (
	"context"
	"fmt"
	"github.com/G-Research/yunikorn-history-server/internal/database/postgres"
	"github.com/G-Research/yunikorn-history-server/internal/database/repository"
	"github.com/G-Research/yunikorn-history-server/internal/health"
	"os/signal"
	"syscall"

	"github.com/oklog/run"

	"github.com/spf13/cobra"

	"github.com/G-Research/yunikorn-history-server/cmd/yunikorn-history-server/info"

	"github.com/G-Research/yunikorn-history-server/internal/config"
	"github.com/G-Research/yunikorn-history-server/internal/log"
	"github.com/G-Research/yunikorn-history-server/internal/webservice"
	"github.com/G-Research/yunikorn-history-server/internal/yunikorn"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yunikorn-history-server",
	Short: "Yunikorn History Server warehouses Yunikorn events.",
	Long:  `Yunikorn History Server is a service that listens for events from the Yunikorn Scheduler and stores them in a database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate(); err != nil {
			return err
		}

		cfg, err := config.New(ConfigFile)
		if err != nil {
			return err
		}

		return Run(context.Background(), cfg)
	},
}

// Run is the main entry point for the yunikorn history server.
func Run(ctx context.Context, cfg *config.Config) error {
	log.Init(&cfg.LogConfig)

	log.Logger.Infow(
		"starting yunikorn history server",
		"version", info.Version, "buildTime", info.BuildTime, "commit", info.Commit,
	)

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()

	log.ToContext(ctx, log.Logger)

	pool, err := postgres.NewConnectionPool(ctx, &cfg.PostgresConfig)
	if err != nil {
		return fmt.Errorf("cannot parse Postgres connection config: %w", err)
	}
	mainRepository, err := repository.NewPostgresRepository(pool)
	if err != nil {
		log.Logger.Error("could not create db repository")
		panic(err)
	}
	eventRepository := repository.NewInMemoryEventRepository()

	g := run.Group{}

	client := yunikorn.NewRESTClient(&cfg.YunikornConfig)
	service := yunikorn.NewService(mainRepository, eventRepository, client)
	g.Add(
		func() error {
			return service.RunEventCollector(ctx)
		},
		func(err error) {
			service.Shutdown()
		},
	)
	g.Add(
		func() error {
			return service.RunDataSync(ctx)
		},
		func(err error) {},
	)

	healthService := health.New(info.Version, health.NewYunikornComponent(client), health.NewPostgresComponent(pool))

	ws := webservice.NewWebService(&cfg.YHSConfig, mainRepository, eventRepository, healthService)
	g.Add(
		func() error {
			return ws.Start(ctx)
		},
		func(err error) {
			_ = ws.Shutdown(ctx)
		},
	)

	if err = g.Run(); err != nil {
		log.Logger.Warnf("group stopped because of an error: %v", err)
	}

	return nil
}

func New() *cobra.Command {
	rootCmd.PersistentFlags().StringVarP(&ConfigFile, "config", "c", ConfigFile, "path to the configuration file")
	rootCmd.AddCommand(newMigrateCmd())
	return rootCmd
}

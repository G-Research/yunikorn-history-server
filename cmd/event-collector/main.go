package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/G-Research/yunikorn-history-server/internal/config"
	"github.com/G-Research/yunikorn-history-server/internal/repository"
	"github.com/G-Research/yunikorn-history-server/internal/webservice"
	"github.com/G-Research/yunikorn-history-server/internal/ykclient"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var (
	httpProto     string
	ykHost        string
	ykPort        int
	yhsServerAddr string
	eventCounts   config.EventTypeCounts
)

func loadConfig(cfgFile string) (*koanf.Koanf, error) {
	k := koanf.New(".")

	// Try to load from the config file if it's provided
	if cfgFile != "" {
		if _, err := os.Stat(cfgFile); err == nil {
			if err := k.Load(file.Provider(cfgFile), yaml.Parser()); err == nil {
				// Successfully loaded from file, return
				return k, nil
			}
		}
	}

	// If there's no config file or there's an error reading it, default to env vars
	// YHS_PARENT1_CHILD1_NAME will be converted into "parent1.child1.name"
	if err := k.Load(env.Provider("YHS_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "YHS_")), "_", ".", -1)
	}), nil); err != nil {
		return nil, fmt.Errorf("error loading environment variables: %v", err)
	}

	return k, nil
}

func main() {
	cfgFile := ""
	if len(os.Args) == 2 {
		cfgFile = os.Args[1]
	}

	k, err := loadConfig(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	httpProto = k.String("yunikorn.protocol")
	ykHost = k.String("yunikorn.host")
	ykPort = k.Int("yunikorn.port")
	yhsServerAddr = k.String("yhs.serverAddr")

	pgCfg := config.PostgresConfig{
		Host:     k.String("db.host"),
		Port:     k.Int("db.port"),
		Username: k.String("db.user"),
		Password: k.String("db.password"),
		DbName:   k.String("db.dbname"),
	}

	if k.Int("pool_max_conns") > 0 {
		pgCfg.PoolMaxConns = k.Int("pool_max_conns")
	}
	if k.Int("pool_min_conns") > 0 {
		pgCfg.PoolMinConns = k.Int("pool_min_conns")
	}
	if k.Duration("pool_max_conn_lifetime") > time.Duration(0) {
		pgCfg.PoolMaxConnLifetime = k.Duration("pool_max_conn_lifetime")
	}
	if k.Duration("pool_max_conn_idletime") > time.Duration(0) {
		pgCfg.PoolMaxConnIdleTime = k.Duration("pool_max_conn_idletime")
	}

	eventCounts = config.EventTypeCounts{}

	ecConfig := config.ECConfig{
		PostgresConfig: pgCfg,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo, err := repository.NewECRepo(ctx, &ecConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create db repository: %v\n", err)
		os.Exit(1)
	}

	repo.Setup(ctx)

	ctx = context.WithValue(ctx, config.EventCounts, eventCounts)

	client := ykclient.NewClient(httpProto, ykHost, ykPort, repo)
	client.Run(ctx)

	ws := webservice.NewWebService(yhsServerAddr, repo)
	ws.Start(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	<-signalChan

	fmt.Println("Received signal, YHS shutting down...")
}

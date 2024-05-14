package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/G-Research/yunikorn-history-server/internal/config"
	"github.com/G-Research/yunikorn-history-server/internal/repository"
	"github.com/G-Research/yunikorn-history-server/internal/webservice"
	"github.com/G-Research/yunikorn-history-server/internal/ykclient"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var (
	httpProto     string
	ykHost        string
	ykPort        int
	yhsServerAddr string
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s path/to/config.yml\n", os.Args[0])
		os.Exit(1)
	}

	cfgFile := os.Args[1]
	if _, err := os.Stat(cfgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot open config file %s: %v\n", cfgFile, err)
		os.Exit(1)
	}

	var k = koanf.New(".")
	if err := k.Load(file.Provider(cfgFile), yaml.Parser()); err != nil {
		fmt.Fprintf(os.Stderr, "error loading file: %v", err)
		os.Exit(1)
	}

	httpProto = k.String("yunikorn.protocol")
	ykHost = k.String("yunikorn.host")
	ykPort = k.Int("yunikorn.port")
	yhsServerAddr = k.String("yhs.serverAddr")

	ecConfig := config.ECConfig{
		PostgresConfig: config.PostgresConfig{
			Host:     k.String("db.host"),
			Port:     k.Int("db.port"),
			Username: k.String("db.user"),
			Password: k.String("db.password"),
			DbName:   k.String("db.dbname"),
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo, err := repository.NewECRepo(ctx, &ecConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create db repository: %v\n", err)
		os.Exit(1)
	}

	repo.Setup(ctx)

	client := ykclient.NewClient(httpProto, ykHost, ykPort, repo)
	client.Run(ctx)

	ws := webservice.NewWebService(yhsServerAddr, repo)
	ws.Start(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	<-signalChan

	fmt.Println("Received signal, YHS shutting down...")
}

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
)

var (
	httpProto     = "http"
	ykHost        = "127.0.0.1"
	ykPort        = 9889
	yhsServerAddr = ":8989"
)

func main() {

	ecConfig := config.ECConfig{
		PostgresConfig: config.PostgresConfig{
			Connection: map[string]string{
				"dbname":   "yhs",
				"user":     "postgres",
				"password": "psw",
				"host":     "localhost",
				"port":     "5432",
			},
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

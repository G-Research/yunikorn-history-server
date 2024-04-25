package main

import (
	"context"
	"fmt"
	"os"

	//"github.com/apache/yunikorn-scheduler-interface/lib/go/common"
	"richscott/yhs/internal/event-collector/config"
	"richscott/yhs/internal/event-collector/repository"
	"richscott/yhs/internal/event-collector/ykclient"
)

var (
	httpProto = "http"
	ykHost    = "127.0.0.1"
	ykPort    = 9889
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

	err, repo := repository.NewECRepo(&ecConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create db repository: %v\n", err)
		os.Exit(1)
	}

	repo.Setup(context.Background())

	client := ykclient.NewClient(httpProto, ykHost, ykPort, repo)
	if err := client.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "could not run client: %v\n", err)
		os.Exit(1)
	}
}

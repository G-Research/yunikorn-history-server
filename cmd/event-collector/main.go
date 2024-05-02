package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	//"github.com/apache/yunikorn-scheduler-interface/lib/go/common"
	"richscott/yhs/internal/config"
	"richscott/yhs/internal/repository"
	"richscott/yhs/internal/webservice"
	"richscott/yhs/internal/ykclient"
)

var (
	httpProto     = "http"
	ykHost        = "127.0.0.1"
	ykPort        = 9889
	yhsServerAddr = ":8989"
)

// TODO:
// - Start an appropriate context with cancel, pass it around the services
// - Add a graceful shutdown mechanism to handle OS signals and cancel the context
// - Add a logger (zap) to the services
// - Add a configiration handler through yaml, and make all variables configurable
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

	err, repo := repository.NewECRepo(&ecConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create db repository: %v\n", err)
		os.Exit(1)
	}

	repo.Setup(ctx)

	ws := webservice.NewWebService(yhsServerAddr, repo)

	client := ykclient.NewClient(httpProto, ykHost, ykPort, repo)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := client.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "could not run client: %v\n", err)
			os.Exit(1)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := ws.Start(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "could not run webservice: %v\n", err)
			os.Exit(1)
		}
	}()
	wg.Wait()

}

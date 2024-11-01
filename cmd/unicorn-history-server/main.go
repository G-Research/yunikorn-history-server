package main

import (
	"context"
	"os"

	_ "go.uber.org/mock/mockgen/model"

	"github.com/G-Research/unicorn-history-server/cmd/unicorn-history-server/commands"
)

func main() {
	ctx := context.Background()
	if err := commands.New().ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

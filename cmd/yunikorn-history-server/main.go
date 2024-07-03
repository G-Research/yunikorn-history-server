package main

import (
	"context"
	"os"

	_ "go.uber.org/mock/mockgen/model"

	"github.com/G-Research/yunikorn-history-server/cmd/yunikorn-history-server/commands"
)

func main() {
	ctx := context.Background()
	if err := commands.New().ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

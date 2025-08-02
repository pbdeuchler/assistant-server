package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/pbdeuchler/assistant-server/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cfg := cmd.LoadConfig()
	_ = cmd.Serve(ctx, cfg)
}

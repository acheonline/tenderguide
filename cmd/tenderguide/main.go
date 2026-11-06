package main

import (
	"context"
	"log/slog"
	"os"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.Error("tenderguide failed", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	_ = ctx
	slog.Info("tenderguide started")
	return nil
}

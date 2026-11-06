package main

import (
	"context"
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	if err := run(context.Background()); err != nil {
		slog.Error("tenderguide failed", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	_ = ctx
	slog.Info("tenderguide starting")
	return nil
}

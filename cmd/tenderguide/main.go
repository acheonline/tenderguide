package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: tenderguide <command> [flags]")
		fmt.Fprintln(os.Stderr, "commands:")
		fmt.Fprintln(os.Stderr, "  load --dir <path> [--db <path>]")
		fmt.Fprintln(os.Stderr, "  load-doc --dir <path> [--db <path>]")
		fmt.Fprintln(os.Stderr, "  show --request <id> [--db <path>] [--format text|json]")
		return 2
	}

	ctx := context.Background()
	switch args[0] {
	case "load":
		return runLoad(ctx, args[1:])
	case "load-doc":
		return runLoadDoc(ctx, args[1:])
	case "show":
		return runShow(ctx, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		return 2
	}
}

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/acheonline/tenderguide/internal/loader"
	"github.com/acheonline/tenderguide/internal/parser"
	"github.com/acheonline/tenderguide/internal/storage"
)

const defaultDBPath = "./tenderguide.db"

func runLoad(ctx context.Context, args []string) int {
	fs := flag.NewFlagSet("load", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dir := fs.String("dir", "", "path to directory with DOCX files")
	db := fs.String("db", defaultDBPath, "path to SQLite database")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *dir == "" {
		fmt.Fprintln(os.Stderr, "usage: tenderguide load --dir <path> [--db <path>]")
		return 2
	}

	return load(ctx, *dir, *db, os.Stdout, os.Stderr)
}

func load(ctx context.Context, dir, dbPath string, stdout, stderr io.Writer) int {
	requestID := filepath.Base(filepath.Clean(dir))

	inputs, err := loader.Load(dir)
	if err != nil {
		fmt.Fprintf(stderr, "load: %v\n", err)
		return 1
	}

	all := parser.Extract(requestID, inputs)

	dsn := "file:" + dbPath + "?_pragma=foreign_keys(1)"
	store, err := storage.Open(ctx, dsn)
	if err != nil {
		fmt.Fprintf(stderr, "load: %v\n", err)
		return 1
	}
	defer store.Close()

	id, err := store.Save(ctx, all)
	if err != nil {
		if errors.Is(err, storage.ErrDuplicateRequest) {
			fmt.Fprintf(stderr, "duplicate request_id: %s\n", requestID)
			return 1
		}
		fmt.Fprintf(stderr, "load: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "OK: request %s saved (id=%d)\n", requestID, id)
	return 0
}

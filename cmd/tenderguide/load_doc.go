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

func runLoadDoc(ctx context.Context, args []string) int {
	fs := flag.NewFlagSet("load-doc", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dir := fs.String("dir", "", "path to directory with documentation DOCX files")
	db := fs.String("db", defaultDBPath, "path to SQLite database")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *dir == "" {
		fmt.Fprintln(os.Stderr, "usage: tenderguide load-doc --dir <path> [--db <path>]")
		return 2
	}

	return loadDoc(ctx, *dir, *db, os.Stdout, os.Stderr)
}

func loadDoc(ctx context.Context, dir, dbPath string, stdout, stderr io.Writer) int {
	docID := filepath.Base(filepath.Clean(dir))

	inputs, err := loader.Load(dir)
	if err != nil {
		fmt.Fprintf(stderr, "load-doc: %v\n", err)
		return 1
	}

	all := parser.ExtractDoc(docID, inputs)

	dsn := "file:" + dbPath + "?_pragma=foreign_keys(1)"
	store, err := storage.Open(ctx, dsn)
	if err != nil {
		fmt.Fprintf(stderr, "load-doc: %v\n", err)
		return 1
	}
	defer store.Close()

	id, err := store.SaveDoc(ctx, all)
	if err != nil {
		if errors.Is(err, storage.ErrDuplicateDoc) {
			fmt.Fprintf(stderr, "duplicate doc_id: %s\n", docID)
			return 1
		}
		fmt.Fprintf(stderr, "load-doc: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "OK: doc %s saved (id=%d)\n", docID, id)
	return 0
}

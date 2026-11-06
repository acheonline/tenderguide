package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	ctx := context.Background()
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: tenderguide <load|load-doc|show|show-doc> [flags]")
		os.Exit(2)
	}

	var code int
	switch os.Args[1] {
	case "load":
		code = runLoad(ctx, os.Args[2:])
	case "load-doc":
		code = runLoadDoc(ctx, os.Args[2:])
	case "show":
		code = runShow(ctx, os.Args[2:])
	case "show-doc":
		code = runShowDoc(ctx, os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		code = 2
	}
	os.Exit(code)
}

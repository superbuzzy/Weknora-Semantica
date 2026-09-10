package main

import (
	"context"
	"fmt"
	"os"

	"cobraknowledge.local/cobra-knowledge/internal/mcp"
)

func main() {
	server, err := mcp.NewFromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cobra context mcp init:", err)
		os.Exit(1)
	}
	if err := server.Serve(context.Background(), os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "cobra context mcp:", err)
		os.Exit(1)
	}
}

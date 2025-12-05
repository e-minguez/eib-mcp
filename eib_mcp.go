package main

import (
	"fmt"
	"os"

	"github.com/e-minguez/eib-mcp/mcp"
)

func main() {
	server := mcp.NewServer(os.Stdin, os.Stdout)
	if err := server.Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

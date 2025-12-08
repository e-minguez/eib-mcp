// Package main is the entry point for the Edge Image Builder (EIB) MCP Server.
//
// It initializes the MCP server and starts listening for JSON-RPC 2.0 messages
// on Standard Input and writing responses to Standard Output.
package main

import (
	"fmt"
	"os"

	"github.com/e-minguez/eib-mcp/mcp"
)

// main initializes and runs the EIB MCP server.
//
// It creates a new Server instance connected to os.Stdin and os.Stdout,
// and starts the server loop. If the server encounters a fatal error,
// it prints the error to os.Stderr and exits with status code 1.
func main() {
	server := mcp.NewServer(os.Stdin, os.Stdout)
	if err := server.Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

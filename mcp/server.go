// Package mcp implements the Model Context Protocol (MCP) server logic.
//
// It handles JSON-RPC 2.0 requests and responses, providing specific tools
// for generating EIB configurations.
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/e-minguez/eib-mcp/schema"
	"github.com/e-minguez/eib-mcp/tool"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request.
//
// It encapsulates the method to be called, its parameters, and the request ID.
type JSONRPCRequest struct {
	// JSONRPC specifies the version of the JSON-RPC protocol. Must be "2.0".
	JSONRPC string `json:"jsonrpc"`
	// Method is the name of the method to be invoked.
	Method string `json:"method"`
	// Params contains the parameter values for the method.
	Params json.RawMessage `json:"params,omitempty"`
	// ID is a unique identifier established by the client.
	ID interface{} `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
//
// It contains the result of a successful method invocation or an error object.
type JSONRPCResponse struct {
	// JSONRPC specifies the version of the JSON-RPC protocol. Must be "2.0".
	JSONRPC string `json:"jsonrpc"`
	// Result contains the data returned by the invoked method.
	Result interface{} `json:"result,omitempty"`
	// Error provides details about any error that occurred during execution.
	Error *JSONRPCError `json:"error,omitempty"`
	// ID matches the identifier of the request this response corresponds to.
	ID interface{} `json:"id"`
}

// JSONRPCError represents a JSON-RPC 2.0 error.
//
// It provides a code and a message to describe the error.
type JSONRPCError struct {
	// Code is an integer indicating the error type.
	Code int `json:"code"`
	// Message is a concise string description of the error.
	Message string `json:"message"`
	// Data provides additional information about the error.
	Data interface{} `json:"data,omitempty"`
}

// Server implements the MCP server.
//
// It reads JSON-RPC requests from an input stream and writes responses
// to an output stream.
type Server struct {
	in  io.Reader
	out io.Writer
}

// NewServer creates a new MCP server.
//
// It takes an input reader and an output writer for communication.
//
// Parameters:
//   - in: The io.Reader to read requests from.
//   - out: The io.Writer to write responses to.
//
// Returns:
//   - *Server: A pointer to the newly created Server instance.
func NewServer(in io.Reader, out io.Writer) *Server {
	return &Server{in: in, out: out}
}

// Serve starts the server loop.
//
// It continuously reads from the input stream, processes requests,
// and writes responses to the output stream until the input is closed
// or an error occurs.
//
// Returns:
//   - error: An error if reading from the input fails, or nil on clean exit.
func (s *Server) Serve() error {
	scanner := bufio.NewScanner(s.in)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			// Ignore invalid JSON or log it?
			// For now, just continue or send parse error if we can identify it's a request.
			continue
		}

		resp := s.handleRequest(&req)
		if resp != nil {
			bytes, err := json.Marshal(resp)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to marshal response: %v\n", err)
				continue
			}
			s.out.Write(bytes)
			s.out.Write([]byte("\n"))
		}
	}
	return scanner.Err()
}

// handleRequest processes a single JSON-RPC request and returns a response.
//
// It routes the request to the appropriate handler based on the method name.
//
// Parameters:
//   - req: The incoming JSON-RPC request.
//
// Returns:
//   - *JSONRPCResponse: The response to be sent back to the client, or nil if no response is needed (e.g. notifications).
func (s *Server) handleRequest(req *JSONRPCRequest) *JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	default:
		// Ignore notifications or unknown methods
		if req.ID != nil {
			return &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &JSONRPCError{
					Code:    -32601,
					Message: "Method not found",
				},
			}
		}
		return nil
	}
}

// handleInitialize handles the "initialize" method.
//
// It returns the server's protocol version, capabilities, and information.
//
// Parameters:
//   - req: The initialize request.
//
// Returns:
//   - *JSONRPCResponse: The response containing server details.
func (s *Server) handleInitialize(req *JSONRPCRequest) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "eib-mcp",
				"version": "0.1.0",
			},
		},
	}
}

// handleToolsList handles the "tools/list" method.
//
// It returns a list of available tools, including "generate_config",
// along with their descriptions and input schemas.
//
// Parameters:
//   - req: The tools/list request.
//
// Returns:
//   - *JSONRPCResponse: The response containing the list of tools.
func (s *Server) handleToolsList(req *JSONRPCRequest) *JSONRPCResponse {
	// Load schema to embed in tool definition
	schemaBytes := schema.GetRawSchema()
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
		// Should not happen with embedded valid JSON
		schemaMap = map[string]interface{}{"type": "object", "error": "failed to parse schema"}
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name": "generate_config",
					"description": `Generates a valid edge-image-builder YAML configuration file.
IMPORTANT GUIDELINES:
1. "kubernetes.helm.charts.repositoryName" MUST match a "name" in "kubernetes.helm.repositories".
2. "kubernetes.nodes" MUST NOT contain IP addresses (only hostname, type, initializer).
3. "operatingSystem.time" MUST use "timezone" (lowercase), NOT "timeZone".
4. Passwords: You can put plaintext in "encryptedPassword" or "password". The tool will automatically encrypt it.

Example Structure:
apiVersion: "1.0"
image:
  imageType: "iso"
  arch: "x86_64"
  baseImage: "sles15.iso"
  outputImageName: "output"
operatingSystem:
  users:
    - username: "root"
      encryptedPassword: "..."
  isoConfiguration:
    installDevice: "/dev/sda"
  time:
    timezone: "UTC"
    ntp:
      servers:
        - "pool.ntp.org"
kubernetes:
  version: "1.29.0"
  network:
    apiVIP: "1.2.3.4"
  nodes:
    - hostname: "node1"
      type: "server"
  helm:
    charts:
      - name: "chart"
        repositoryName: "repo"
        version: "1.0.0"
    repositories:
      - name: "repo"
        url: "https://charts.example.com"`,
					"inputSchema": schemaMap,
				},
			},
		},
	}
}

// handleToolsCall handles the "tools/call" method.
//
// It executes the requested tool (currently only "generate_config")
// with the provided arguments.
//
// Parameters:
//   - req: The tools/call request containing the tool name and arguments.
//
// Returns:
//   - *JSONRPCResponse: The response containing the tool's output or an error.
func (s *Server) handleToolsCall(req *JSONRPCRequest) *JSONRPCResponse {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &JSONRPCError{Code: -32700, Message: "Parse error"},
		}
	}

	if params.Name != "generate_config" {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &JSONRPCError{Code: -32601, Message: "Tool not found"},
		}
	}

	yamlOutput, err := tool.GenerateConfig(params.Arguments)
	if err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &JSONRPCError{Code: -32000, Message: err.Error()},
		}
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": yamlOutput,
				},
			},
		},
	}
}

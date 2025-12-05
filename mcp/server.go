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
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      interface{}   `json:"id"`
}

// JSONRPCError represents a JSON-RPC 2.0 error.
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Server implements the MCP server.
type Server struct {
	in  io.Reader
	out io.Writer
}

// NewServer creates a new MCP server.
func NewServer(in io.Reader, out io.Writer) *Server {
	return &Server{in: in, out: out}
}

// Serve starts the server loop.
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

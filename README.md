# Edge Image Builder MCP Server

This repository contains a Model Context Protocol (MCP) server that provides tools for generating SUSE Edge Image Builder (EIB) configuration files.

## Overview

The EIB MCP Server exposes a `generate_config` tool that allows AI agents to validate and generate YAML configuration files based on the official EIB schema. This ensures that generated configurations are syntactically correct and adhere to the required structure.

## Features

- **Schema Validation**: Uses the embedded EIB JSON schema to validate inputs.
- **YAML Generation**: Converts valid JSON inputs into properly formatted YAML.
- **MCP Protocol**: Implements the Model Context Protocol (JSON-RPC 2.0) over Stdio.

## Installation

### Prerequisites

- Go 1.21 or later

### Building

Clone the repository and build the binary:

```bash
git clone https://github.com/e-minguez/eib-mcp.git
cd eib-mcp
make build
```

## Usage

The server is designed to be run by an MCP client. It communicates via Standard Input/Output (Stdio).

### Client Configuration

You can add the server to `gemini-cli` using the `mcp add` command:

```bash
gemini mcp add eib-mcp /absolute/path/to/eib-mcp/eib-mcp
```

### Example Usage

Once the server is added, you can ask Gemini to generate configurations:

> "Generate an Edge Image Builder configuration for an ISO image based on slmicro-6.2.iso, targeting x86_64 architecture. The output name should be 'my-edge-image' and it should install to /dev/sda. It should deploy a 3 nodes kubernetes cluster with nodes names "node1", "node2" and "node3" as:
* hostname: node1, IP: 1.1.1.1, role: initializer
* hostname: node2, IP: 1.1.1.2, role: agent
* hostname: node3, IP: 1.1.1.3, role: agent
The kubernetes version should be k3s 1.33.4-k3s1 and it should deploy a cert-manager helm chart (the latest one available). It should create a user called "suse" with password "suse" and set ntp to "foo.ntp.org". The VIP address for the API should be 1.2.3.4"

Gemini will use the `generate_config` tool to create the valid YAML.

#### `generate_config`

Generates a valid `edge-image-builder` YAML configuration file.

**Input:**

A JSON object matching the EIB configuration schema.

**Output:**

A YAML string representing the configuration.

## Development

### Project Structure

- `eib_mcp.go`: Main entry point.
- `mcp/`: MCP server implementation.
- `schema/`: Schema loading and embedding.
- `tool/`: Tool logic and validation.

### Testing

You can verify the server functionality using the provided test request:

```bash
./eib-mcp < test_request.json
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

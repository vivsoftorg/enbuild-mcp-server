# ENBUILD MCP Server

A Model Context Protocol (MCP) server for the ENBUILD Platform that provides tools for managing ENBUILD catalogs through the Amazon Q interface.

## Overview

This MCP server integrates with the ENBUILD SDK to provide a set of tools for managing ENBUILD catalogs:

- `list_catalogs`: Lists all catalogs for a given VCS type (GITHUB or GITLAB)
- `get_catalog_details`: Fetches details of all catalogs that match a specific catalog ID
- `search_catalogs`: Search for catalogs using name, filtered by catalog type and VCS

## Installation

### Prerequisites

- Go 1.23 or later
- AWS credentials configured

### Building from source

```bash
git clone https://github.com/vivsoftorg/mcp-server-enbuild.git
cd mcp-server-enbuild
go build
```

## Usage

> **Note:** The ENBUILD API token and base URL are required. You must provide them either via command-line flags or environment variables.

### Starting the server

The server supports two transport modes:

#### Stdio Transport (default)

```bash
./mcp-server-enbuild
```

or explicitly:

```bash
./mcp-server-enbuild --transport stdio
```

#### SSE Transport

```bash
./mcp-server-enbuild --transport sse --sse-address :8080
```

### Command-line Options

```
 -base-url string
    	Base URL for the ENBUILD (default "https://enbuild.vivplatform.io")
  -debug
    	Enable debug mode for the ENBUILD client
  -log-level string
    	Log level (debug, info, warn, error) (default "info")
  -password string
    	password for ENBUILD
  -sse-address string
    	The host and port to start the SSE server on (default ":8080")
  -transport string
    	Transport type (stdio or sse) (default "stdio")
  -username string
    	username for ENBUILD
```

You can also set the following environment variables instead of using command-line flags:
- `ENBUILD_USERNAME`: username for ENBUILD
- `ENBUILD_PASSWORD`: password for ENBUILD
- `ENBUILD_BASE_URL`: Base URL for the ENBUILD

### Registering with Amazon Q

To use this MCP server with Amazon Q, add it to your Amazon Q configuration:

```bash
# For stdio transport
q config add-mcp-server enbuild stdio

# For SSE transport
q config add-mcp-server enbuild http://localhost:8080
```

### Usage with VS Code

Add the following JSON block to your User Settings (JSON) file in VS Code. You can do this by pressing Ctrl + Shift + P and typing Preferences: Open User Settings (JSON).

```json
{
  "servers": {
    "enbuild": {
      "type": "stdio",
      "command": "/usr/local/bin/mcp-server-enbuild",
      "args": [
        "--base-url",
        "https://enbuild-dev.vivplatform.io"
      ],
      "env": {
        "ENBUILD_BASE_URL": "https://enbuild-dev.vivplatform.io",
        "ENBUILD_USERNAME": "username",
        "ENBUILD_PASSWORD": "password"
      },
    }
  }
}

```

### Registering with Other tools like Claude Desktop

```
{
  "mcpServers": {
    "enbuild": {
      "type": "stdio",
      "command": "/usr/local/bin/mcp-server-enbuild",
      "args": [
        "--base-url",
        "https://enbuild-dev.vivplatform.io"
      ],
      "env": {
        "ENBUILD_BASE_URL": "https://enbuild.vivplatform.io",
        "ENBUILD_USERNAME": "username",
        "ENBUILD_PASSWORD": "password"
      }
    }
  }
}
```

## Tool Configuration
The MCP server provides the following tools:
- `search_catalogs`: List all catalogs for a specific VCS
- `get_catalog_details`: Get catalog details by ID

### Using the tools

Once registered, you can use the ENBUILD tools in Amazon Q:

```
# List all catalogs for a specific VCS
search_catalogs --vcs "GITHUB"

# Get catalog details
get_catalog_details --id "catalog-id"

# Search for catalogs by name with required filters
search_catalogs --name "terraform" --type "terraform" --vcs "GITHUB"
```

All tools return a consistent JSON response format:

```json
{
  "success": true,
  "message": "Successfully retrieved catalogs",
  "count": 5,
  "data": [
    {
      "id": "catalog-id",
      "name": "catalog-name",
      "type": "terraform",
      "vcs": "GITHUB",
      "slug": "catalog-slug"
    }
  ]
}
```

## Development

### Adding new tools

To add new tools, modify the `registerTools` function in `mcpenbuild.go` and implement the corresponding handler function.

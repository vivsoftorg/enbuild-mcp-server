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
        Base URL for the ENBUILD API
  -debug
        Enable debug mode for the ENBUILD client
  -log-level string
        Log level (debug, info, warn, error) (default "info")
  -sse-address string
        The host and port to start the SSE server on (default ":8080")
  -t string
        Transport type (stdio or sse) (default "stdio")
  -token string
        API token for ENBUILD
  -transport string
        Transport type (stdio or sse) (default "stdio")
```

You can also set the following environment variables instead of using command-line flags:
- `ENBUILD_API_TOKEN`: API token for ENBUILD
- `ENBUILD_BASE_URL`: Base URL for the ENBUILD API

### Registering with Amazon Q

To use this MCP server with Amazon Q, add it to your Amazon Q configuration:

```bash
# For stdio transport
q config add-mcp-server enbuild stdio

# For SSE transport
q config add-mcp-server enbuild http://localhost:8080
```

### Registering with Other tools like Claude Desktop

```
{
  "mcpServers": {
    "enbuild": {
      "command": "mcp-server-enbuild",
      "args": [],
      "env": {
        "ENBUILD_BASE_URL": "https://enbuild-dev.vivplatform.io/enbuild-bk/",
        "ENBUILD_API_TOKEN": "your_api_token_here"
      }
    }
  }
}
```

### Using the tools

Once registered, you can use the ENBUILD tools in Amazon Q:

```
# List all catalogs for a specific VCS
enbuild___list_catalogs --vcs "GITHUB"

# Get catalog details
enbuild___get_catalog_details --id "catalog-id"

# Search for catalogs by name with required filters
enbuild___search_catalogs --name "terraform" --type "terraform" --vcs "GITHUB"
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

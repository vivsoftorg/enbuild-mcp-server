# ENBUILD MCP Server

A Model Context Protocol (MCP) server for the ENBUILD Platform that provides tools for managing ENBUILD catalogs through the Amazon Q interface.

## Overview

This MCP server integrates with the ENBUILD SDK to provide a set of tools for managing ENBUILD catalogs:

- `list-catalogs`: List all ENBUILD catalogs
- `get-catalog`: Get details of a specific ENBUILD catalog
- `search-catalogs`: Search for ENBUILD catalogs by name
- `filter-catalogs-by-type`: Filter ENBUILD catalogs by type
- `filter-catalogs-by-vcs`: Filter ENBUILD catalogs by VCS

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
# List all catalogs
enbuild___list-catalogs

# Get catalog details
enbuild___get-catalog --id "catalog-id"

# Search for catalogs by name
enbuild___search-catalogs --name "terraform"

# Filter catalogs by type
enbuild___filter-catalogs-by-type --type "terraform"

# Filter catalogs by VCS
enbuild___filter-catalogs-by-vcs --vcs "github"
```

## Development

### Adding new tools

To add new tools, modify the `registerTools` function in `mcpenbuild.go` and implement the corresponding handler function.



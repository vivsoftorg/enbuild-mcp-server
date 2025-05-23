# ENBUILD MCP Server

[![Go Reference](https://pkg.go.dev/badge/github.com/vivsoftorg/mcp-server-enbuild.svg)](https://pkg.go.dev/github.com/vivsoftorg/mcp-server-enbuild)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A Model Context Protocol (MCP) server for the ENBUILD Platform. Provides tools for managing ENBUILD catalogs and integrates with Amazon Q, VS Code, and other MCP-compatible clients.

---

## Features

- List all ENBUILD catalogs for a given VCS (GITHUB or GITLAB)
- Fetch details for a specific catalog by ID
- Search catalogs by name, type, and VCS
- Supports stdio and SSE transports
- Easy integration with Amazon Q, VS Code, and other tools

---

## Quickstart

### Prerequisites

- Go 1.23 or later
- AWS credentials (if required by your environment)

### Build

```bash
git clone https://github.com/vivsoftorg/mcp-server-enbuild.git
cd mcp-server-enbuild
go build -o mcp-server-enbuild
```

### Run

#### Stdio (default)

```bash
./mcp-server-enbuild
```

#### SSE (HTTP)

```bash
./mcp-server-enbuild --transport sse --sse-address :8080
```

---

## Configuration

You can configure the server using command-line flags or environment variables:

| Flag            | Env Var              | Description                                   | Default                        |
|-----------------|---------------------|-----------------------------------------------|--------------------------------|
| `-base-url`     | `ENBUILD_BASE_URL`   | Base URL for ENBUILD                          | https://enbuild.vivplatform.io |
| `-username`     | `ENBUILD_USERNAME`   | Username for ENBUILD                          |                                |
| `-password`     | `ENBUILD_PASSWORD`   | Password for ENBUILD                          |                                |
| `-transport`    |                      | Transport type: `stdio` or `sse`              | stdio                          |
| `-sse-address`  |                      | Host:port for SSE server                      | :8080                          |
| `-log-level`    |                      | Log level: debug, info, warn, error           | info                           |
| `-debug`        |                      | Enable debug mode                             | false                          |

---

## Usage

### Register with Amazon Q

```bash
# Stdio
q config add-mcp-server enbuild stdio

# SSE
q config add-mcp-server enbuild http://localhost:8080
```

### VS Code Example

Add to your User Settings (JSON):

```json
{
  "servers": {
    "enbuild": {
      "type": "stdio",
      "command": "/usr/local/bin/mcp-server-enbuild",
      "args": ["--base-url", "https://enbuild-dev.vivplatform.io"],
      "env": {
        "ENBUILD_BASE_URL": "https://enbuild-dev.vivplatform.io",
        "ENBUILD_USERNAME": "username",
        "ENBUILD_PASSWORD": "password"
      }
    }
  }
}
```

### Claude Desktop Example

```json
{
  "mcpServers": {
    "enbuild": {
      "type": "stdio",
      "command": "/usr/local/bin/mcp-server-enbuild",
      "args": ["--base-url", "https://enbuild-dev.vivplatform.io"],
      "env": {
        "ENBUILD_BASE_URL": "https://enbuild.vivplatform.io",
        "ENBUILD_USERNAME": "username",
        "ENBUILD_PASSWORD": "password"
      }
    }
  }
}
```

---

## Tools

The following tools are provided:

- `search_catalogs`: List all catalogs for a specific VCS
- `get_catalog_details`: Get catalog details by ID

### Example Usage

```bash
# List all catalogs for a specific VCS
enbuild search_catalogs --vcs "GITHUB"

# Get catalog details
enbuild get_catalog_details --id "catalog-id"

# Search for catalogs by name, type, and VCS
enbuild search_catalogs --name "terraform" --type "terraform" --vcs "GITHUB"
```

All tools return a consistent JSON response:

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

---

## Development

To add new tools, update the `registerTools` function in [`mcpenbuild.go`](mcpenbuild.go) and implement the corresponding handler.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Contributing

Contributions are welcome! Please open issues or pull requests on GitHub.

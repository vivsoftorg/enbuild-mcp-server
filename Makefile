.PHONY: build run clean fmt test

build:
	@echo "Building MCP server..."
	go build -o bin/mcp-server-enbuild mcpenbuild.go

run: build
	@echo "Running MCP server using mcphost..."
	/usr/local/bin/mcphost -m ollama:llama3.2:latest --config .vscode/mcphost-mcp.json

clean:
	@echo "Cleaning up..."
	rm -f bin/mcp-server-enbuild

fmt:
	@echo "Formatting Go code..."
	go fmt ./

test:
	@echo "Running tests..."
	go test ./...

.DEFAULT_GOAL := build

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	localenbuild "github.com/vivsoftorg/mcp-server-enbuild/pkg/enbuild"
)

// Define the server configuration
const (
	serverName        = "enbuild"
	serverDescription = "MCP Server for ENBUILD Platform"
	serverVersion     = "0.0.1"
)

// Configuration for the ENBUILD client
type enbuildConfig struct {
	token   string
	debug   bool
	baseURL string
}

func (ec *enbuildConfig) addFlags() {
	flag.StringVar(&ec.token, "token", "", "API token for ENBUILD")
	flag.BoolVar(&ec.debug, "debug", false, "Enable debug mode for the ENBUILD client")
	flag.StringVar(&ec.baseURL, "base-url", "", "Base URL for the ENBUILD API")
}

// CatalogResponse is a standardized response format for catalog data
type CatalogResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Count   int    `json:"count,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func newServer() *server.MCPServer {
	s := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	registerTools(s)
	return s
}

func registerTools(s *server.MCPServer) {
	// Register the list_catalogs tool
	s.AddTool(mcp.NewTool("list_catalogs",
		mcp.WithDescription("Lists all catalogs for a given VCS type (GITHUB or GITLAB)"),
		mcp.WithString("vcs",
			mcp.Description("VCS to filter by (GITHUB or GITLAB)"),
			mcp.Required(),
		),
		mcp.WithString("token",
			mcp.Description("API token to use"),
		),
	), listCatalogs)

	// Register the get_catalog_details tool
	s.AddTool(mcp.NewTool("get_catalog_details",
		mcp.WithDescription("Fetches details of all catalogs that match a specific catalog ID"),
		mcp.WithString("id",
			mcp.Description("ID of the catalog"),
			mcp.Required(),
		),
		mcp.WithString("token",
			mcp.Description("API token to use"),
		),
	), getCatalogDetails)

	// Register the search_catalogs tool
	s.AddTool(mcp.NewTool("search_catalogs",
		mcp.WithDescription("Search for catalogs using name, filtered by catalog type and VCS"),
		mcp.WithString("name",
			mcp.Description("Name to search for"),
			mcp.Required(),
		),
		mcp.WithString("type",
			mcp.Description("Type to filter by (e.g., terraform, ansible)"),
			mcp.Required(),
		),
		mcp.WithString("vcs",
			mcp.Description("VCS to filter by (GITHUB or GITLAB)"),
			mcp.Required(),
		),
		mcp.WithString("token",
			mcp.Description("API token to use"),
		),
	), searchCatalogs)
}

// Run starts the MCP server with the specified transport
func run(transport, addr string, logLevel slog.Level, ec enbuildConfig) error {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))
	s := newServer()

	switch transport {
	case "stdio":
		srv := server.NewStdioServer(s)
		slog.Info("Starting ENBUILD MCP server using stdio transport")
		return srv.Listen(context.Background(), os.Stdin, os.Stdout)
	case "sse":
		srv := server.NewSSEServer(s)
		slog.Info("Starting ENBUILD MCP server using SSE transport", "address", addr)
		if err := srv.Start(addr); err != nil {
			return fmt.Errorf("Server error: %v", err)
		}
	default:
		return fmt.Errorf(
			"Invalid transport type: %s. Must be 'stdio' or 'sse'",
			transport,
		)
	}
	return nil
}

func main() {
	var transport string
	flag.StringVar(&transport, "t", "stdio", "Transport type (stdio or sse)")
	flag.StringVar(
		&transport,
		"transport",
		"stdio",
		"Transport type (stdio or sse)",
	)
	addr := flag.String("sse-address", ":8080", "The host and port to start the SSE server on")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	
	var ec enbuildConfig
	ec.addFlags()
	
	flag.Parse()

	// Set environment variables from flags if provided
	if ec.token != "" {
		os.Setenv("ENBUILD_API_TOKEN", ec.token)
	}
	if ec.baseURL != "" {
		os.Setenv("ENBUILD_BASE_URL", ec.baseURL)
	}

	// Check for environment variables if flags not provided
	if ec.token == "" {
		ec.token = os.Getenv("ENBUILD_API_TOKEN")
	}
	if ec.baseURL == "" {
		ec.baseURL = os.Getenv("ENBUILD_BASE_URL")
	}

	if err := run(transport, *addr, parseLevel(*logLevel), ec); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func parseLevel(level string) slog.Level {
	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		return slog.LevelInfo
	}
	return l
}

// initializeClient creates a new ENBUILD client with the provided token
func initializeClient(token string) (*localenbuild.Client, error) {
	options := []localenbuild.ClientOption{}
	
	// Use token from parameter or environment variable
	if token != "" {
		options = append(options, localenbuild.WithAuthToken(token))
	} else if envToken := os.Getenv("ENBUILD_API_TOKEN"); envToken != "" {
		options = append(options, localenbuild.WithAuthToken(envToken))
	}
	
	// Use base URL from environment variable if available
	if baseURL := os.Getenv("ENBUILD_BASE_URL"); baseURL != "" {
		// Add base URL option if your client supports it
		// options = append(options, localenbuild.WithBaseURL(baseURL))
	}

	return localenbuild.NewClient(options...)
}

// Tool implementations
func listCatalogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	vcs, _ := request.Params.Arguments["vcs"].(string)
	token, _ := request.Params.Arguments["token"].(string)

	// Initialize ENBUILD SDK client
	client, err := initializeClient(token)
	if err != nil {
		return formatErrorResponse("Failed to initialize ENBUILD client", err)
	}

	// Filter catalogs by VCS
	catalogs, err := client.FilterCatalogsByVCS(vcs)
	if err != nil {
		return formatErrorResponse("Failed to list catalogs", err)
	}

	// Format the response
	response := CatalogResponse{
		Success: true,
		Count:   len(catalogs),
		Data:    catalogs,
		Message: fmt.Sprintf("Successfully retrieved %d catalogs for VCS: %s", len(catalogs), vcs),
	}

	return formatJSONResponse(response)
}

func getCatalogDetails(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, _ := request.Params.Arguments["id"].(string)
	token, _ := request.Params.Arguments["token"].(string)

	// Initialize ENBUILD SDK client
	client, err := initializeClient(token)
	if err != nil {
		return formatErrorResponse("Failed to initialize ENBUILD client", err)
	}

	// Get catalog details
	catalog, err := client.GetCatalog(id, nil)
	if err != nil {
		return formatErrorResponse("Failed to get catalog details", err)
	}

	// Format the response
	response := CatalogResponse{
		Success: true,
		Count:   1,
		Data:    catalog,
		Message: fmt.Sprintf("Successfully retrieved details for catalog ID: %s", id),
	}

	return formatJSONResponse(response)
}

func searchCatalogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, _ := request.Params.Arguments["name"].(string)
	catalogType, _ := request.Params.Arguments["type"].(string)
	vcs, _ := request.Params.Arguments["vcs"].(string)
	token, _ := request.Params.Arguments["token"].(string)

	// Initialize ENBUILD SDK client
	client, err := initializeClient(token)
	if err != nil {
		return formatErrorResponse("Failed to initialize ENBUILD client", err)
	}

	// Create options for filtering
	options := &localenbuild.CatalogListOptions{
		Name: name,
	}
	
	// Add type filter if provided
	if catalogType != "" {
		options.Type = catalogType
	}
	
	// Add VCS filter if provided
	if vcs != "" {
		options.VCS = vcs
	}

	// Search catalogs with filters
	catalogs, err := client.ListCatalogs(options)
	if err != nil {
		return formatErrorResponse("Failed to search catalogs", err)
	}

	// Format the response
	filterDesc := fmt.Sprintf("name: '%s'", name)
	if catalogType != "" {
		filterDesc += fmt.Sprintf(", type: '%s'", catalogType)
	}
	if vcs != "" {
		filterDesc += fmt.Sprintf(", vcs: '%s'", vcs)
	}
	
	response := CatalogResponse{
		Success: true,
		Count:   len(catalogs),
		Data:    catalogs,
		Message: fmt.Sprintf("Found %d catalogs matching filters: %s", len(catalogs), filterDesc),
	}

	return formatJSONResponse(response)
}

// Helper functions for consistent response formatting
func formatJSONResponse(response CatalogResponse) (*mcp.CallToolResult, error) {
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error formatting JSON response: %v", err)
	}
	
	return mcp.NewToolResultText(string(jsonData)), nil
}

func formatErrorResponse(message string, err error) (*mcp.CallToolResult, error) {
	response := CatalogResponse{
		Success: false,
		Message: fmt.Sprintf("%s: %v", message, err),
	}
	
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error formatting error response: %v", err)
	}
	
	return mcp.NewToolResultText(string(jsonData)), nil
}

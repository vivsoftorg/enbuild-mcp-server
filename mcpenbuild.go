package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

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
	// Register the list-catalogs tool
	s.AddTool(mcp.NewTool("list-catalogs",
		mcp.WithDescription("List all ENBUILD catalogs"),
		mcp.WithString("token",
			mcp.Description("API token to use"),
		),
	), listCatalogs)

	// Register the get-catalog tool
	s.AddTool(mcp.NewTool("get-catalog",
		mcp.WithDescription("Get details of a specific ENBUILD catalog"),
		mcp.WithString("id",
			mcp.Description("ID of the catalog"),
			mcp.Required(),
		),
		mcp.WithString("token",
			mcp.Description("API token to use"),
		),
	), getCatalog)

	// Register the search-catalogs tool
	s.AddTool(mcp.NewTool("search-catalogs",
		mcp.WithDescription("Search for ENBUILD catalogs by name"),
		mcp.WithString("name",
			mcp.Description("Name to search for"),
			mcp.Required(),
		),
		mcp.WithString("token",
			mcp.Description("API token to use"),
		),
	), searchCatalogs)

	// Register the filter-catalogs-by-type tool
	s.AddTool(mcp.NewTool("filter-catalogs-by-type",
		mcp.WithDescription("Filter ENBUILD catalogs by type"),
		mcp.WithString("type",
			mcp.Description("Type to filter by (e.g., terraform, ansible)"),
			mcp.Required(),
		),
		mcp.WithString("token",
			mcp.Description("API token to use"),
		),
	), filterCatalogsByType)

	// Register the filter-catalogs-by-vcs tool
	s.AddTool(mcp.NewTool("filter-catalogs-by-vcs",
		mcp.WithDescription("Filter ENBUILD catalogs by VCS"),
		mcp.WithString("vcs",
			mcp.Description("VCS to filter by (e.g., github, gitlab)"),
			mcp.Required(),
		),
		mcp.WithString("token",
			mcp.Description("API token to use"),
		),
	), filterCatalogsByVCS)
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

// Tool implementations
func listCatalogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	token, _ := request.Params.Arguments["token"].(string)

	// Initialize ENBUILD SDK client
	options := []localenbuild.ClientOption{}
	if token != "" {
		options = append(options, localenbuild.WithAuthToken(token))
	}

	client, err := localenbuild.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ENBUILD client: %v", err)
	}

	// Get catalogs
	catalogs, err := client.ListCatalogs()
	if err != nil {
		return nil, fmt.Errorf("failed to list catalogs: %v", err)
	}

	// Format the response
	var result strings.Builder
	result.WriteString("Catalogs:\n")
	for _, catalog := range catalogs {
		result.WriteString(fmt.Sprintf("- ID: %v, Name: %s, Type: %s, VCS: %s, Slug: %s\n",
			catalog.ID, catalog.Name, catalog.Type, catalog.VCS, catalog.Slug))
	}

	return mcp.NewToolResultText(result.String()), nil
}

func getCatalog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, _ := request.Params.Arguments["id"].(string)
	token, _ := request.Params.Arguments["token"].(string)

	// Initialize ENBUILD SDK client
	options := []localenbuild.ClientOption{}
	if token != "" {
		options = append(options, localenbuild.WithAuthToken(token))
	}

	client, err := localenbuild.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ENBUILD client: %v", err)
	}

	// Get catalog details
	catalog, err := client.GetCatalog(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog details: %v", err)
	}

	// Format the response
	var result strings.Builder
	result.WriteString("Catalog Details:\n")
	result.WriteString(fmt.Sprintf("- ID: %v\n", catalog.ID))
	result.WriteString(fmt.Sprintf("- Name: %s\n", catalog.Name))
	result.WriteString(fmt.Sprintf("- Description: %s\n", catalog.Description))
	result.WriteString(fmt.Sprintf("- Type: %s\n", catalog.Type))
	result.WriteString(fmt.Sprintf("- VCS: %s\n", catalog.VCS))
	result.WriteString(fmt.Sprintf("- Slug: %s\n", catalog.Slug))
	result.WriteString(fmt.Sprintf("- Version: %s\n", catalog.Version))

	return mcp.NewToolResultText(result.String()), nil
}

func searchCatalogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, _ := request.Params.Arguments["name"].(string)
	token, _ := request.Params.Arguments["token"].(string)

	// Initialize ENBUILD SDK client
	options := []localenbuild.ClientOption{}
	if token != "" {
		options = append(options, localenbuild.WithAuthToken(token))
	}

	client, err := localenbuild.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ENBUILD client: %v", err)
	}

	// Search catalogs
	catalogs, err := client.SearchCatalogs(name)
	if err != nil {
		return nil, fmt.Errorf("failed to search catalogs: %v", err)
	}

	// Format the response
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Search Results for '%s':\n", name))
	if len(catalogs) == 0 {
		result.WriteString("No catalogs found matching the search criteria.\n")
	} else {
		for _, catalog := range catalogs {
			result.WriteString(fmt.Sprintf("- ID: %v, Name: %s, Type: %s, VCS: %s, Slug: %s\n",
				catalog.ID, catalog.Name, catalog.Type, catalog.VCS, catalog.Slug))
		}
	}

	return mcp.NewToolResultText(result.String()), nil
}

func filterCatalogsByType(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	catalogType, _ := request.Params.Arguments["type"].(string)
	token, _ := request.Params.Arguments["token"].(string)

	// Initialize ENBUILD SDK client
	options := []localenbuild.ClientOption{}
	if token != "" {
		options = append(options, localenbuild.WithAuthToken(token))
	}

	client, err := localenbuild.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ENBUILD client: %v", err)
	}

	// Filter catalogs by type
	catalogs, err := client.FilterCatalogsByType(catalogType)
	if err != nil {
		return nil, fmt.Errorf("failed to filter catalogs: %v", err)
	}

	// Format the response
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Catalogs of type '%s':\n", catalogType))
	if len(catalogs) == 0 {
		result.WriteString("No catalogs found matching the filter criteria.\n")
	} else {
		for _, catalog := range catalogs {
			result.WriteString(fmt.Sprintf("- ID: %v, Name: %s, Type: %s, VCS: %s, Slug: %s\n",
				catalog.ID, catalog.Name, catalog.Type, catalog.VCS, catalog.Slug))
		}
	}

	return mcp.NewToolResultText(result.String()), nil
}

func filterCatalogsByVCS(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	vcs, _ := request.Params.Arguments["vcs"].(string)
	token, _ := request.Params.Arguments["token"].(string)

	// Initialize ENBUILD SDK client
	options := []localenbuild.ClientOption{}
	if token != "" {
		options = append(options, localenbuild.WithAuthToken(token))
	}

	client, err := localenbuild.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ENBUILD client: %v", err)
	}

	// Filter catalogs by VCS
	catalogs, err := client.FilterCatalogsByVCS(vcs)
	if err != nil {
		return nil, fmt.Errorf("failed to filter catalogs: %v", err)
	}

	// Format the response
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Catalogs from VCS '%s':\n", vcs))
	if len(catalogs) == 0 {
		result.WriteString("No catalogs found matching the filter criteria.\n")
	} else {
		for _, catalog := range catalogs {
			result.WriteString(fmt.Sprintf("- ID: %v, Name: %s, Type: %s, VCS: %s, Slug: %s\n",
				catalog.ID, catalog.Name, catalog.Type, catalog.VCS, catalog.Slug))
		}
	}

	return mcp.NewToolResultText(result.String()), nil
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/olekukonko/tablewriter"
	"github.com/vivsoftorg/enbuild-sdk-go/pkg/enbuild"
	localenbuild "github.com/vivsoftorg/mcp-server-enbuild/pkg/enbuild"
)

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
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Count   int         `json:"count,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Table   string      `json:"table,omitempty"` // Added for table representation
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
	s.AddTool(mcp.NewTool("list_catalogs",
		mcp.WithDescription("Lists all catalogs for a given VCS type (GITHUB or GITLAB). Use this tool when the user wants to list or view all catalogs for a specific version control system (VCS)."),
		mcp.WithString("vcs",
			mcp.Description("VCS to filter by (GITHUB or GITLAB)"),
			mcp.Required(),
		),
		mcp.WithString("token",
			mcp.Description("API token to use"),
		),
	), listCatalogs)

	s.AddTool(mcp.NewTool("get_catalog_details",
		mcp.WithDescription("Fetches details of all catalogs that match a specific catalog ID. Use this tool when the user wants to get detailed information about a specific catalog."),
		mcp.WithString("id",
			mcp.Description("ID of the catalog"),
			mcp.Required(),
		),
		mcp.WithString("token",
			mcp.Description("API token to use"),
		),
	), getCatalogDetails)

	s.AddTool(mcp.NewTool("search_catalogs",
		mcp.WithDescription("Search for catalogs using name, filtered by catalog type and VCS. Use this tool when the user wants to search for catalogs by name with specific filters."),
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

func run(transport, addr string, logLevel string, ec enbuildConfig) error {
	log.SetFlags(0)
	log.Printf("[INFO] Starting ENBUILD MCP server with transport: %s", transport)

	s := newServer()

	switch transport {
	case "stdio":
		srv := server.NewStdioServer(s)
		log.Println("Starting ENBUILD MCP server using stdio transport")
		return srv.Listen(context.Background(), os.Stdin, os.Stdout)
	case "sse":
		srv := server.NewSSEServer(s)
		log.Printf("Starting ENBUILD MCP server using SSE transport on address: %s", addr)
		if err := srv.Start(addr); err != nil {
			return fmt.Errorf("server error: %v", err)
		}
	default:
		return fmt.Errorf("invalid transport type: %s. Must be 'stdio' or 'sse'", transport)
	}
	return nil
}

func main() {
	var transport string
	flag.StringVar(&transport, "transport", "stdio", "Transport type (stdio or sse)")
	addr := flag.String("sse-address", ":8080", "The host and port to start the SSE server on")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")

	var ec enbuildConfig
	ec.addFlags()

	flag.Parse()

	// Check for token and base URL from flags or environment variables
	if ec.token == "" {
		ec.token = os.Getenv("ENBUILD_API_TOKEN")
		if ec.token == "" {
			log.Fatalf("Error: ENBUILD API token is required. Provide it via --token flag or ENBUILD_API_TOKEN environment variable")
		}
	}

	if ec.baseURL == "" {
		ec.baseURL = os.Getenv("ENBUILD_BASE_URL")
		if ec.baseURL == "" {
			log.Fatalf("Error: ENBUILD base URL is required. Provide it via --base-url flag or ENBUILD_BASE_URL environment variable")
		}
	}

	// Set environment variables from flags for consistent access
	os.Setenv("ENBUILD_API_TOKEN", ec.token)
	os.Setenv("ENBUILD_BASE_URL", ec.baseURL)

	if err := run(transport, *addr, *logLevel, ec); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// initializeClient creates a new ENBUILD client with the provided token
func initializeClient(token string) (*localenbuild.Client, error) {
	options := []localenbuild.ClientOption{}

	// Use token from parameter or environment variable
	if token != "" {
		options = append(options, localenbuild.WithAuthToken(token))
	} else if envToken := os.Getenv("ENBUILD_API_TOKEN"); envToken != "" {
		options = append(options, localenbuild.WithAuthToken(envToken))
	} else {
		return nil, fmt.Errorf("API token is required but not provided")
	}

	// Use base URL from environment variable
	baseURL := os.Getenv("ENBUILD_BASE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required but not provided")
	}

	options = append(options, localenbuild.WithBaseURL(baseURL))

	return localenbuild.NewClient(options...)
}

// Helper function to get token from request
func getToken(request mcp.CallToolRequest) string {
	token, _ := request.Params.Arguments["token"].(string)
	return token
}

// Helper function to get search parameters from request
func getSearchParams(request mcp.CallToolRequest) (string, string, string) {
	name, _ := request.Params.Arguments["name"].(string)
	catalogType, _ := request.Params.Arguments["type"].(string)
	vcs, _ := request.Params.Arguments["vcs"].(string)
	return name, catalogType, vcs
}

// Tool implementations
func listCatalogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	vcs, _ := request.Params.Arguments["vcs"].(string)

	// Validate VCS parameter
	if vcs == "" {
		return formatErrorResponse("Missing required parameter", fmt.Errorf("VCS parameter is required (GITHUB or GITLAB)"))
	}

	// Normalize VCS value to uppercase
	vcs = strings.ToUpper(vcs)

	// Validate VCS value
	if vcs != "GITHUB" && vcs != "GITLAB" {
		return formatErrorResponse("Invalid VCS value", fmt.Errorf("VCS must be either GITHUB or GITLAB"))
	}

	// Initialize ENBUILD SDK client
	client, err := initializeClient(getToken(request))
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
	id, ok := request.Params.Arguments["id"].(string)
	if !ok || id == "" {
		return formatErrorResponse("Missing required parameter", fmt.Errorf("catalog ID is required"))
	}

	// Initialize ENBUILD SDK client
	client, err := initializeClient(getToken(request))
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
	name, catalogType, vcs := getSearchParams(request)

	// Validate required parameters
	if name == "" {
		return formatErrorResponse("Missing required parameter", fmt.Errorf("name parameter is required"))
	}

	if catalogType == "" {
		return formatErrorResponse("Missing required parameter", fmt.Errorf("type parameter is required"))
	}

	if vcs == "" {
		return formatErrorResponse("Missing required parameter", fmt.Errorf("VCS parameter is required (GITHUB or GITLAB)"))
	}

	// Normalize VCS value to uppercase
	vcs = strings.ToUpper(vcs)

	// Validate VCS value
	if vcs != "GITHUB" && vcs != "GITLAB" {
		return formatErrorResponse("Invalid VCS value", fmt.Errorf("VCS must be either GITHUB or GITLAB"))
	}

	// Initialize ENBUILD SDK client
	client, err := initializeClient(getToken(request))
	if err != nil {
		return formatErrorResponse("Failed to initialize ENBUILD client", err)
	}

	// Create options for filtering
	options := &localenbuild.CatalogListOptions{
		Name: name,
		Type: catalogType,
		VCS:  vcs,
	}

	// Search catalogs with filters
	catalogs, err := client.ListCatalogs(options)
	if err != nil {
		return formatErrorResponse("Failed to search catalogs", err)
	}

	// Format the response
	filterDesc := fmt.Sprintf("name: '%s', type: '%s', vcs: '%s'", name, catalogType, vcs)

	response := CatalogResponse{
		Success: true,
		Count:   len(catalogs),
		Data:    catalogs,
		Message: fmt.Sprintf("Found %d catalogs matching filters: %s", len(catalogs), filterDesc),
	}

	return formatJSONResponse(response)
}

// formatJSONResponse formats the response as JSON and includes a table representation
func formatJSONResponse(response CatalogResponse) (*mcp.CallToolResult, error) {
	// Generate table representation if there's data
	if response.Success && response.Count > 0 {
		tableStr, err := generateTableOutput(response.Data)
		if err == nil {
			response.Table = tableStr
		}
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error formatting JSON response: %v", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// formatErrorResponse formats an error response
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

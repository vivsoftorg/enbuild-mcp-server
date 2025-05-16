package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	localenbuild "github.com/vivsoftorg/mcp-server-enbuild/pkg/enbuild"
)

const (
	serverName        = "enbuild"
	serverDescription = "MCP Server for ENBUILD Platform"
	serverVersion     = "0.0.1"
)

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

type CatalogResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Count   int         `json:"count,omitempty"`
	Data    interface{} `json:"data,omitempty"`
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
		mcp.WithDescription("Lists all Enbuild catalogs for a given VCS (GITHUB or GITLAB); use when the user wants to view or count catalogs by VCS type."),
		mcp.WithString("vcs", mcp.Description("VCS to filter by (github or gitlab)"), mcp.Required()),
	), listCatalogs)

	s.AddTool(mcp.NewTool("get_catalog_details",
		mcp.WithDescription("Fetches detailed info for catalogs matching a specific catalog ID; use after the user selects a catalog from a list or search."),
		mcp.WithString("id", mcp.Description("ID of the catalog"), mcp.Required()),
	), getCatalogDetails)

	s.AddTool(mcp.NewTool("search_catalogs",
		mcp.WithDescription("Searches Enbuild catalogs by name, filtered by type (e.g., helm, terraform) and VCS (GITHUB/GITLAB); use for keyword-based catalog queries."),
		mcp.WithString("name", mcp.Description("Keyword to search in names")),
		mcp.WithString("type", mcp.Description("Type to filter by (e.g., terraform, ansible)")),
		mcp.WithString("vcs", mcp.Description("VCS to filter by (github or gitlab)"), mcp.Required()),
	), searchCatalogs)
}

func run(transport, addr, logLevel string, ec enbuildConfig) error {
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

	envSetup(&ec)

	if err := run(transport, *addr, *logLevel, ec); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func envSetup(ec *enbuildConfig) {
	if ec.token == "" {
		ec.token = os.Getenv("ENBUILD_API_TOKEN")
		fatalIfEmpty(ec.token, "ENBUILD API token is required. Use --token flag or ENBUILD_API_TOKEN environment variable")
	}

	if ec.baseURL == "" {
		ec.baseURL = os.Getenv("ENBUILD_BASE_URL")
		fatalIfEmpty(ec.baseURL, "ENBUILD base URL is required. Use --base-url flag or ENBUILD_BASE_URL environment variable")
	}

	os.Setenv("ENBUILD_API_TOKEN", ec.token)
	os.Setenv("ENBUILD_BASE_URL", ec.baseURL)
}

func fatalIfEmpty(value, message string) {
	if value == "" {
		log.Fatal(message)
	}
}

func initializeClient(token string) (*localenbuild.Client, error) {
	options := []localenbuild.ClientOption{localenbuild.WithAuthToken(token)}

	baseURL := os.Getenv("ENBUILD_BASE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required but not provided")
	}
	options = append(options, localenbuild.WithBaseURL(baseURL))

	return localenbuild.NewClient(options...)
}

func listCatalogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	vcs := request.Params.Arguments["vcs"].(string)
	client, err := initializeClient(getToken(request))
	if err != nil {
		return formatErrorResponse("Failed to initialize ENBUILD client", err)
	}

	catalogs, err := client.FilterCatalogsByVCS(vcs)
	if err != nil {
		return formatErrorResponse("Failed to list catalogs", err)
	}

	response := CatalogResponse{
		Success: true,
		Count:   len(catalogs),
		Data:    catalogs,
		Message: fmt.Sprintf("Successfully retrieved %d catalogs for VCS: %s", len(catalogs), vcs),
	}

	return formatJSONResponse(response)
}

func getCatalogDetails(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := request.Params.Arguments["id"].(string)
	client, err := initializeClient(getToken(request))
	if err != nil {
		return formatErrorResponse("Failed to initialize ENBUILD client", err)
	}

	catalog, err := client.GetCatalog(id, nil)
	if err != nil {
		return formatErrorResponse("Failed to get catalog details", err)
	}

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
	client, err := initializeClient(getToken(request))
	if err != nil {
		return formatErrorResponse("Failed to initialize ENBUILD client", err)
	}

	options := &localenbuild.CatalogListOptions{Name: name, Type: catalogType, VCS: vcs}

	catalogs, err := client.ListCatalogs(options)
	if err != nil {
		return formatErrorResponse("Failed to search catalogs", err)
	}

	filterDesc := createFilterDescription(name, catalogType, vcs)

	response := CatalogResponse{
		Success: true,
		Count:   len(catalogs),
		Data:    catalogs,
		Message: fmt.Sprintf("Found %d catalogs matching filters: %s", len(catalogs), filterDesc),
	}

	return formatJSONResponse(response)
}

func getSearchParams(request mcp.CallToolRequest) (string, string, string) {
	name, _ := request.Params.Arguments["name"].(string)
	catalogType, _ := request.Params.Arguments["type"].(string)
	vcs, _ := request.Params.Arguments["vcs"].(string)
	return name, catalogType, vcs
}

func getToken(request mcp.CallToolRequest) string {
	token, _ := request.Params.Arguments["token"].(string)
	return token
}

func createFilterDescription(name, catalogType, vcs string) string {
	desc := fmt.Sprintf("name: '%s'", name)
	if catalogType != "" {
		desc += fmt.Sprintf(", type: '%s'", catalogType)
	}
	if vcs != "" {
		desc += fmt.Sprintf(", vcs: '%s'", vcs)
	}
	return desc
}

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

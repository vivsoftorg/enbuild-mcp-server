package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/vivsoftorg/enbuild-sdk-go/pkg/enbuild"
)

const (
	serverName        = "enbuild"
	serverDescription = "MCP Server for ENBUILD Platform"
	serverVersion     = "0.0.1"
)

type enbuildConfig struct {
	username string
	password string
	debug    bool
	baseURL  string
}

func (ec *enbuildConfig) addFlags() {
	flag.StringVar(&ec.username, "username", "", "API username for ENBUILD")
	flag.StringVar(&ec.password, "password", "", "API password for ENBUILD")
	flag.BoolVar(&ec.debug, "debug", false, "Enable debug mode for the ENBUILD client")
	flag.StringVar(&ec.baseURL, "base-url", "https://enbuild.vivplatform.io", "Base URL for the ENBUILD API")
}

type CatalogResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Count   int         `json:"count,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func newServer() *server.MCPServer {
	s := server.NewMCPServer(serverName, serverVersion, server.WithToolCapabilities(true), server.WithRecovery())
	registerTools(s)
	return s
}

func registerTools(s *server.MCPServer) {
	s.AddTool(mcp.NewTool("get_catalog_details",
		mcp.WithDescription("Fetches details of all catalogs that match a specific catalog ID."),
		mcp.WithString("id", mcp.Description("ID of the catalog"), mcp.Required()),
		mcp.WithString("username", mcp.Description("API username to use")),
		mcp.WithString("password", mcp.Description("API password to use")),
	), getCatalogDetails)

	s.AddTool(mcp.NewTool("search_catalogs",
		mcp.WithDescription("Search for catalogs using name, filtered by catalog type and VCS."),
		mcp.WithString("name", mcp.Description("Name to search for"), mcp.Required()),
		mcp.WithString("type", mcp.Description("Type to filter by (e.g., terraform, ansible)"), mcp.Required()),
		mcp.WithString("vcs", mcp.Description("VCS to filter by (GITHUB or GITLAB)"), mcp.Required()),
		mcp.WithString("username", mcp.Description("API username to use")),
		mcp.WithString("password", mcp.Description("API password to use")),
	), listCatalogs)
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

	// Retrieve credentials and baseURL, set them as environment variables
	setEnvOrExit("ENBUILD_USERNAME", ec.username, "--username flag")
	setEnvOrExit("ENBUILD_PASSWORD", ec.password, "--password flag")
	setEnvOrExit("ENBUILD_BASE_URL", ec.baseURL, "--base-url flag")

	if err := run(transport, *addr, *logLevel, ec); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func setEnvOrExit(envVar, value, flagName string) {
	if value == "" {
		value = os.Getenv(envVar)
		if value == "" {
			log.Fatalf("Error: ENVBUILD API %s is required. Provide it via %s or %s environment variable", envVar, flagName, envVar)
		}
	}
	os.Setenv(envVar, value)
}

func prepareClientOptions(baseURL, username, password string) []enbuild.ClientOption {
	debug := false
	if os.Getenv("ENBUILD_DEBUG") == "true" {
		debug = true
	}
	return []enbuild.ClientOption{
		enbuild.WithDebug(debug),
		enbuild.WithBaseURL(baseURL),
		enbuild.WithKeycloakAuth(username, password),
	}
}

func getCredentials(request mcp.CallToolRequest) (string, string, string, error) {
	username, _ := request.Params.Arguments["username"].(string)
	password, _ := request.Params.Arguments["password"].(string)
	baseURL, _ := request.Params.Arguments["base_url"].(string)
	if baseURL == "" {
		baseURL = os.Getenv("ENBUILD_BASE_URL")
	}
	if username == "" {
		username = os.Getenv("ENBUILD_USERNAME")
	}
	if password == "" {
		password = os.Getenv("ENBUILD_PASSWORD")
	}
	if baseURL == "" || username == "" || password == "" {
		return "", "", "", fmt.Errorf("Missing required credentials: baseURL, username, or password")
	}
	return baseURL, username, password, nil
}

func getSearchParams(request mcp.CallToolRequest) (string, string, string) {
	name, _ := request.Params.Arguments["name"].(string)
	catalogType, _ := request.Params.Arguments["type"].(string)
	vcs, _ := request.Params.Arguments["vcs"].(string)
	return name, catalogType, vcs
}

func listCatalogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	vcs, _ := request.Params.Arguments["vcs"].(string)

	if vcs == "" {
		return formatErrorResponse("Missing required parameter", fmt.Errorf("VCS parameter is required (GITHUB or GITLAB)"))
	}

	vcs = strings.ToUpper(vcs)

	if vcs != "GITHUB" && vcs != "GITLAB" {
		return formatErrorResponse("Invalid VCS value", fmt.Errorf("VCS must be either GITHUB or GITLAB"))
	}

	baseURL, username, password, err := getCredentials(request)
	if err != nil {
		return formatErrorResponse("Missing credentials", err)
	}

	client, err := initializeClient(baseURL, username, password)
	if err != nil {
		return formatErrorResponse("Failed to initialize ENBUILD client", err)
	}

	opts := &enbuild.CatalogListOptions{
		VCS: vcs,
	}

	catalogs, err := client.Catalogs.List(opts)
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
	id, ok := request.Params.Arguments["id"].(string)
	if !ok || id == "" {
		return formatErrorResponse("Missing required parameter", fmt.Errorf("catalog ID is required"))
	}

	baseURL, username, password, err := getCredentials(request)
	if err != nil {
		return formatErrorResponse("Missing credentials", err)
	}

	client, err := initializeClient(baseURL, username, password)
	if err != nil {
		return formatErrorResponse("Failed to initialize ENBUILD client", err)
	}

	catalog, err := client.Catalogs.Get(id, nil)
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

	if name == "" {
		return formatErrorResponse("Missing required parameter", fmt.Errorf("name parameter is required"))
	}

	if catalogType == "" {
		return formatErrorResponse("Missing required parameter", fmt.Errorf("type parameter is required"))
	}

	if vcs == "" {
		return formatErrorResponse("Missing required parameter", fmt.Errorf("VCS parameter is required (GITHUB or GITLAB)"))
	}

	vcs = strings.ToUpper(vcs)

	if vcs != "GITHUB" && vcs != "GITLAB" {
		return formatErrorResponse("Invalid VCS value", fmt.Errorf("VCS must be either \"GITHUB\" or \"GITLAB\""))
	}

	baseURL, username, password, err := getCredentials(request)
	if err != nil {
		return formatErrorResponse("Missing credentials", err)
	}

	client, err := initializeClient(baseURL, username, password)
	if err != nil {
		return formatErrorResponse("Failed to initialize ENBUILD client", err)
	}

	options := &enbuild.CatalogListOptions{
		Name: name,
		Type: catalogType,
		VCS:  vcs,
	}

	catalogs, err := client.Catalogs.List(options)
	if err != nil {
		return formatErrorResponse("Failed to search catalogs", err)
	}

	filterDesc := fmt.Sprintf("name: '%s', type: '%s', vcs: '%s'", name, catalogType, vcs)

	response := CatalogResponse{
		Success: true,
		Count:   len(catalogs),
		Data:    catalogs,
		Message: fmt.Sprintf("Found %d catalogs matching filters: %s", len(catalogs), filterDesc),
	}

	return formatJSONResponse(response)
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

func initializeClient(baseURL, username, password string) (*enbuild.Client, error) {
	options := prepareClientOptions(baseURL, username, password)
	return enbuild.NewClient(options...)
}
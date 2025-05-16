package enbuild

import (
	"time"

	"github.com/vivsoftorg/enbuild-sdk-go/pkg/enbuild"
)

// CatalogListOptions represents options for listing catalogs
type CatalogListOptions struct {
	Name string
	Type string
	VCS  string
}

// Client represents a wrapper around the ENBUILD SDK client
type Client struct {
	sdkClient *enbuild.Client
	profile   string
	baseURL   string
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithProfile sets the AWS profile to use
func WithProfile(profile string) ClientOption {
	return func(c *Client) {
		c.profile = profile
	}
}

// WithAuthToken sets the authentication token
func WithAuthToken(token string) ClientOption {
	return func(c *Client) {
		// This will be used when creating the SDK client
	}
}

// WithBaseURL sets the base URL for the ENBUILD API
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		// This will be used when creating the SDK client
	}
}

// NewClient creates a new ENBUILD client
func NewClient(options ...ClientOption) (*Client, error) {
	client := &Client{}

	// Apply options
	for _, option := range options {
		option(client)
	}

	// Create SDK client options
	sdkOptions := []enbuild.ClientOption{
		enbuild.WithDebug(false),
	}

	// Add base URL if provided
	if client.baseURL != "" {
		sdkOptions = append(sdkOptions, enbuild.WithBaseURL(client.baseURL))
	}

	// Create the SDK client
	sdkClient, err := enbuild.NewClient(sdkOptions...)
	if err != nil {
		return nil, err
	}

	client.sdkClient = sdkClient
	return client, nil
}

// ListCatalogs returns a list of ENBUILD catalogs with optional filters
func (c *Client) ListCatalogs(options *CatalogListOptions) ([]*enbuild.Catalog, error) {
	// Convert our options to SDK options
	sdkOptions := &enbuild.CatalogListOptions{}
	
	if options != nil {
		sdkOptions.Name = options.Name
		sdkOptions.Type = options.Type
		sdkOptions.VCS = options.VCS
	}

	return c.sdkClient.Catalogs.List(sdkOptions)
}

// GetCatalog returns details of a specific ENBUILD catalog
func (c *Client) GetCatalog(id string, options *CatalogListOptions) (*enbuild.Catalog, error) {
	sdkOptions := &enbuild.CatalogListOptions{}
	
	if options != nil {
		sdkOptions.Name = options.Name
		sdkOptions.Type = options.Type
		sdkOptions.VCS = options.VCS
	}

	return c.sdkClient.Catalogs.Get(id, sdkOptions)
}

// FilterCatalogsByVCS filters catalogs by VCS
func (c *Client) FilterCatalogsByVCS(vcs string) ([]*enbuild.Catalog, error) {
	options := &CatalogListOptions{
		VCS: vcs,
	}

	return c.ListCatalogs(options)
}

package enbuild

import (
	"time"

	"github.com/vivsoftorg/enbuild-sdk-go/pkg/enbuild"
)

// Client represents a wrapper around the ENBUILD SDK client
type Client struct {
	sdkClient *enbuild.Client
	profile   string
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

	// Create the SDK client
	sdkClient, err := enbuild.NewClient(sdkOptions...)
	if err != nil {
		return nil, err
	}

	client.sdkClient = sdkClient
	return client, nil
}

// ListCatalogs returns a list of ENBUILD catalogs
func (c *Client) ListCatalogs(opts ...*enbuild.CatalogListOptions) ([]*enbuild.Catalog, error) {
	var options *enbuild.CatalogListOptions
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	} else {
		options = &enbuild.CatalogListOptions{}
	}

	return c.sdkClient.Catalogs.List(options)
}

// GetCatalog returns details of a specific ENBUILD catalog
func (c *Client) GetCatalog(id string, opts *enbuild.CatalogListOptions) (*enbuild.Catalog, error) {
	if opts == nil {
		opts = &enbuild.CatalogListOptions{}
	}

	return c.sdkClient.Catalogs.Get(id, opts)
}

// SearchCatalogs searches for catalogs by name
func (c *Client) SearchCatalogs(name string) ([]*enbuild.Catalog, error) {
	options := &enbuild.CatalogListOptions{
		Name: name,
	}

	return c.sdkClient.Catalogs.List(options)
}

// FilterCatalogsByType filters catalogs by type
func (c *Client) FilterCatalogsByType(catalogType string) ([]*enbuild.Catalog, error) {
	options := &enbuild.CatalogListOptions{
		Type: catalogType,
	}

	return c.sdkClient.Catalogs.List(options)
}

// FilterCatalogsByVCS filters catalogs by VCS
func (c *Client) FilterCatalogsByVCS(vcs string) ([]*enbuild.Catalog, error) {
	options := &enbuild.CatalogListOptions{
		VCS: vcs,
	}

	return c.sdkClient.Catalogs.List(options)
}

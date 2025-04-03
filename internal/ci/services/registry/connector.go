package registry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RegistryType represents the type of Docker registry
type RegistryType string

const (
	// DockerHub represents Docker Hub registry
	DockerHub RegistryType = "dockerhub"
	// Harbor represents Harbor registry
	Harbor RegistryType = "harbor"
	// Generic represents a generic Docker registry
	Generic RegistryType = "generic"
)

// RegistryConnector handles connections to Docker registries
type RegistryConnector struct {
	client *http.Client
}

// RegistryConfig contains configuration for connecting to a Docker registry
type RegistryConfig struct {
	Type     RegistryType `json:"type"`
	URL      string       `json:"url"`
	Username string       `json:"username"`
	Password string       `json:"password"`
	Insecure bool         `json:"insecure"`
}

// RegistryAuthResponse represents the response from a Docker registry authentication
type RegistryAuthResponse struct {
	Token       string    `json:"token"`
	AccessToken string    `json:"access_token"`
	ExpiresIn   int       `json:"expires_in"`
	IssuedAt    time.Time `json:"issued_at"`
}

// NewRegistryConnector creates a new RegistryConnector instance
func NewRegistryConnector() *RegistryConnector {
	return &RegistryConnector{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetRegistryURL returns the appropriate URL for the registry type
func GetRegistryURL(config RegistryConfig) string {
	switch config.Type {
	case DockerHub:
		return "https://registry.hub.docker.com"
	case Harbor:
		return config.URL
	case Generic:
		return config.URL
	default:
		return config.URL
	}
}

// Authenticate authenticates with the Docker registry
func (c *RegistryConnector) Authenticate(ctx context.Context, config RegistryConfig) (string, error) {
	switch config.Type {
	case DockerHub:
		return c.authenticateDockerHub(ctx, config)
	case Harbor:
		return c.authenticateHarbor(ctx, config)
	case Generic:
		return c.authenticateGeneric(ctx, config)
	default:
		return c.authenticateGeneric(ctx, config)
	}
}

// authenticateDockerHub authenticates with Docker Hub
func (c *RegistryConnector) authenticateDockerHub(ctx context.Context, config RegistryConfig) (string, error) {
	url := "https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/alpine:pull,push"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add basic auth header
	auth := base64.StdEncoding.EncodeToString([]byte(config.Username + ":" + config.Password))
	req.Header.Add("Authorization", "Basic "+auth)
	
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate with Docker Hub: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to authenticate with Docker Hub: %s", body)
	}
	
	var authResp RegistryAuthResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode auth response: %w", err)
	}
	
	// Use token or access_token, whichever is available
	token := authResp.Token
	if token == "" {
		token = authResp.AccessToken
	}
	
	return token, nil
}

// authenticateHarbor authenticates with Harbor
func (c *RegistryConnector) authenticateHarbor(ctx context.Context, config RegistryConfig) (string, error) {
	// Harbor uses basic auth for API calls
	auth := base64.StdEncoding.EncodeToString([]byte(config.Username + ":" + config.Password))
	return auth, nil
}

// authenticateGeneric authenticates with a generic Docker registry
func (c *RegistryConnector) authenticateGeneric(ctx context.Context, config RegistryConfig) (string, error) {
	// Most registries use basic auth
	auth := base64.StdEncoding.EncodeToString([]byte(config.Username + ":" + config.Password))
	return auth, nil
}

// PushImage pushes an image to the Docker registry
func (c *RegistryConnector) PushImage(ctx context.Context, config RegistryConfig, imageName string) error {
	// Get authentication token
	token, err := c.Authenticate(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	
	// Construct the URL for the push operation
	registryURL := GetRegistryURL(config)
	url := fmt.Sprintf("%s/v2/%s/manifests/latest", registryURL, imageName)
	
	req, err := http.NewRequestWithContext(ctx, "PUT", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add auth header
	if config.Type == DockerHub {
		req.Header.Add("Authorization", "Bearer "+token)
	} else {
		req.Header.Add("Authorization", "Basic "+token)
	}
	
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to push image: %s", body)
	}
	
	return nil
}

// ListRepositories lists repositories in the Docker registry
func (c *RegistryConnector) ListRepositories(ctx context.Context, config RegistryConfig) ([]string, error) {
	// Get authentication token
	token, err := c.Authenticate(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}
	
	// Construct the URL for the list operation
	registryURL := GetRegistryURL(config)
	url := fmt.Sprintf("%s/v2/_catalog", registryURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add auth header
	if config.Type == DockerHub {
		req.Header.Add("Authorization", "Bearer "+token)
	} else {
		req.Header.Add("Authorization", "Basic "+token)
	}
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list repositories: %s", body)
	}
	
	var result struct {
		Repositories []string `json:"repositories"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result.Repositories, nil
}

// ListTags lists tags for a repository in the Docker registry
func (c *RegistryConnector) ListTags(ctx context.Context, config RegistryConfig, repository string) ([]string, error) {
	// Get authentication token
	token, err := c.Authenticate(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}
	
	// Construct the URL for the list operation
	registryURL := GetRegistryURL(config)
	url := fmt.Sprintf("%s/v2/%s/tags/list", registryURL, repository)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add auth header
	if config.Type == DockerHub {
		req.Header.Add("Authorization", "Bearer "+token)
	} else {
		req.Header.Add("Authorization", "Basic "+token)
	}
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list tags: %s", body)
	}
	
	var result struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result.Tags, nil
}

// DeleteTag deletes a tag from a repository in the Docker registry
func (c *RegistryConnector) DeleteTag(ctx context.Context, config RegistryConfig, repository, tag string) error {
	// Get authentication token
	token, err := c.Authenticate(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	
	// Construct the URL for the delete operation
	registryURL := GetRegistryURL(config)
	url := fmt.Sprintf("%s/v2/%s/manifests/%s", registryURL, repository, tag)
	
	// First, we need to get the digest for the tag
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add auth header
	if config.Type == DockerHub {
		req.Header.Add("Authorization", "Bearer "+token)
	} else {
		req.Header.Add("Authorization", "Basic "+token)
	}
	
	// Add accept header for manifest v2
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get manifest: %w", err)
	}
	
	digest := resp.Header.Get("Docker-Content-Digest")
	resp.Body.Close()
	
	if digest == "" {
		return fmt.Errorf("failed to get digest for tag %s", tag)
	}
	
	// Now delete the tag using the digest
	url = fmt.Sprintf("%s/v2/%s/manifests/%s", registryURL, repository, digest)
	req, err = http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add auth header
	if config.Type == DockerHub {
		req.Header.Add("Authorization", "Bearer "+token)
	} else {
		req.Header.Add("Authorization", "Basic "+token)
	}
	
	resp, err = c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete tag: %s", body)
	}
	
	return nil
}

// GetRegistryType determines the registry type from the URL
func GetRegistryType(url string) RegistryType {
	if strings.Contains(url, "docker.io") || strings.Contains(url, "registry.hub.docker.com") {
		return DockerHub
	} else if strings.Contains(url, "harbor") {
		return Harbor
	} else {
		return Generic
	}
}

// ValidateRegistryConfig validates the registry configuration
func ValidateRegistryConfig(config RegistryConfig) error {
	if config.URL == "" {
		return fmt.Errorf("registry URL is required")
	}
	
	if config.Username == "" {
		return fmt.Errorf("registry username is required")
	}
	
	if config.Password == "" {
		return fmt.Errorf("registry password is required")
	}
	
	return nil
}

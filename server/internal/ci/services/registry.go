package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/vanhcao3/pipeslicerCI/internal/ci/models"
	"github.com/vanhcao3/pipeslicerCI/internal/ci/repository"
)

// RegistryService handles business logic for registry operations
type RegistryService struct {
	repo *repository.RegistryRepository
}

// NewRegistryService creates a new instance of RegistryService
func NewRegistryService(repo *repository.RegistryRepository) *RegistryService {
	return &RegistryService{
		repo: repo,
	}
}

// CreateRegistry creates a new registry
func (s *RegistryService) CreateRegistry(ctx context.Context, registry *models.Registry) error {
	// Check if registry with same name already exists
	existing, err := s.repo.GetByName(ctx, registry.Name)
	if err != nil && !errors.Is(err, repository.ErrRegistryNotFound) {
		return err
	}
	if existing != nil {
		return errors.New("registry with this name already exists")
	}

	return s.repo.Create(ctx, registry)
}

// GetRegistry retrieves a registry by ID
func (s *RegistryService) GetRegistry(ctx context.Context, id uint) (*models.Registry, error) {
	return s.repo.GetByID(ctx, id)
}

// GetRegistryByName retrieves a registry by name
func (s *RegistryService) GetRegistryByName(ctx context.Context, name string) (*models.Registry, error) {
	return s.repo.GetByName(ctx, name)
}

// ListRegistries retrieves all registries
func (s *RegistryService) ListRegistries(ctx context.Context) ([]models.Registry, error) {
	return s.repo.List(ctx)
}

// UpdateRegistry updates an existing registry
func (s *RegistryService) UpdateRegistry(ctx context.Context, registry *models.Registry) error {
	// Check if registry exists
	existing, err := s.repo.GetByID(ctx, registry.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return repository.ErrRegistryNotFound
	}

	// Check if new name conflicts with another registry
	if registry.Name != existing.Name {
		nameExists, err := s.repo.GetByName(ctx, registry.Name)
		if err != nil && !errors.Is(err, repository.ErrRegistryNotFound) {
			return err
		}
		if nameExists != nil {
			return errors.New("registry with this name already exists")
		}
	}

	return s.repo.Update(ctx, registry)
}

// DeleteRegistry deletes a registry by ID
func (s *RegistryService) DeleteRegistry(ctx context.Context, id uint) error {
	// Check if registry exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return repository.ErrRegistryNotFound
	}

	return s.repo.Delete(ctx, id)
}

// TestConnectionResponse represents the response from testing a registry connection
type TestConnectionResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// TestConnection tests the connection to a registry
func (s *RegistryService) TestConnection(ctx context.Context, id uint) (*TestConnectionResponse, error) {
	// Get the registry by ID
	registry, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) {
			return nil, repository.ErrRegistryNotFound
		}
		return nil, err
	}

	// Create a Docker client
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	// Try to authenticate with the registry
	authConfig := types.AuthConfig{
		Username:      registry.Username,
		Password:      registry.Password,
		ServerAddress: registry.URL,
	}

	_, err = client.RegistryLogin(ctx, authConfig)
	if err != nil {
		return &TestConnectionResponse{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to authenticate with registry: %v", err),
		}, nil
	}

	return &TestConnectionResponse{
		Status:  "success",
		Message: "Successfully connected to registry",
	}, nil
}

// DockerImage represents a Docker image in a registry
type DockerImage struct {
	Name        string   `json:"name"`
	Tags        []string `json:"tags"`
	Size        int64    `json:"size"`
	CreatedAt   string   `json:"created_at"`
	LastUpdated string   `json:"last_updated"`
}

// ListImages retrieves all Docker images from a registry
func (s *RegistryService) ListImages(ctx context.Context, registryID uint) ([]DockerImage, error) {
	// Get the registry by ID
	registry, err := s.repo.GetByID(ctx, registryID)
	if err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) {
			return nil, repository.ErrRegistryNotFound
		}
		return nil, err
	}

	// Create a Docker client
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	// Authenticate with the registry
	authConfig := types.AuthConfig{
		Username:      registry.Username,
		Password:      registry.Password,
		ServerAddress: registry.URL,
	}

	// Get authentication token
	authResponse, err := client.RegistryLogin(ctx, authConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with registry: %w", err)
	}

	// Create a registry client
	registryClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Try both HTTP and HTTPS
	protocols := []string{"https", "http"}
	var lastError error

	for _, protocol := range protocols {
		// Construct the registry API URL
		registryURL := fmt.Sprintf("%s://%s/v2", protocol, registry.URL)

		// Get catalog of repositories
		catalogURL := fmt.Sprintf("%s/_catalog", registryURL)
		req, err := http.NewRequestWithContext(ctx, "GET", catalogURL, nil)
		if err != nil {
			lastError = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		// Add authentication header
		req.Header.Set("Authorization", "Bearer "+authResponse.IdentityToken)

		// Send request
		resp, err := registryClient.Do(req)
		if err != nil {
			lastError = fmt.Errorf("failed to get catalog using %s: %w", protocol, err)
			continue
		}
		defer resp.Body.Close()

		// Parse response
		var catalog struct {
			Repositories []string `json:"repositories"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
			lastError = fmt.Errorf("failed to decode catalog response: %w", err)
			continue
		}

		// Get tags for each repository
		var images []DockerImage
		for _, repo := range catalog.Repositories {
			// Get tags for this repository
			tagsURL := fmt.Sprintf("%s/%s/tags/list", registryURL, repo)
			req, err := http.NewRequestWithContext(ctx, "GET", tagsURL, nil)
			if err != nil {
				continue // Skip this repository if we can't create a request
			}

			// Add authentication header
			req.Header.Set("Authorization", "Bearer "+authResponse.IdentityToken)

			// Send request
			resp, err := registryClient.Do(req)
			if err != nil {
				continue // Skip this repository if the request fails
			}

			// Parse response
			var tagsResponse struct {
				Name string   `json:"name"`
				Tags []string `json:"tags"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&tagsResponse); err != nil {
				resp.Body.Close()
				continue // Skip this repository if we can't decode the response
			}
			resp.Body.Close()

			// Only add images that have tags
			if len(tagsResponse.Tags) > 0 {
				images = append(images, DockerImage{
					Name: repo,
					Tags: tagsResponse.Tags,
				})
			}
		}

		// If we got here, we successfully retrieved the images
		return images, nil
	}

	// If we get here, both protocols failed
	return nil, lastError
}

// DockerImageDetail represents detailed information about a Docker image
type DockerImageDetail struct {
	Name        string            `json:"name"`
	Tags        []string          `json:"tags"`
	Size        int64             `json:"size"`
	CreatedAt   string            `json:"created_at"`
	LastUpdated string            `json:"last_updated"`
	Layers      []ImageLayer      `json:"layers"`
	History     []ImageHistory    `json:"history"`
	Config      ImageConfig       `json:"config"`
	Labels      map[string]string `json:"labels"`
}

// ImageLayer represents a layer in a Docker image
type ImageLayer struct {
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
}

// ImageHistory represents a history entry in a Docker image
type ImageHistory struct {
	Created    string `json:"created"`
	CreatedBy  string `json:"created_by"`
	Comment    string `json:"comment"`
	EmptyLayer bool   `json:"empty_layer"`
}

// ImageConfig represents the configuration of a Docker image
type ImageConfig struct {
	Architecture string            `json:"architecture"`
	OS           string            `json:"os"`
	Env          []string          `json:"env"`
	Labels       map[string]string `json:"labels"`
}

// RetagImageRequest represents the request to retag a Docker image
type RetagImageRequest struct {
	SourceImage      string `json:"source_image"`
	SourceTag        string `json:"source_tag"`
	DestinationImage string `json:"destination_image"`
	DestinationTag   string `json:"destination_tag"`
}

// CopyImageRequest represents the request to copy a Docker image between registries
type CopyImageRequest struct {
	SourceRegistryID      uint   `json:"source_registry_id"`
	SourceImage           string `json:"source_image"`
	SourceTag             string `json:"source_tag"`
	DestinationRegistryID uint   `json:"destination_registry_id"`
	DestinationImage      string `json:"destination_image"`
	DestinationTag        string `json:"destination_tag"`
}

// GetImageDetail retrieves detailed information about a specific Docker image
func (s *RegistryService) GetImageDetail(ctx context.Context, registryID uint, imageName, tag string) (*DockerImageDetail, error) {
	// Get the registry by ID
	registry, err := s.repo.GetByID(ctx, registryID)
	if err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) {
			return nil, repository.ErrRegistryNotFound
		}
		return nil, err
	}

	// Create a Docker client
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	// Authenticate with the registry
	authConfig := types.AuthConfig{
		Username:      registry.Username,
		Password:      registry.Password,
		ServerAddress: registry.URL,
	}

	// Get authentication token
	authResponse, err := client.RegistryLogin(ctx, authConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with registry: %w", err)
	}

	// Create a registry client
	registryClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Try both HTTP and HTTPS
	protocols := []string{"https", "http"}
	var lastError error

	for _, protocol := range protocols {
		// Construct the registry API URL
		registryURL := fmt.Sprintf("%s://%s/v2", protocol, registry.URL)

		// Get manifest for the image
		manifestURL := fmt.Sprintf("%s/%s/manifests/%s", registryURL, imageName, tag)
		req, err := http.NewRequestWithContext(ctx, "GET", manifestURL, nil)
		if err != nil {
			lastError = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		// Add authentication header
		req.Header.Set("Authorization", "Bearer "+authResponse.IdentityToken)
		req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

		// Send request
		resp, err := registryClient.Do(req)
		if err != nil {
			lastError = fmt.Errorf("failed to get manifest: %w", err)
			continue
		}
		defer resp.Body.Close()

		// Parse manifest response
		var manifest struct {
			SchemaVersion int    `json:"schemaVersion"`
			MediaType     string `json:"mediaType"`
			Config        struct {
				MediaType string `json:"mediaType"`
				Size      int    `json:"size"`
				Digest    string `json:"digest"`
			} `json:"config"`
			Layers []struct {
				MediaType string `json:"mediaType"`
				Size      int    `json:"size"`
				Digest    string `json:"digest"`
			} `json:"layers"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
			lastError = fmt.Errorf("failed to decode manifest: %w", err)
			continue
		}

		// Get config
		configURL := fmt.Sprintf("%s/%s/blobs/%s", registryURL, imageName, manifest.Config.Digest)
		req, err = http.NewRequestWithContext(ctx, "GET", configURL, nil)
		if err != nil {
			lastError = fmt.Errorf("failed to create config request: %w", err)
			continue
		}

		req.Header.Set("Authorization", "Bearer "+authResponse.IdentityToken)

		resp, err = registryClient.Do(req)
		if err != nil {
			lastError = fmt.Errorf("failed to get config: %w", err)
			continue
		}
		defer resp.Body.Close()

		var config struct {
			Architecture string `json:"architecture"`
			OS           string `json:"os"`
			Config       struct {
				Env    []string          `json:"Env"`
				Labels map[string]string `json:"Labels"`
			} `json:"config"`
			History []struct {
				Created    string `json:"created"`
				CreatedBy  string `json:"created_by"`
				Comment    string `json:"comment"`
				EmptyLayer bool   `json:"empty_layer"`
			} `json:"history"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
			lastError = fmt.Errorf("failed to decode config: %w", err)
			continue
		}

		// Construct response
		detail := &DockerImageDetail{
			Name:    imageName,
			Tags:    []string{tag},
			Size:    int64(manifest.Config.Size),
			Layers:  make([]ImageLayer, len(manifest.Layers)),
			History: make([]ImageHistory, len(config.History)),
			Config: ImageConfig{
				Architecture: config.Architecture,
				OS:           config.OS,
				Env:          config.Config.Env,
				Labels:       config.Config.Labels,
			},
		}

		// Add layers
		for i, layer := range manifest.Layers {
			detail.Layers[i] = ImageLayer{
				Digest: layer.Digest,
				Size:   int64(layer.Size),
			}
		}

		// Add history
		for i, hist := range config.History {
			detail.History[i] = ImageHistory{
				Created:    hist.Created,
				CreatedBy:  hist.CreatedBy,
				Comment:    hist.Comment,
				EmptyLayer: hist.EmptyLayer,
			}
		}

		return detail, nil
	}

	return nil, lastError
}

// RetagImage retags a Docker image within the same registry
func (s *RegistryService) RetagImage(ctx context.Context, registryID uint, req RetagImageRequest) error {
	// Validate request parameters
	if req.SourceImage == "" || req.SourceTag == "" || req.DestinationImage == "" || req.DestinationTag == "" {
		return fmt.Errorf("all fields (source_image, source_tag, destination_image, destination_tag) are required")
	}

	// Get the registry by ID
	registry, err := s.repo.GetByID(ctx, registryID)
	if err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) {
			return repository.ErrRegistryNotFound
		}
		return err
	}

	fmt.Printf("Retagging image in registry: %s (ID: %d)\n", registry.URL, registryID)

	// Create a Docker client
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	// Authenticate with the registry
	authConfig := types.AuthConfig{
		Username:      registry.Username,
		Password:      registry.Password,
		ServerAddress: registry.URL,
	}

	// Get authentication token
	authResponse, err := client.RegistryLogin(ctx, authConfig)
	if err != nil {
		return fmt.Errorf("failed to authenticate with registry: %w", err)
	}

	fmt.Printf("Successfully authenticated with registry: %s\n", registry.URL)

	// Create a registry client
	registryClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Try both HTTP and HTTPS
	protocols := []string{"https", "http"}
	var lastError error

	for _, protocol := range protocols {
		// Construct the registry API URL
		registryURL := fmt.Sprintf("%s://%s/v2", protocol, registry.URL)
		fmt.Printf("Trying protocol: %s, URL: %s\n", protocol, registryURL)

		// Get manifest for the source image
		// Ensure no double slashes by using path.Join or proper string formatting
		sourceManifestURL := fmt.Sprintf("%s/%s/manifests/%s", registryURL, strings.TrimPrefix(req.SourceImage, "/"), req.SourceTag)
		fmt.Printf("Fetching source manifest from: %s\n", sourceManifestURL)

		httpReq, err := http.NewRequestWithContext(ctx, "GET", sourceManifestURL, nil)
		if err != nil {
			lastError = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		// Add authentication header
		httpReq.Header.Set("Authorization", "Bearer "+authResponse.IdentityToken)
		httpReq.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

		// Send request
		resp, err := registryClient.Do(httpReq)
		if err != nil {
			lastError = fmt.Errorf("failed to get source manifest: %w", err)
			continue
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			lastError = fmt.Errorf("failed to get source manifest: status %d, body: %s", resp.StatusCode, string(body))
			fmt.Printf("Error response from registry: %s\n", lastError)
			continue
		}

		// Read manifest content
		manifestContent, err := io.ReadAll(resp.Body)
		if err != nil {
			lastError = fmt.Errorf("failed to read manifest: %w", err)
			continue
		}

		fmt.Printf("Successfully retrieved manifest for %s:%s\n", req.SourceImage, req.SourceTag)

		// Put manifest for the destination image
		// Ensure no double slashes by using path.Join or proper string formatting
		destManifestURL := fmt.Sprintf("%s/%s/manifests/%s", registryURL, strings.TrimPrefix(req.DestinationImage, "/"), req.DestinationTag)
		fmt.Printf("Putting manifest to: %s\n", destManifestURL)

		httpReq, err = http.NewRequestWithContext(ctx, "PUT", destManifestURL, bytes.NewReader(manifestContent))
		if err != nil {
			lastError = fmt.Errorf("failed to create put request: %w", err)
			continue
		}

		// Add authentication header and content type
		httpReq.Header.Set("Authorization", "Bearer "+authResponse.IdentityToken)
		httpReq.Header.Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")

		// Send request
		resp, err = registryClient.Do(httpReq)
		if err != nil {
			lastError = fmt.Errorf("failed to put destination manifest: %w", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			lastError = fmt.Errorf("failed to put destination manifest: status %d, body: %s", resp.StatusCode, string(body))
			fmt.Printf("Error response from registry: %s\n", lastError)
			continue
		}

		fmt.Printf("Successfully retagged image from %s:%s to %s:%s\n",
			req.SourceImage, req.SourceTag, req.DestinationImage, req.DestinationTag)
		return nil
	}

	return fmt.Errorf("failed to retag image: %w", lastError)
}

// deleteBlob deletes a blob (layer or config) from the registry
func deleteBlob(client *http.Client, registryURL, imageName, digest, authToken string) error {
	deleteURL := fmt.Sprintf("%s/%s/blobs/%s", registryURL, imageName, digest)
	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete blob: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete blob: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteImage deletes a Docker image from the registry
func (s *RegistryService) DeleteImage(ctx context.Context, registryID uint, imageName, tag string) error {
	log.Printf("Starting DeleteImage operation for registry ID: %d, image: %s, tag: %s", registryID, imageName, tag)

	// Get the registry by ID
	registry, err := s.repo.GetByID(ctx, registryID)
	if err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) {
			log.Printf("Registry not found with ID: %d", registryID)
			return repository.ErrRegistryNotFound
		}
		log.Printf("Error getting registry by ID %d: %v", registryID, err)
		return err
	}
	log.Printf("Found registry: %s (URL: %s)", registry.Name, registry.URL)

	// Create a Docker client
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Printf("Failed to create Docker client: %v", err)
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	// Authenticate with the registry
	authConfig := types.AuthConfig{
		Username:      registry.Username,
		Password:      registry.Password,
		ServerAddress: registry.URL,
	}

	log.Printf("Attempting to authenticate with registry: %s", registry.URL)
	authResponse, err := client.RegistryLogin(ctx, authConfig)
	if err != nil {
		log.Printf("Failed to authenticate with registry %s: %v", registry.URL, err)
		return fmt.Errorf("failed to authenticate with registry: %w", err)
	}
	log.Printf("Successfully authenticated with registry: %s", registry.URL)

	// Create a registry client
	registryClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Try both HTTP and HTTPS
	protocols := []string{"https", "http"}
	var lastError error

	for _, protocol := range protocols {
		// Construct the registry API URL
		registryURL := fmt.Sprintf("%s://%s/v2", protocol, registry.URL)
		log.Printf("Trying protocol: %s, URL: %s", protocol, registryURL)

		// Get manifest for the image
		manifestURL := fmt.Sprintf("%s/%s/manifests/%s", registryURL, imageName, tag)
		log.Printf("Fetching manifest from: %s", manifestURL)

		req, err := http.NewRequestWithContext(ctx, "GET", manifestURL, nil)
		if err != nil {
			lastError = fmt.Errorf("failed to create request: %w", err)
			log.Printf("Error creating manifest request: %v", err)
			continue
		}

		// Add authentication header
		req.Header.Set("Authorization", "Bearer "+authResponse.IdentityToken)
		req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

		// Send request
		resp, err := registryClient.Do(req)
		if err != nil {
			lastError = fmt.Errorf("failed to get manifest: %w", err)
			log.Printf("Error getting manifest: %v", err)
			continue
		}
		defer resp.Body.Close()

		// Get digest from response header
		digest := resp.Header.Get("Docker-Content-Digest")
		if digest == "" {
			lastError = fmt.Errorf("failed to get digest from response")
			log.Printf("No digest found in response headers. Status code: %d", resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)
			log.Printf("Response body: %s", string(body))
			continue
		}
		log.Printf("Successfully retrieved manifest digest: %s", digest)

		// Parse manifest to get layer digests
		var manifest struct {
			SchemaVersion int    `json:"schemaVersion"`
			MediaType     string `json:"mediaType"`
			Config        struct {
				Digest string `json:"digest"`
			} `json:"config"`
			Layers []struct {
				Digest string `json:"digest"`
			} `json:"layers"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
			lastError = fmt.Errorf("failed to parse manifest: %w", err)
			log.Printf("Error parsing manifest: %v", err)
			continue
		}

		// Get all tags for this image
		tagsURL := fmt.Sprintf("%s/%s/tags/list", registryURL, imageName)
		req, err = http.NewRequestWithContext(ctx, "GET", tagsURL, nil)
		if err != nil {
			lastError = fmt.Errorf("failed to create tags request: %w", err)
			log.Printf("Error creating tags request: %v", err)
			continue
		}

		req.Header.Set("Authorization", "Bearer "+authResponse.IdentityToken)

		resp, err = registryClient.Do(req)
		if err != nil {
			lastError = fmt.Errorf("failed to get tags: %w", err)
			log.Printf("Error getting tags: %v", err)
			continue
		}
		defer resp.Body.Close()

		var tagsResponse struct {
			Name string   `json:"name"`
			Tags []string `json:"tags"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&tagsResponse); err != nil {
			lastError = fmt.Errorf("failed to parse tags response: %w", err)
			log.Printf("Error parsing tags response: %v", err)
			continue
		}

		// Delete all tags pointing to this manifest
		for _, t := range tagsResponse.Tags {
			tagManifestURL := fmt.Sprintf("%s/%s/manifests/%s", registryURL, imageName, t)
			req, err = http.NewRequestWithContext(ctx, "GET", tagManifestURL, nil)
			if err != nil {
				log.Printf("Warning: failed to create request for tag %s: %v", t, err)
				continue
			}

			req.Header.Set("Authorization", "Bearer "+authResponse.IdentityToken)
			req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

			resp, err = registryClient.Do(req)
			if err != nil {
				log.Printf("Warning: failed to get manifest for tag %s: %v", t, err)
				continue
			}

			tagDigest := resp.Header.Get("Docker-Content-Digest")
			resp.Body.Close()

			if tagDigest == digest {
				log.Printf("Deleting tag: %s", t)
				deleteURL := fmt.Sprintf("%s/%s/manifests/%s", registryURL, imageName, tagDigest)
				req, err = http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
				if err != nil {
					log.Printf("Warning: failed to create delete request for tag %s: %v", t, err)
					continue
				}

				req.Header.Set("Authorization", "Bearer "+authResponse.IdentityToken)

				resp, err = registryClient.Do(req)
				if err != nil {
					log.Printf("Warning: failed to delete tag %s: %v", t, err)
					continue
				}
				resp.Body.Close()
			}
		}

		// Delete config layer
		configDigest := manifest.Config.Digest
		log.Printf("Deleting config layer: %s", configDigest)
		if err := deleteBlob(registryClient, registryURL, imageName, configDigest, authResponse.IdentityToken); err != nil {
			log.Printf("Warning: failed to delete config layer: %v", err)
		}

		// Delete all layers
		for _, layer := range manifest.Layers {
			log.Printf("Deleting layer: %s", layer.Digest)
			if err := deleteBlob(registryClient, registryURL, imageName, layer.Digest, authResponse.IdentityToken); err != nil {
				log.Printf("Warning: failed to delete layer %s: %v", layer.Digest, err)
			}
		}

		// Run garbage collection
		gcURL := fmt.Sprintf("%s/_catalog", registryURL)
		req, err = http.NewRequestWithContext(ctx, "GET", gcURL, nil)
		if err != nil {
			log.Printf("Warning: failed to create GC request: %v", err)
		} else {
			req.Header.Set("Authorization", "Bearer "+authResponse.IdentityToken)
			resp, err = registryClient.Do(req)
			if err != nil {
				log.Printf("Warning: failed to trigger GC: %v", err)
			} else {
				resp.Body.Close()
			}
		}

		log.Printf("Successfully deleted image %s:%s from registry %s", imageName, tag, registry.URL)
		return nil
	}

	log.Printf("Failed to delete image after trying all protocols. Last error: %v", lastError)
	return lastError
}

// CopyImage copies a Docker image from one registry to another
func (s *RegistryService) CopyImage(ctx context.Context, req CopyImageRequest) error {
	log.Printf("Starting image copy operation: %+v", req)

	// Get source registry
	sourceRegistry, err := s.repo.GetByID(ctx, req.SourceRegistryID)
	if err != nil {
		log.Printf("Error getting source registry (ID: %d): %v", req.SourceRegistryID, err)
		if errors.Is(err, repository.ErrRegistryNotFound) {
			return fmt.Errorf("source registry not found: %w", err)
		}
		return err
	}
	log.Printf("Found source registry: %s (URL: %s)", sourceRegistry.Name, sourceRegistry.URL)

	// Get destination registry
	destRegistry, err := s.repo.GetByID(ctx, req.DestinationRegistryID)
	if err != nil {
		log.Printf("Error getting destination registry (ID: %d): %v", req.DestinationRegistryID, err)
		if errors.Is(err, repository.ErrRegistryNotFound) {
			return fmt.Errorf("destination registry not found: %w", err)
		}
		return err
	}
	log.Printf("Found destination registry: %s (URL: %s)", destRegistry.Name, destRegistry.URL)

	// Create a Docker client
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Printf("Error creating Docker client: %v", err)
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	// Authenticate with source registry
	sourceAuthConfig := types.AuthConfig{
		Username:      sourceRegistry.Username,
		Password:      sourceRegistry.Password,
		ServerAddress: sourceRegistry.URL,
	}

	log.Printf("Attempting to authenticate with source registry: %s", sourceRegistry.URL)
	sourceAuthResponse, err := client.RegistryLogin(ctx, sourceAuthConfig)
	if err != nil {
		log.Printf("Error authenticating with source registry: %v", err)
		return fmt.Errorf("failed to authenticate with source registry: %w", err)
	}
	log.Printf("Successfully authenticated with source registry")

	// Authenticate with destination registry
	destAuthConfig := types.AuthConfig{
		Username:      destRegistry.Username,
		Password:      destRegistry.Password,
		ServerAddress: destRegistry.URL,
	}

	log.Printf("Attempting to authenticate with destination registry: %s", destRegistry.URL)
	destAuthResponse, err := client.RegistryLogin(ctx, destAuthConfig)
	if err != nil {
		log.Printf("Error authenticating with destination registry: %v", err)
		return fmt.Errorf("failed to authenticate with destination registry: %w", err)
	}
	log.Printf("Successfully authenticated with destination registry")

	// Create registry clients
	registryClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Try both HTTP and HTTPS for source registry
	sourceProtocols := []string{"https", "http"}
	var lastError error

	for _, sourceProtocol := range sourceProtocols {
		// Construct the source registry API URL
		sourceRegistryURL := fmt.Sprintf("%s://%s/v2", sourceProtocol, sourceRegistry.URL)
		log.Printf("Trying source protocol: %s, URL: %s", sourceProtocol, sourceRegistryURL)

		// Get manifest from source registry
		sourceManifestURL := fmt.Sprintf("%s/%s/manifests/%s", sourceRegistryURL, req.SourceImage, req.SourceTag)
		log.Printf("Fetching manifest from: %s", sourceManifestURL)

		httpReq, err := http.NewRequestWithContext(ctx, "GET", sourceManifestURL, nil)
		if err != nil {
			lastError = fmt.Errorf("failed to create source request: %w", err)
			log.Printf("Error creating source request: %v", err)
			continue
		}

		// Add source authentication header
		httpReq.Header.Set("Authorization", "Bearer "+sourceAuthResponse.IdentityToken)
		httpReq.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

		// Send request
		resp, err := registryClient.Do(httpReq)
		if err != nil {
			lastError = fmt.Errorf("failed to get source manifest: %w", err)
			log.Printf("Error getting source manifest: %v", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("Unexpected status code from source registry: %d, body: %s", resp.StatusCode, string(body))
			resp.Body.Close()
			lastError = fmt.Errorf("failed to get source manifest: status %d", resp.StatusCode)
			continue
		}

		// Read manifest content
		manifestContent, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastError = fmt.Errorf("failed to read manifest: %w", err)
			log.Printf("Error reading manifest: %v", err)
			continue
		}
		log.Printf("Successfully retrieved manifest from source registry")

		// Parse manifest to get layer digests
		var manifest struct {
			SchemaVersion int    `json:"schemaVersion"`
			MediaType     string `json:"mediaType"`
			Config        struct {
				MediaType string `json:"mediaType"`
				Size      int    `json:"size"`
				Digest    string `json:"digest"`
			} `json:"config"`
			Layers []struct {
				MediaType string `json:"mediaType"`
				Size      int    `json:"size"`
				Digest    string `json:"digest"`
			} `json:"layers"`
		}

		if err := json.Unmarshal(manifestContent, &manifest); err != nil {
			lastError = fmt.Errorf("failed to parse manifest: %w", err)
			log.Printf("Error parsing manifest: %v", err)
			continue
		}

		// Try both HTTP and HTTPS for destination registry
		destProtocols := []string{"https", "http"}
		for _, destProtocol := range destProtocols {
			// Construct the destination registry API URL
			destRegistryURL := fmt.Sprintf("%s://%s/v2", destProtocol, destRegistry.URL)
			log.Printf("Trying destination protocol: %s, URL: %s", destProtocol, destRegistryURL)

			// First, copy all blobs (layers and config)
			blobsToCopy := append([]struct {
				MediaType string `json:"mediaType"`
				Size      int    `json:"size"`
				Digest    string `json:"digest"`
			}{manifest.Config}, manifest.Layers...)
			for _, blob := range blobsToCopy {
				// Check if blob exists in destination
				checkURL := fmt.Sprintf("%s/%s/blobs/%s", destRegistryURL, req.DestinationImage, blob.Digest)
				checkReq, err := http.NewRequestWithContext(ctx, "HEAD", checkURL, nil)
				if err != nil {
					log.Printf("Error creating blob check request: %v", err)
					continue
				}
				checkReq.Header.Set("Authorization", "Bearer "+destAuthResponse.IdentityToken)

				checkResp, err := registryClient.Do(checkReq)
				if err != nil {
					log.Printf("Error checking blob existence: %v", err)
					continue
				}
				checkResp.Body.Close()

				// If blob exists (200 OK), skip copying
				if checkResp.StatusCode == http.StatusOK {
					log.Printf("Blob %s already exists in destination registry", blob.Digest)
					continue
				}

				// Get blob from source
				sourceBlobURL := fmt.Sprintf("%s/%s/blobs/%s", sourceRegistryURL, req.SourceImage, blob.Digest)
				sourceReq, err := http.NewRequestWithContext(ctx, "GET", sourceBlobURL, nil)
				if err != nil {
					log.Printf("Error creating source blob request: %v", err)
					continue
				}
				sourceReq.Header.Set("Authorization", "Bearer "+sourceAuthResponse.IdentityToken)

				sourceResp, err := registryClient.Do(sourceReq)
				if err != nil {
					log.Printf("Error getting source blob: %v", err)
					continue
				}

				if sourceResp.StatusCode != http.StatusOK {
					body, _ := io.ReadAll(sourceResp.Body)
					log.Printf("Error getting source blob: status %d, body: %s", sourceResp.StatusCode, string(body))
					sourceResp.Body.Close()
					continue
				}

				// Initiate blob upload to destination
				initiateURL := fmt.Sprintf("%s/%s/blobs/uploads/", destRegistryURL, req.DestinationImage)
				initiateReq, err := http.NewRequestWithContext(ctx, "POST", initiateURL, nil)
				if err != nil {
					sourceResp.Body.Close()
					log.Printf("Error creating initiate upload request: %v", err)
					continue
				}
				initiateReq.Header.Set("Authorization", "Bearer "+destAuthResponse.IdentityToken)

				initiateResp, err := registryClient.Do(initiateReq)
				if err != nil {
					sourceResp.Body.Close()
					log.Printf("Error initiating blob upload: %v", err)
					continue
				}

				if initiateResp.StatusCode != http.StatusAccepted {
					body, _ := io.ReadAll(initiateResp.Body)
					log.Printf("Error initiating blob upload: status %d, body: %s", initiateResp.StatusCode, string(body))
					initiateResp.Body.Close()
					sourceResp.Body.Close()
					continue
				}

				// Get upload URL from Location header
				uploadURL := initiateResp.Header.Get("Location")
				initiateResp.Body.Close()

				// Upload blob to destination
				uploadReq, err := http.NewRequestWithContext(ctx, "PUT", uploadURL+"&digest="+blob.Digest, sourceResp.Body)
				if err != nil {
					sourceResp.Body.Close()
					log.Printf("Error creating upload request: %v", err)
					continue
				}
				uploadReq.Header.Set("Authorization", "Bearer "+destAuthResponse.IdentityToken)
				uploadReq.Header.Set("Content-Type", "application/octet-stream")

				uploadResp, err := registryClient.Do(uploadReq)
				sourceResp.Body.Close()
				if err != nil {
					log.Printf("Error uploading blob: %v", err)
					continue
				}
				uploadResp.Body.Close()

				if uploadResp.StatusCode != http.StatusCreated {
					body, _ := io.ReadAll(uploadResp.Body)
					log.Printf("Error uploading blob: status %d, body: %s", uploadResp.StatusCode, string(body))
					continue
				}

				log.Printf("Successfully copied blob %s", blob.Digest)
			}

			// Now copy the manifest
			destManifestURL := fmt.Sprintf("%s/%s/manifests/%s", destRegistryURL, req.DestinationImage, req.DestinationTag)
			log.Printf("Putting manifest to: %s", destManifestURL)

			httpReq, err = http.NewRequestWithContext(ctx, "PUT", destManifestURL, bytes.NewReader(manifestContent))
			if err != nil {
				lastError = fmt.Errorf("failed to create destination request: %w", err)
				log.Printf("Error creating destination request: %v", err)
				continue
			}

			// Add destination authentication header
			httpReq.Header.Set("Authorization", "Bearer "+destAuthResponse.IdentityToken)
			httpReq.Header.Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")

			// Send request
			resp, err = registryClient.Do(httpReq)
			if err != nil {
				lastError = fmt.Errorf("failed to put destination manifest: %w", err)
				log.Printf("Error putting manifest to destination: %v", err)
				continue
			}

			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				log.Printf("Unexpected status code from destination registry: %d, body: %s", resp.StatusCode, string(body))
				lastError = fmt.Errorf("failed to put destination manifest: status %d", resp.StatusCode)
				continue
			}

			log.Printf("Successfully copied image from %s:%s to %s:%s", req.SourceImage, req.SourceTag, req.DestinationImage, req.DestinationTag)
			return nil
		}
	}

	log.Printf("Failed to copy image after trying all protocols. Last error: %v", lastError)
	return lastError
}

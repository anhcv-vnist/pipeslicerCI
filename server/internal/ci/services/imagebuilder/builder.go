package imagebuilder

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/vanhcao3/pipeslicerCI/internal/ci"
)

// ImageBuilder is responsible for building Docker images and pushing them to a registry
type ImageBuilder struct {
	workspace ci.Workspace
	registry  string
	username  string
	password  string
}

// ImageBuildResult contains the result of a Docker image build operation
type ImageBuildResult struct {
	Service   string
	Tag       string
	Commit    string
	Branch    string
	BuildTime time.Time
	Success   bool
	Output    string
	Error     error
}

// NewImageBuilder creates a new ImageBuilder instance
func NewImageBuilder(ws ci.Workspace, registry, username, password string) *ImageBuilder {
	return &ImageBuilder{
		workspace: ws,
		registry:  registry,
		username:  username,
		password:  password,
	}
}

// BuildAndPushImage builds a Docker image for the specified service and pushes it to the registry
func (b *ImageBuilder) BuildAndPushImage(ctx context.Context, servicePath, tag string) (*ImageBuildResult, error) {
	result := &ImageBuildResult{
		Service:   filepath.Base(servicePath),
		Tag:       tag,
		Commit:    b.workspace.Commit(),
		Branch:    b.workspace.Branch(),
		BuildTime: time.Now(),
		Success:   false,
	}

	// Login to registry only if username and password are provided
	var loginOutput []byte
	if b.username != "" && b.password != "" {
		loginCmd := []string{"docker", "login", b.registry, "-u", b.username, "-p", b.password}
		var err error
		loginOutput, err = b.workspace.ExecuteCommand(ctx, loginCmd[0], loginCmd[1:])
		if err != nil {
			result.Output = string(loginOutput)
			result.Error = fmt.Errorf("failed to login to registry: %w", err)
			return result, result.Error
		}
	}

	// Build image
	fullPath := filepath.Join(b.workspace.Dir(), servicePath)
	imageName := fmt.Sprintf("%s/%s:%s", b.registry, result.Service, tag)

	// Check if Dockerfile exists in the service directory
	dockerfilePath := filepath.Join(fullPath, "Dockerfile")
	checkCmd := []string{"test", "-f", dockerfilePath}
	_, checkErr := b.workspace.ExecuteCommand(ctx, checkCmd[0], checkCmd[1:])
	if checkErr != nil {
		result.Output = string(loginOutput) + "\n" + fmt.Sprintf("Dockerfile not found at %s", dockerfilePath)
		result.Error = fmt.Errorf("failed to build service %s: Dockerfile not found", result.Service)
		return result, result.Error
	}

	buildCmd := []string{"docker", "build", "-t", imageName, fullPath}
	buildOutput, err := b.workspace.ExecuteCommand(ctx, buildCmd[0], buildCmd[1:])
	if err != nil {
		result.Output = string(loginOutput) + "\n" + string(buildOutput)
		result.Error = fmt.Errorf("failed to build image: %w", err)
		return result, result.Error
	}

	// Push image
	pushCmd := []string{"docker", "push", imageName}
	pushOutput, err := b.workspace.ExecuteCommand(ctx, pushCmd[0], pushCmd[1:])
	if err != nil {
		result.Output = string(loginOutput) + "\n" + string(buildOutput) + "\n" + string(pushOutput)
		result.Error = fmt.Errorf("failed to push image: %w", err)
		return result, result.Error
	}

	// Tag as latest if not already
	if tag != "latest" {
		latestImageName := fmt.Sprintf("%s/%s:latest", b.registry, result.Service)
		tagCmd := []string{"docker", "tag", imageName, latestImageName}
		tagOutput, err := b.workspace.ExecuteCommand(ctx, tagCmd[0], tagCmd[1:])
		if err != nil {
			// Non-critical error, just log it
			result.Output = string(loginOutput) + "\n" + string(buildOutput) + "\n" + string(pushOutput) + "\n" + string(tagOutput)
			result.Output += fmt.Sprintf("\nWarning: Failed to tag image as latest: %v", err)
		} else {
			// Push latest tag
			pushLatestCmd := []string{"docker", "push", latestImageName}
			pushLatestOutput, err := b.workspace.ExecuteCommand(ctx, pushLatestCmd[0], pushLatestCmd[1:])
			if err != nil {
				// Non-critical error, just log it
				result.Output = string(loginOutput) + "\n" + string(buildOutput) + "\n" + string(pushOutput) + "\n" + string(tagOutput) + "\n" + string(pushLatestOutput)
				result.Output += fmt.Sprintf("\nWarning: Failed to push latest tag: %v", err)
			} else {
				result.Output = string(loginOutput) + "\n" + string(buildOutput) + "\n" + string(pushOutput) + "\n" + string(tagOutput) + "\n" + string(pushLatestOutput)
			}
		}
	} else {
		result.Output = string(loginOutput) + "\n" + string(buildOutput) + "\n" + string(pushOutput)
	}

	result.Success = true
	return result, nil
}

// BuildMultipleServices builds and pushes Docker images for multiple services
func (b *ImageBuilder) BuildMultipleServices(ctx context.Context, servicePaths []string, tag string) ([]*ImageBuildResult, error) {
	var results []*ImageBuildResult

	for _, servicePath := range servicePaths {
		result, err := b.BuildAndPushImage(ctx, servicePath, tag)
		results = append(results, result)
		if err != nil {
			return results, fmt.Errorf("failed to build service %s: %w", servicePath, err)
		}
	}

	return results, nil
}

// DetectChangedServices analyzes git changes to determine which services need to be rebuilt
func (b *ImageBuilder) DetectChangedServices(ctx context.Context, baseBranch, currentBranch string) ([]string, error) {
	// First, make sure we have both branches available
	// Fetch the base branch if it's not the current one
	if baseBranch != b.workspace.Branch() {
		fetchCmd := []string{"git", "fetch", "origin", baseBranch + ":" + baseBranch}
		fetchOutput, err := b.workspace.ExecuteCommand(ctx, fetchCmd[0], fetchCmd[1:])
		if err != nil {
			// If fetch fails, try a different approach - fetch all branches
			fetchAllCmd := []string{"git", "fetch", "--all"}
			fetchAllOutput, err := b.workspace.ExecuteCommand(ctx, fetchAllCmd[0], fetchAllCmd[1:])
			if err != nil {
				return nil, fmt.Errorf("failed to fetch branches: %s\n%w", string(fetchAllOutput), err)
			}
		} else {
			// Log the fetch output for debugging
			fmt.Printf("Fetch output: %s\n", string(fetchOutput))
		}
	}

	// Now try to get the diff between branches
	// Use git log to find changed files instead of git diff, which is more reliable in this context
	logCmd := []string{"git", "log", "--name-only", "--pretty=format:", baseBranch + ".." + currentBranch}
	logOutput, err := b.workspace.ExecuteCommand(ctx, logCmd[0], logCmd[1:])
	if err != nil {
		return nil, fmt.Errorf("failed to get git log: %w", err)
	}

	// Parse the output to get changed files
	changedFiles := strings.Split(string(logOutput), "\n")

	// If no files were found, try an alternative approach
	if len(changedFiles) == 0 || (len(changedFiles) == 1 && changedFiles[0] == "") {
		// Try using git diff with origin/ prefix
		diffCmd := []string{"git", "diff", "--name-only", fmt.Sprintf("origin/%s...origin/%s", baseBranch, currentBranch)}
		diffOutput, err := b.workspace.ExecuteCommand(ctx, diffCmd[0], diffCmd[1:])
		if err != nil {
			// If that fails too, just return all services as changed
			// This is a fallback to ensure the build doesn't fail completely
			lsCmd := []string{"find", "micro-services", "-maxdepth", "1", "-type", "d", "-not", "-path", "micro-services"}
			lsOutput, err := b.workspace.ExecuteCommand(ctx, lsCmd[0], lsCmd[1:])
			if err != nil {
				return nil, fmt.Errorf("failed to list services: %w", err)
			}

			services := strings.Split(string(lsOutput), "\n")
			var allServices []string
			for _, service := range services {
				if service != "" {
					allServices = append(allServices, service)
				}
			}

			// Log that we're returning all services due to inability to determine changes
			fmt.Printf("Warning: Could not determine changed files. Returning all services as changed.\n")
			return allServices, nil
		}

		changedFiles = strings.Split(string(diffOutput), "\n")
	}

	// Map to store unique services that need rebuilding
	serviceMap := make(map[string]bool)

	// Analyze changed files to determine affected services
	for _, file := range changedFiles {
		if file == "" {
			continue
		}

		// Check if the file is in a service directory
		// This assumes a structure like micro-services/service-name/...
		parts := strings.Split(file, "/")
		if len(parts) >= 2 && parts[0] == "micro-services" {
			serviceMap[filepath.Join(parts[0], parts[1])] = true
		}

		// Check for changes in shared libraries or configuration
		if strings.HasPrefix(file, "shared/") || file == "docker-compose.yml" {
			// If shared code changes, we might need to rebuild all services
			// This is a simplified approach - in a real system, you'd have a more sophisticated dependency analysis
			lsCmd := []string{"find", "micro-services", "-maxdepth", "1", "-type", "d", "-not", "-path", "micro-services"}
			lsOutput, err := b.workspace.ExecuteCommand(ctx, lsCmd[0], lsCmd[1:])
			if err != nil {
				return nil, fmt.Errorf("failed to list services: %w", err)
			}

			services := strings.Split(string(lsOutput), "\n")
			for _, service := range services {
				if service != "" {
					serviceMap[service] = true
				}
			}
			break
		}
	}

	// Convert map to slice
	var changedServices []string
	for service := range serviceMap {
		changedServices = append(changedServices, service)
	}

	return changedServices, nil
}

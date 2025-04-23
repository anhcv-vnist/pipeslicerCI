package imagebuilder

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
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

// ChangedServiceInfo contains information about a changed service
type ChangedServiceInfo struct {
	Path          string `json:"path"`
	HasDockerfile bool   `json:"hasDockerfile"`
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
func (b *ImageBuilder) DetectChangedServices(ctx context.Context, baseBranch, currentBranch string) ([]ChangedServiceInfo, error) {
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
			var allServices []ChangedServiceInfo
			for _, service := range services {
				if service != "" {
					// Check if the service has a Dockerfile
					dockerfilePath := filepath.Join(b.workspace.Dir(), service, "Dockerfile")
					hasDockerfile := false
					checkCmd := []string{"test", "-f", dockerfilePath}
					_, checkErr := b.workspace.ExecuteCommand(ctx, checkCmd[0], checkCmd[1:])
					if checkErr == nil {
						hasDockerfile = true
					}

					allServices = append(allServices, ChangedServiceInfo{
						Path:          service,
						HasDockerfile: hasDockerfile,
					})
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

	// Convert map to slice with Dockerfile information
	var changedServices []ChangedServiceInfo
	for service := range serviceMap {
		// Check if the service has a Dockerfile
		dockerfilePath := filepath.Join(b.workspace.Dir(), service, "Dockerfile")
		hasDockerfile := false
		checkCmd := []string{"test", "-f", dockerfilePath}
		_, checkErr := b.workspace.ExecuteCommand(ctx, checkCmd[0], checkCmd[1:])
		if checkErr == nil {
			hasDockerfile = true
		}

		changedServices = append(changedServices, ChangedServiceInfo{
			Path:          service,
			HasDockerfile: hasDockerfile,
		})
	}

	return changedServices, nil
}

// DetectChangedServicesBetweenCommits analyzes git changes between two commits to determine which services need to be rebuilt
func (b *ImageBuilder) DetectChangedServicesBetweenCommits(ctx context.Context, baseCommit, currentCommit string) ([]string, error) {
	// Log the commits being compared
	fmt.Printf("Comparing commits: %s -> %s\n", baseCommit, currentCommit)

	// Ensure baseCommit is before currentCommit
	err := b.ensureCommitOrder(ctx, &baseCommit, &currentCommit)
	if err != nil {
		return nil, fmt.Errorf("failed to validate commit order: %w", err)
	}

	// Fetch the specific commits without creating local branches
	// Use --depth=1 to avoid creating local branches
	fetchCmd := []string{"git", "fetch", "origin", baseCommit + ":" + baseCommit, currentCommit + ":" + currentCommit, "--depth=1"}
	fetchOutput, err := b.workspace.ExecuteCommand(ctx, fetchCmd[0], fetchCmd[1:])
	if err != nil {
		// If fetch fails, try a different approach - fetch all with depth=1
		fetchAllCmd := []string{"git", "fetch", "--all", "--depth=1"}
		fetchAllOutput, err := b.workspace.ExecuteCommand(ctx, fetchAllCmd[0], fetchAllCmd[1:])
		if err != nil {
			return nil, fmt.Errorf("failed to fetch commits: %s\n%w", string(fetchAllOutput), err)
		}
		fmt.Printf("Fetch all output: %s\n", string(fetchAllOutput))
	} else {
		fmt.Printf("Fetch output: %s\n", string(fetchOutput))
	}

	// Get the diff between the two commits
	// Use git diff with explicit commit hashes
	diffCmd := []string{"git", "diff", "--name-only", baseCommit, currentCommit}
	fmt.Printf("Executing command: %s %s\n", diffCmd[0], strings.Join(diffCmd[1:], " "))
	diffOutput, err := b.workspace.ExecuteCommand(ctx, diffCmd[0], diffCmd[1:])
	if err != nil {
		fmt.Printf("Error running git diff: %v\n", err)
		// If diff fails, try an alternative approach using git log
		logCmd := []string{"git", "log", "--name-only", "--pretty=format:", baseCommit + ".." + currentCommit}
		fmt.Printf("Trying alternative command: %s %s\n", logCmd[0], strings.Join(logCmd[1:], " "))
		logOutput, err := b.workspace.ExecuteCommand(ctx, logCmd[0], logCmd[1:])
		if err != nil {
			return nil, fmt.Errorf("failed to get changes between commits: %w", err)
		}
		diffOutput = logOutput
	}

	// Log the output of the git command
	fmt.Printf("Git command output: %s\n", string(diffOutput))

	// Parse the output to get changed files
	changedFiles := strings.Split(string(diffOutput), "\n")
	fmt.Printf("Changed files: %v\n", changedFiles)

	// If no files were found, try another approach
	if len(changedFiles) == 0 || (len(changedFiles) == 1 && changedFiles[0] == "") {
		fmt.Println("No changed files found with first method, trying alternative approach")

		// Try using git show to get the files changed in the current commit
		showCmd := []string{"git", "show", "--name-only", "--pretty=format:", currentCommit}
		fmt.Printf("Trying git show: %s %s\n", showCmd[0], strings.Join(showCmd[1:], " "))
		showOutput, err := b.workspace.ExecuteCommand(ctx, showCmd[0], showCmd[1:])
		if err != nil {
			fmt.Printf("Error running git show: %v\n", err)
			// If all methods fail, return empty list
			return []string{}, nil
		}

		changedFiles = strings.Split(string(showOutput), "\n")
		fmt.Printf("Files from git show: %v\n", changedFiles)

		if len(changedFiles) == 0 || (len(changedFiles) == 1 && changedFiles[0] == "") {
			fmt.Println("No changed files found with any method")
			return []string{}, nil
		}
	}

	// Map to store unique services that need rebuilding
	serviceMap := make(map[string]bool)

	// Analyze changed files to determine affected services
	for _, file := range changedFiles {
		if file == "" {
			continue
		}

		fmt.Printf("Analyzing changed file: %s\n", file)

		// Check if the file is in a service directory
		// This assumes a structure like micro-services/service-name/...
		parts := strings.Split(file, "/")
		if len(parts) >= 2 && parts[0] == "micro-services" {
			servicePath := filepath.Join(parts[0], parts[1])
			fmt.Printf("Found service: %s\n", servicePath)
			serviceMap[servicePath] = true
		}

		// Check for changes in shared libraries or configuration
		if strings.HasPrefix(file, "shared/") || file == "docker-compose.yml" {
			fmt.Println("Found change in shared code or configuration, marking all services as changed")
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

	fmt.Printf("Final list of changed services: %v\n", changedServices)
	return changedServices, nil
}

// ensureCommitOrder ensures that baseCommit is before currentCommit
// If not, it swaps them and returns an error if validation fails
func (b *ImageBuilder) ensureCommitOrder(ctx context.Context, baseCommit, currentCommit *string) error {
	// First, check if both commits exist
	if err := b.checkCommitExists(ctx, *baseCommit); err != nil {
		return fmt.Errorf("base commit does not exist: %w", err)
	}

	if err := b.checkCommitExists(ctx, *currentCommit); err != nil {
		return fmt.Errorf("current commit does not exist: %w", err)
	}

	// Check if the commits are in the correct chronological order
	isBaseBeforeCurrent, err := b.isCommitBefore(ctx, *baseCommit, *currentCommit)
	if err != nil {
		return fmt.Errorf("failed to determine commit order: %w", err)
	}

	if !isBaseBeforeCurrent {
		fmt.Printf("Swapping commits: %s is not before %s\n", *baseCommit, *currentCommit)
		// Swap the commits
		temp := *baseCommit
		*baseCommit = *currentCommit
		*currentCommit = temp
		fmt.Printf("New order: %s -> %s\n", *baseCommit, *currentCommit)
	}

	return nil
}

// checkCommitExists checks if a commit exists in the repository
func (b *ImageBuilder) checkCommitExists(ctx context.Context, commit string) error {
	// Use git rev-parse to check if the commit exists without creating a branch
	cmd := []string{"git", "rev-parse", "--verify", commit + "^{commit}"}
	_, err := b.workspace.ExecuteCommand(ctx, cmd[0], cmd[1:])
	if err != nil {
		return fmt.Errorf("commit %s does not exist: %w", commit, err)
	}
	return nil
}

// isCommitBefore checks if the first commit is chronologically before the second commit
func (b *ImageBuilder) isCommitBefore(ctx context.Context, commit1, commit2 string) (bool, error) {
	// Get the commit dates
	date1, err := b.getCommitDate(ctx, commit1)
	if err != nil {
		return false, err
	}

	date2, err := b.getCommitDate(ctx, commit2)
	if err != nil {
		return false, err
	}

	// Compare the dates
	return date1.Before(date2), nil
}

// getCommitDate gets the commit date for a given commit hash
func (b *ImageBuilder) getCommitDate(ctx context.Context, commit string) (time.Time, error) {
	// Get the commit date using git show
	cmd := []string{"git", "show", "-s", "--format=%ct", commit}
	output, err := b.workspace.ExecuteCommand(ctx, cmd[0], cmd[1:])
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get commit date: %w", err)
	}

	// Parse the timestamp
	timestamp, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse commit timestamp: %w", err)
	}

	// Convert to time.Time
	return time.Unix(timestamp, 0), nil
}

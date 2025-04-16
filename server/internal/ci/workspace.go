package ci

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	//"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"gopkg.in/yaml.v3"
)

func NewWorkspaceFromGit(root, url, branch string) (*workspaceImpl, error) {
	dir, err := os.MkdirTemp(root, "workspace")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// usr, err := user.Current()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get current user: %w", err)
	// }

	// // Try multiple possible SSH key locations
	// possibleKeyPaths := []string{
	// 	filepath.Join(usr.HomeDir, ".ssh", "vnist_e25519"),
	// 	// Add other potential paths if needed
	// }

	// var sshAuth *ssh.PublicKeys
	// var lastErr error

	// for _, keyPath := range possibleKeyPaths {
	// 	if _, err := os.Stat(keyPath); err == nil {
	// 		sshAuth, err = ssh.NewPublicKeysFromFile("git", keyPath, "")
	// 		if err == nil {
	// 			// Successfully loaded the key
	// 			break
	// 		}
	// 		lastErr = err
	// 	}
	// }

	// if sshAuth == nil {
	// 	if lastErr != nil {
	// 		return nil, fmt.Errorf("failed to load any SSH key: %w", lastErr)
	// 	}
	// 	return nil, fmt.Errorf("no SSH keys found in ~/.ssh/id_rsa or ~/.ssh/id_ed25519")
	// }

	// Debug output for troubleshooting
	log.Printf("Cloning repository %s (branch: %s) to %s", url, branch, dir)

	// Use HTTPS URL instead of SSH URL
	//httpURL := strings.Replace(url, "git@github.com:", "https://github.com/", 1)
	//log.Printf("Falling back to HTTPS URL: %s", httpURL)

	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:               url,
		ReferenceName:     plumbing.NewBranchReferenceName(branch),
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Depth:             1,
	})
	if err != nil {
		return nil, fmt.Errorf("git clone failed: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository head: %w", err)
	}

	return &workspaceImpl{
		dir:    dir,
		branch: branch,
		commit: ref.Hash().String(),
		env:    []string{},
	}, nil
}

func NewWorkspaceFromDir(dir string) (*workspaceImpl, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, err
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	return &workspaceImpl{
		dir:    dir,
		branch: ref.Name().Short(),
		commit: ref.Hash().String(),
		env:    []string{},
	}, nil
}

// NewWorkspaceFromPath creates a new workspace from a local path
func NewWorkspaceFromPath(path string) (*workspaceImpl, error) {
	// Check if the path exists
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}

	// Open the repository
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the current branch
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository head: %w", err)
	}

	return &workspaceImpl{
		dir:    path,
		branch: ref.Name().Short(),
		commit: ref.Hash().String(),
		env:    []string{},
	}, nil
}

type workspaceImpl struct {
	branch string
	commit string
	dir    string
	env    []string
}

func (ws *workspaceImpl) Branch() string {
	return ws.branch
}

func (ws *workspaceImpl) Commit() string {
	return ws.commit
}

func (ws *workspaceImpl) Dir() string {
	return ws.dir
}

func (ws *workspaceImpl) Env() []string {
	return ws.env
}

func (ws *workspaceImpl) LoadPipeline(yamlContent []byte) (*Pipeline, error) {
	var pipeline Pipeline
	err := yaml.Unmarshal(yamlContent, &pipeline)
	if err != nil {
		return nil, err
	}
	return &pipeline, nil
}

func (ws *workspaceImpl) ExecuteCommand(ctx context.Context, cmd string, args []string) ([]byte, error) {
	command := exec.CommandContext(ctx, cmd, args...)
	command.Dir = ws.dir
	command.Env = append(command.Environ(), ws.Env()...)

	return command.CombinedOutput()
}

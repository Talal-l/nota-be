package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// HookType represents the type of git hook
type HookType string

const (
	PrePush   HookType = "pre-push"
	PreCommit HookType = "pre-commit"
)

// SetupGitHook installs a git hook of the specified type
func SetupGitHook(hookType HookType) error {
	// Get the module root directory by finding go.mod
	moduleRoot, err := FindModuleRoot()
	if err != nil {
		return fmt.Errorf("error finding module root: %v", err)
	}

	// Create hooks directory
	hooksDir := filepath.Join(moduleRoot, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("error creating hooks directory: %v", err)
	}

	// Compile the hook
	hookBinary := filepath.Join(hooksDir, string(hookType))
	cmd := exec.Command("go", "build", "-o", hookBinary, fmt.Sprintf("./cmd/%s", hookType))
	cmd.Dir = moduleRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error compiling %s hook: %v", hookType, err)
	}

	// Make the binary executable
	if err := os.Chmod(hookBinary, 0755); err != nil {
		return fmt.Errorf("error making %s hook executable: %v", hookType, err)
	}

	return nil
}

package caddy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	ErrBinaryNotFound  = errors.New("caddy binary not found")
	ErrExecutionFailed = errors.New("failed to get Caddy version")
)

type CommandRunner func(ctx context.Context, name string, arg ...string) ([]byte, error)

type SystemVersionProvider interface {
	GetBinaryVersion(ctx context.Context) (string, error)
}

type ExecSystemVersionProvider struct {
	customPath string
	runner     CommandRunner
}

func NewSystemVersionProvider(customPath string) *ExecSystemVersionProvider {
	return &ExecSystemVersionProvider{
		customPath: customPath,
		runner: func(ctx context.Context, name string, arg ...string) ([]byte, error) {
			cmd := exec.CommandContext(ctx, name, arg...)
			return cmd.Output()
		},
	}
}

func NewSystemVersionProviderWithRunner(customPath string, runner CommandRunner) *ExecSystemVersionProvider {
	return &ExecSystemVersionProvider{
		customPath: customPath,
		runner:     runner,
	}
}

func (p *ExecSystemVersionProvider) GetBinaryVersion(ctx context.Context) (string, error) {
	caddyPath, err := p.resolveCaddyPath()
	if err != nil {
		return "", ErrBinaryNotFound
	}

	execCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	output, err := p.runner(execCtx, caddyPath, "version")
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrExecutionFailed, err)
	}

	version := strings.TrimSpace(string(output))
	if version == "" {
		return "", ErrExecutionFailed
	}

	// Extract first line if output contains multiple lines
	if idx := strings.Index(version, "\n"); idx != -1 {
		version = strings.TrimSpace(version[:idx])
	}

	// Normalize e.g. "2.10.2" -> "v2.10.2"
	if !strings.HasPrefix(version, "v") && len(version) > 0 && version[0] >= '0' && version[0] <= '9' {
		version = "v" + version
	}

	return version, nil
}

func (p *ExecSystemVersionProvider) resolveCaddyPath() (string, error) {
	if p.customPath != "" {
		if _, err := os.Stat(p.customPath); err == nil {
			return p.customPath, nil
		}
		return "", ErrBinaryNotFound
	}

	// Check PATH
	if path, err := exec.LookPath("caddy"); err == nil {
		return path, nil
	}

	// Fallback standard binary paths
	wellKnownPaths := []string{
		"/usr/bin/caddy",
		"/usr/local/bin/caddy",
		"/bin/caddy",
	}

	for _, path := range wellKnownPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", ErrBinaryNotFound
}

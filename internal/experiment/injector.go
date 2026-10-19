package experiment

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// ComposeInjector applies a fault by recreating one compose service with overridden environment.
//
// It writes a throwaway compose override file instead of editing .env: the base configuration
// is never touched, so rollback is just "recreate the service from the base compose file",
// and a crashed runner cannot leave a modified .env behind.
type ComposeInjector struct {
	composeFile string
	overrideDir string
}

func NewComposeInjector(composeFile string) (*ComposeInjector, error) {
	abs, err := filepath.Abs(composeFile)
	if err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp("", "flyte-experiment-")
	if err != nil {
		return nil, err
	}
	return &ComposeInjector{composeFile: abs, overrideDir: dir}, nil
}

func (c *ComposeInjector) Inject(ctx context.Context, service string, env map[string]string) error {
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	fmt.Fprintf(&b, "services:\n  %s:\n    environment:\n", service)
	for _, k := range keys {
		fmt.Fprintf(&b, "      %s: %q\n", k, env[k])
	}

	override := filepath.Join(c.overrideDir, "fault.override.yaml")
	if err := os.WriteFile(override, []byte(b.String()), 0o644); err != nil {
		return err
	}
	return c.compose(ctx, []string{c.composeFile, override}, service)
}

// Rollback recreates the service from the base compose file only (compose sees the config
// changed back and recreates the container). Harmless when no fault is active.
func (c *ComposeInjector) Rollback(ctx context.Context, service string) error {
	return c.compose(ctx, []string{c.composeFile}, service)
}

func (c *ComposeInjector) Close() error {
	return os.RemoveAll(c.overrideDir)
}

func (c *ComposeInjector) compose(ctx context.Context, files []string, service string) error {
	args := []string{"compose"}
	for _, f := range files {
		args = append(args, "-f", f)
	}
	// --no-build: env is read at runtime (cleanenv), so no rebuild is needed to change it.
	args = append(args, "up", "-d", "--no-deps", "--no-build", service)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = filepath.Dir(c.composeFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return nil
}

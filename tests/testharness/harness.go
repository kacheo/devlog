// Package testharness provides helpers for subprocess-based devlog CLI tests.
// It builds the devlog binary once per test process (or uses DEVLOG_TEST_BINARY
// to skip the build) and provides an isolated Workspace per test.
package testharness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var (
	buildOnce sync.Once
	builtPath string
	buildErr  error
)

// Build returns the path to the devlog binary, compiling it if necessary.
// Set DEVLOG_TEST_BINARY=/path/to/devlog to skip the build (useful in CI).
func Build(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		if pre := os.Getenv("DEVLOG_TEST_BINARY"); pre != "" {
			builtPath = pre
			return
		}
		tmp, err := os.MkdirTemp("", "devlog-test-bin-*")
		if err != nil {
			buildErr = fmt.Errorf("creating temp dir: %w", err)
			return
		}
		name := "devlog"
		if runtime.GOOS == "windows" {
			name = "devlog.exe"
		}
		builtPath = filepath.Join(tmp, name)
		cmd := exec.Command("go", "build", "-o", builtPath, ".")
		cmd.Dir = moduleRoot()
		if out, err := cmd.CombinedOutput(); err != nil {
			buildErr = fmt.Errorf("go build: %v\n%s", err, out)
		}
	})
	if buildErr != nil {
		t.Fatalf("testharness.Build: %v", buildErr)
	}
	return builtPath
}

// Workspace is an isolated devlog environment backed by a temp directory.
// Each test should create its own Workspace via New.
type Workspace struct {
	Dir    string
	Binary string
	t      *testing.T
}

// New creates a Workspace with an isolated DEVLOG_DIR in a t.TempDir().
func New(t *testing.T) *Workspace {
	t.Helper()
	return &Workspace{Dir: t.TempDir(), Binary: Build(t), t: t}
}

// Run executes devlog with args and returns stdout, stderr, and the exit error.
// A non-nil error means the process exited non-zero.
func (w *Workspace) Run(args ...string) (stdout, stderr string, err error) {
	w.t.Helper()
	cmd := exec.Command(w.Binary, args...)
	cmd.Env = w.Env()
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// MustRun calls Run and fatals if the command exits non-zero.
// Returns stdout for further inspection.
func (w *Workspace) MustRun(args ...string) string {
	w.t.Helper()
	stdout, stderr, err := w.Run(args...)
	if err != nil {
		w.t.Fatalf("devlog %s failed: %v\nstderr: %s", strings.Join(args, " "), err, stderr)
	}
	return stdout
}

// RunJSON runs devlog with --json prepended and unmarshals stdout into v.
func (w *Workspace) RunJSON(v any, args ...string) error {
	w.t.Helper()
	stdout, _, err := w.Run(append([]string{"--json"}, args...)...)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(stdout), v)
}

// Env returns the isolated environment for subprocess invocation.
// All DEVLOG_* vars from the parent process are stripped; DEVLOG_DIR is set to w.Dir.
func (w *Workspace) Env() []string {
	var env []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "DEVLOG_") {
			continue
		}
		env = append(env, e)
	}
	return append(env, "DEVLOG_DIR="+w.Dir)
}

// moduleRoot walks up from this source file to find the directory containing go.mod.
func moduleRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("go.mod not found in any parent directory")
		}
		dir = parent
	}
}

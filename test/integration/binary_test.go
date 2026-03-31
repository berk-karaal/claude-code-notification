package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	binaryPath string
	coverDir   string
)

func TestMain(m *testing.M) {
	// Compile binary to temp directory — reused across all tests
	tmpDir, err := os.MkdirTemp("", "ccn-integration-*")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(tmpDir)

	binaryPath = filepath.Join(tmpDir, "notification")
	// Compile with coverage instrumentation so integration tests contribute to coverage
	cmd := exec.Command("go", "build", "-cover", "-coverpkg=github.com/berk-karaal/claude-code-notification/...", "-o", binaryPath, "github.com/berk-karaal/claude-code-notification/cmd/claude-code-notification")
	if out, err := cmd.CombinedOutput(); err != nil {
		panic("failed to compile binary: " + string(out))
	}

	// INTEGRATION_COVER_DIR: if set, coverage data persists there for merging;
	// otherwise use a temp dir that gets cleaned up.
	coverDir = os.Getenv("INTEGRATION_COVER_DIR")
	if coverDir == "" {
		coverDir, err = os.MkdirTemp("", "ccn-coverage-*")
		if err != nil {
			panic("failed to create coverage dir: " + err.Error())
		}
		defer os.RemoveAll(coverDir)
	}

	os.Exit(m.Run())
}

// runBinary executes the compiled binary with the given stdin and returns stdout, stderr, exit code.
func runBinary(t *testing.T, stdin string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdin = bytes.NewBufferString(stdin)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	// Config isolation: set HOME to temp dir so binary doesn't read developer's real config
	// GOCOVERDIR tells the coverage-instrumented binary where to write coverage profiles
	cmd.Env = append(os.Environ(), "HOME="+t.TempDir(), "GOCOVERDIR="+coverDir)

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

// TestBinary_MalformedJSON_ExitsZero verifies that malformed JSON input causes exit 0
// with an error logged to stderr.
func TestBinary_MalformedJSON_ExitsZero(t *testing.T) {
	_, stderr, code := runBinary(t, `{bad json}`)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stderr, "failed to decode") {
		t.Errorf("expected stderr to contain 'failed to decode', got %q", stderr)
	}
}

// TestBinary_StopHookActive_ExitsZero verifies the stop_hook_active guard.
// When stop_hook_active=true, the binary exits 0 immediately without dispatching.
func TestBinary_StopHookActive_ExitsZero(t *testing.T) {
	input := `{"hook_event_name":"Stop","cwd":"/tmp/proj","stop_hook_active":true}`
	_, _, code := runBinary(t, input)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
}

// TestBinary_Version_OutputsVersionString verifies the --version flag outputs the
// version string set via ldflags. Compiles a separate binary with ldflags.
func TestBinary_Version_OutputsVersionString(t *testing.T) {
	tmpDir := t.TempDir()
	versionBin := filepath.Join(tmpDir, "notification-versioned")
	buildCmd := exec.Command("go", "build", "-cover", "-coverpkg=github.com/berk-karaal/claude-code-notification/...", "-ldflags", "-X main.version=0.0.1-test", "-o", versionBin, "github.com/berk-karaal/claude-code-notification/cmd/claude-code-notification")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build with ldflags failed: %s", string(out))
	}

	cmd := exec.Command(versionBin, "--version")
	cmd.Env = append(os.Environ(), "GOCOVERDIR="+coverDir)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}
	got := strings.TrimSpace(string(out))
	if got != "0.0.1-test" {
		t.Errorf("expected version '0.0.1-test', got %q", got)
	}
}

// TestBinary_ValidStopEvent_ExitsZero verifies that a valid Stop event with
// stop_hook_active=false exits 0. No backends are enabled in isolated config.
func TestBinary_ValidStopEvent_ExitsZero(t *testing.T) {
	cwd := t.TempDir()
	input := `{"hook_event_name":"Stop","cwd":"` + cwd + `","stop_hook_active":false}`
	_, _, code := runBinary(t, input)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
}

// TestBinary_EmptyStdin_ExitsZero verifies that empty stdin causes exit 0
// (decode error handled gracefully).
func TestBinary_EmptyStdin_ExitsZero(t *testing.T) {
	_, _, code := runBinary(t, "")
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
}

// TestBinary_NotificationEvent_ExitsZero verifies that a valid Notification event exits 0.
func TestBinary_NotificationEvent_ExitsZero(t *testing.T) {
	cwd := t.TempDir()
	input := `{"hook_event_name":"Notification","cwd":"` + cwd + `","stop_hook_active":false}`
	_, _, code := runBinary(t, input)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
}

// TestBinary_StopFailureEvent_ExitsZero verifies that a valid StopFailure event exits 0.
func TestBinary_StopFailureEvent_ExitsZero(t *testing.T) {
	cwd := t.TempDir()
	input := `{"hook_event_name":"StopFailure","cwd":"` + cwd + `","stop_hook_active":false}`
	_, _, code := runBinary(t, input)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
}

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAdapterUsageAndHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if code := Run([]string{"adapter"}, &stdout, &stderr); code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "usage="+adapterUsage) {
		t.Fatalf("expected adapter usage output, got %q", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"adapter", "--help"}, &stdout, &stderr); code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "usage="+adapterUsage) {
		t.Fatalf("expected adapter help usage, got %q", stdout.String())
	}
}

func TestRunAdapterSyncHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if code := Run([]string{"adapter", "sync", "--help"}, &stdout, &stderr); code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "usage="+adapterSyncUsage) {
		t.Fatalf("expected adapter sync help usage, got %q", stdout.String())
	}
}

func TestRunAdapterSyncDryRunIsReadOnly(t *testing.T) {
	projectRoot := t.TempDir()
	managedRoot := filepath.Join(projectRoot, ".runecontext", "adapters", "generic", "managed")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--dry-run", "--path", projectRoot, "generic"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["command"], adapterSyncCommand; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if got, want := fields["mutation_performed"], "false"; got != want {
		t.Fatalf("expected mutation_performed %q, got %q", want, got)
	}
	if got, want := fields["network_access"], "false"; got != want {
		t.Fatalf("expected network_access %q, got %q", want, got)
	}
	if _, err := os.Stat(managedRoot); !os.IsNotExist(err) {
		t.Fatalf("expected dry-run to avoid managed-root writes, got err=%v", err)
	}
}

func TestRunAdapterSyncAppliesManagedFilesAndManifest(t *testing.T) {
	projectRoot := t.TempDir()
	userConfigPath := createUserOwnedConfig(t, projectRoot)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--path", projectRoot, "opencode"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["mutation_performed"], "true"; got != want {
		t.Fatalf("expected mutation_performed %q, got %q", want, got)
	}
	if got := fields["changed_file_count"]; got == "0" {
		t.Fatalf("expected changed files on first sync, got %#v", fields)
	}

	managedReadmePath := filepath.Join(projectRoot, ".runecontext", "adapters", "opencode", "managed", "README.md")
	if _, err := os.Stat(managedReadmePath); err != nil {
		t.Fatalf("expected managed adapter README to exist: %v", err)
	}
	assertAdapterManifestConvenience(t, projectRoot)
	assertAdapterSyncBoundaries(t, userConfigPath, projectRoot)

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"adapter", "sync", "--path", projectRoot, "opencode"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected idempotent sync success, got %d (%s)", code, stderr.String())
	}
	fields = parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["changed_file_count"], "0"; got != want {
		t.Fatalf("expected idempotent sync changed_file_count %q, got %q", want, got)
	}
}

func createUserOwnedConfig(t *testing.T, projectRoot string) string {
	t.Helper()
	userConfigPath := filepath.Join(projectRoot, ".opencode", "config.yml")
	if err := os.MkdirAll(filepath.Dir(userConfigPath), 0o755); err != nil {
		t.Fatalf("mkdir user config dir: %v", err)
	}
	if err := os.WriteFile(userConfigPath, []byte("user_owned: true\n"), 0o644); err != nil {
		t.Fatalf("write user config file: %v", err)
	}
	return userConfigPath
}

func assertUserOwnedConfigPreserved(t *testing.T, userConfigPath string) {
	t.Helper()
	userData, err := os.ReadFile(userConfigPath)
	if err != nil {
		t.Fatalf("read user config after sync: %v", err)
	}
	if string(userData) != "user_owned: true\n" {
		t.Fatalf("expected user config boundary to be preserved, got %q", string(userData))
	}
}

func assertAdapterSyncBoundaries(t *testing.T, userConfigPath, projectRoot string) {
	t.Helper()
	assertUserOwnedConfigPreserved(t, userConfigPath)
	if _, err := os.Stat(filepath.Join(projectRoot, "adapters")); !os.IsNotExist(err) {
		t.Fatalf("expected sync to avoid user-owned adapter source tree writes, got err=%v", err)
	}
}

func assertAdapterManifestConvenience(t *testing.T, projectRoot string) {
	t.Helper()
	manifestPath := filepath.Join(projectRoot, ".runecontext", "adapters", "opencode", "sync-manifest.yaml")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("expected adapter manifest to exist: %v", err)
	}
	if !strings.Contains(string(manifestData), "manifest_kind: convenience_metadata") {
		t.Fatalf("expected convenience manifest marker, got %q", string(manifestData))
	}
}

func TestRunAdapterSyncUnknownToolFails(t *testing.T) {
	projectRoot := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--path", projectRoot, "missing-tool"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "adapter \"missing-tool\" not found") {
		t.Fatalf("expected unknown adapter output, got %q", stderr.String())
	}
}

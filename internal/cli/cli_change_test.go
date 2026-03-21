package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunChangeNewCreatesChange(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repoRoot)
	projectRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "change-workflow", "template-project"), projectRoot)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--title", "Add cache invalidation", "--type", "feature", "--size", "small", "--bundle", "base", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["command"], "change_new"; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if got, want := fields["change_mode"], "minimum"; got != want {
		t.Fatalf("expected change_mode %q, got %q", want, got)
	}
	if got := fields["change_id"]; !strings.HasPrefix(got, "CHG-20") {
		t.Fatalf("expected change_id output, got %q", stdout.String())
	}
	if got, want := fields["review_diff_required"], "true"; got != want {
		t.Fatalf("expected review_diff_required %q, got %q", want, got)
	}
	changeDir := filepath.Join(projectRoot, "runecontext", "changes", fields["change_id"])
	if _, err := os.Stat(filepath.Join(changeDir, "proposal.md")); err != nil {
		t.Fatalf("expected proposal.md to exist: %v", err)
	}
}

func TestRunChangeNewExplicitDotPathUsesExplicitRoot(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repoRoot)
	projectRoot, err := os.MkdirTemp(repoRoot, "cli-explicit-root-")
	if err != nil {
		t.Fatalf("mktemp under repo root: %v", err)
	}
	defer os.RemoveAll(projectRoot)
	copyDirForCLI(t, repoFixtureRoot(t, "change-workflow", "template-project"), projectRoot)
	nestedRoot := filepath.Join(projectRoot, "nested")
	if err := os.MkdirAll(nestedRoot, 0o755); err != nil {
		t.Fatalf("mkdir nested root: %v", err)
	}
	t.Chdir(nestedRoot)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--title", "Add cache invalidation", "--type", "feature", "--size", "small", "--bundle", "base", "--path", "."}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected invalid exit code for explicit current-dir root, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "runecontext.yaml") {
		t.Fatalf("expected explicit-root lookup failure mentioning runecontext.yaml, got %q", stderr.String())
	}
}

func TestRunChangeShapeRefreshesStandards(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repoRoot)
	projectRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "change-workflow", "template-project"), projectRoot)
	changeID := runCLIChangeNewForTest(t, projectRoot, "Add cache invalidation")
	statusPath := filepath.Join(projectRoot, "runecontext", "changes", changeID, "status.yaml")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	updated := strings.Replace(string(data), "- base", "- security", 1)
	if err := os.WriteFile(statusPath, []byte(updated), 0o644); err != nil {
		t.Fatalf("write status: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "shape", changeID, "--task", "Implement cache invalidation flow.", "--reference", "docs/cache.md", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["standards_refresh"], "updated"; got != want {
		t.Fatalf("expected standards_refresh %q, got %q", want, got)
	}
	if got, want := fields["added_standard_1"], "standards/security/review.md"; got != want {
		t.Fatalf("expected added standard %q, got %q", want, got)
	}
	if got, want := fields["review_diff_required"], "true"; got != want {
		t.Fatalf("expected review_diff_required %q, got %q", want, got)
	}
	if got, want := fields["changed_file_1_action"], "created"; got != want {
		t.Fatalf("expected first file action %q, got %q", want, got)
	}
}

func TestRunChangeCloseOutputsClosedChange(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repoRoot)
	projectRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "change-workflow", "template-project"), projectRoot)
	changeID := runCLIChangeNewForTest(t, projectRoot, "Add cache invalidation")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "close", changeID, "--verification-status", "passed", "--closed-at", "2026-03-20", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["change_status"], "closed"; got != want {
		t.Fatalf("expected change_status %q, got %q", want, got)
	}
	if got, want := fields["closed_at"], "2026-03-20"; got != want {
		t.Fatalf("expected closed_at %q, got %q", want, got)
	}
}

func TestRunChangeReallocateOutputsNewChangeID(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	changeID := runCLIChangeNewForTest(t, projectRoot, "Add cache invalidation")
	appendCLIProposalSelfReference(t, projectRoot, changeID)
	fields := runCLIChangeReallocate(t, projectRoot, changeID)
	newID := assertCLIReallocateFields(t, fields, changeID)
	assertCLIReallocatedProposal(t, projectRoot, changeID, newID)
}

func TestRunChangeReallocateUsageErrors(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "reallocate"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code for missing ID, got %d", code)
	}
	if !strings.Contains(stderr.String(), "change reallocate requires exactly one change ID") {
		t.Fatalf("expected missing-ID usage output, got %q", stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"change", "reallocate", "CHG-2026-001-a3f2-auth-gateway", "--bogus"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code for unknown flag, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown change reallocate flag") {
		t.Fatalf("expected unknown-flag output, got %q", stderr.String())
	}
}

func TestRunChangeNewRejectsMissingValueBeforeNextFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"change", "new", "--title", "--type", "feature"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code for missing title value, got %d", code)
	}
	if !strings.Contains(stderr.String(), "--title requires a value") {
		t.Fatalf("expected missing-value output, got %q", stderr.String())
	}
}

func TestRunChangeNewDryRunDoesNotPersistChange(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--title", "Dry run change", "--type", "feature", "--size", "small", "--bundle", "base", "--dry-run", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["dry_run"], "true"; got != want {
		t.Fatalf("expected dry_run %q, got %q", want, got)
	}
	changeID := fields["change_id"]
	if changeID == "" {
		t.Fatalf("expected change_id in dry-run output, got %q", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "runecontext", "changes", changeID)); !os.IsNotExist(err) {
		t.Fatalf("expected dry-run to avoid persisted change, got err=%v", err)
	}
}

func TestRunStatusOutputsCounts(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	firstID := runCLIChangeNewForTest(t, projectRoot, "Add cache invalidation")
	secondID := runCLIChangeNewForTest(t, projectRoot, "Revise cache invalidation")
	runCLIChangeClose(t, projectRoot, firstID, []string{"--verification-status", "skipped", "--superseded-by", secondID, "--closed-at", "2026-03-20", "--path", projectRoot})
	runCLIChangeClose(t, projectRoot, secondID, []string{"--verification-status", "passed", "--closed-at", "2026-03-21", "--path", projectRoot})
	fields := runCLIStatus(t, projectRoot)
	if got, want := fields["active_count"], "0"; got != want {
		t.Fatalf("expected active_count %q, got %q", want, got)
	}
	if got, want := fields["closed_count"], "1"; got != want {
		t.Fatalf("expected closed_count %q, got %q", want, got)
	}
	if got, want := fields["superseded_count"], "1"; got != want {
		t.Fatalf("expected superseded_count %q, got %q", want, got)
	}
	if got, want := fields["superseded_1_id"], firstID; got != want {
		t.Fatalf("expected superseded change %q, got %q", want, got)
	}
	if got, want := fields["closed_1_id"], secondID; got != want {
		t.Fatalf("expected closed change %q, got %q", want, got)
	}
}

func prepareCLIWorkflowProject(t *testing.T) string {
	t.Helper()
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repoRoot)
	projectRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "change-workflow", "template-project"), projectRoot)
	return projectRoot
}

func appendCLIProposalSelfReference(t *testing.T, projectRoot, changeID string) {
	t.Helper()
	proposalPath := filepath.Join(projectRoot, "runecontext", "changes", changeID, "proposal.md")
	data, err := os.ReadFile(proposalPath)
	if err != nil {
		t.Fatalf("read proposal: %v", err)
	}
	updated := strings.ReplaceAll(string(data), "\r\n", "\n") + "\nSee changes/" + changeID + "/proposal.md#summary for the current change summary.\n"
	if err := os.WriteFile(proposalPath, []byte(updated), 0o644); err != nil {
		t.Fatalf("write proposal: %v", err)
	}
}

func runCLIChangeReallocate(t *testing.T, projectRoot, changeID string) map[string]string {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "reallocate", changeID, "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	return parseCLIKeyValueOutput(t, stdout.String())
}

func assertCLIReallocateFields(t *testing.T, fields map[string]string, changeID string) string {
	t.Helper()
	if got, want := fields["command"], "change_reallocate"; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if got, want := fields["old_change_id"], changeID; got != want {
		t.Fatalf("expected old_change_id %q, got %q", want, got)
	}
	newID := fields["change_id"]
	if newID == "" || newID == changeID {
		t.Fatalf("expected a new change ID, got %#v", fields)
	}
	if got := fields["rewritten_reference_count"]; got != "1" {
		t.Fatalf("expected one rewritten reference, got %q", got)
	}
	if got := fields["warning_count"]; got != "0" {
		t.Fatalf("expected no warnings, got %q", got)
	}
	return newID
}

func assertCLIReallocatedProposal(t *testing.T, projectRoot, oldID, newID string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(projectRoot, "runecontext", "changes", oldID)); !os.IsNotExist(err) {
		t.Fatalf("expected old change directory to be removed, got err=%v", err)
	}
	proposalData, err := os.ReadFile(filepath.Join(projectRoot, "runecontext", "changes", newID, "proposal.md"))
	if err != nil {
		t.Fatalf("read reallocated proposal: %v", err)
	}
	if !strings.Contains(string(proposalData), "changes/"+newID+"/proposal.md#summary") {
		t.Fatalf("expected CLI reallocation to rewrite local reference, got:\n%s", string(proposalData))
	}
}

func runCLIChangeClose(t *testing.T, projectRoot, changeID string, args []string) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	fullArgs := append([]string{"change", "close", changeID}, args...)
	if code := Run(fullArgs, &stdout, &stderr); code != 0 {
		t.Fatalf("change close failed: %d (%s)", code, stderr.String())
	}
}

func runCLIStatus(t *testing.T, projectRoot string) map[string]string {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"status", projectRoot}, &stdout, &stderr); code != 0 {
		t.Fatalf("status command failed: %d (%s)", code, stderr.String())
	}
	return parseCLIKeyValueOutput(t, stdout.String())
}

func TestRunStatusRejectsUnknownFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"status", "--bogus"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code for unknown flag, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown status flag") {
		t.Fatalf("expected unknown status flag output, got %q", stderr.String())
	}
}

func TestRunStatusRejectsDryRunFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"status", "--dry-run"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code for unsupported flag, got %d", code)
	}
	if !strings.Contains(stderr.String(), "--dry-run is only supported for write commands") {
		t.Fatalf("expected unsupported-flag output, got %q", stderr.String())
	}
}

func TestRunStatusJSONOutputEnvelope(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"status", "--json", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	var payload struct {
		SchemaVersion int               `json:"schema_version"`
		Result        string            `json:"result"`
		Command       string            `json:"command"`
		ExitCode      int               `json:"exit_code"`
		FailureClass  string            `json:"failure_class"`
		Data          map[string]string `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON output, got err=%v payload=%q", err, stdout.String())
	}
	if payload.SchemaVersion != 1 || payload.Result != "ok" || payload.Command != "status" || payload.ExitCode != 0 || payload.FailureClass != "none" {
		t.Fatalf("unexpected JSON envelope: %#v", payload)
	}
	if payload.Data["result"] != "ok" || payload.Data["command"] != "status" {
		t.Fatalf("expected command data fields in JSON output, got %#v", payload.Data)
	}
}

func TestRunValidateRejectsInvalidProposal(t *testing.T) {
	root := fixtureRoot(t, "reject-proposal-invalid")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", root}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected validation failure exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "error_path=") || !strings.Contains(stderr.String(), "proposal.md") {
		t.Fatalf("expected proposal path in output, got %q", stderr.String())
	}
}

func TestDryRunRejectsExternalSymlink(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	externalDir := t.TempDir()
	externalFile := filepath.Join(externalDir, "external.txt")
	if err := os.WriteFile(externalFile, []byte("external content"), 0o644); err != nil {
		t.Fatalf("write external file: %v", err)
	}
	symlinkPath := filepath.Join(projectRoot, "runecontext", "standards", "external-link.md")
	if err := os.Symlink(externalFile, symlinkPath); err != nil {
		t.Fatalf("create external symlink: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--title", "Symlink test", "--type", "feature", "--size", "small", "--bundle", "base", "--dry-run", "--path", projectRoot}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected failure for external symlink, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "rejects") && !strings.Contains(stderr.String(), "outside project root") {
		t.Fatalf("expected symlink rejection error, got %q", stderr.String())
	}
}

func TestDryRunRejectsAbsoluteSymlink(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	targetFile := filepath.Join(projectRoot, "runecontext", "standards", "project-review.md")
	symlinkPath := filepath.Join(projectRoot, "runecontext", "standards", "absolute-link.md")
	if err := os.Symlink(targetFile, symlinkPath); err != nil {
		t.Fatalf("create absolute symlink: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--title", "Absolute symlink test", "--type", "feature", "--size", "small", "--bundle", "base", "--dry-run", "--path", projectRoot}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected failure for absolute symlink, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "absolute symlink") {
		t.Fatalf("expected absolute symlink rejection, got %q", stderr.String())
	}
}

func TestDryRunPreservesDiscoverySemantics(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repoRoot)
	projectRoot, err := os.MkdirTemp(repoRoot, "cli-nearest-ancestor-")
	if err != nil {
		t.Fatalf("mktemp under repo root: %v", err)
	}
	defer os.RemoveAll(projectRoot)
	copyDirForCLI(t, repoFixtureRoot(t, "change-workflow", "template-project"), projectRoot)
	nestedRoot := filepath.Join(projectRoot, "nested")
	if err := os.MkdirAll(nestedRoot, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	t.Chdir(nestedRoot)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--title", "Project root dry run", "--type", "feature", "--size", "small", "--bundle", "base", "--dry-run"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success for nearest-ancestor dry-run, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if fields["dry_run"] != "true" {
		t.Fatalf("expected dry_run=true, got %q", fields["dry_run"])
	}
	if fields["root"] != nestedRoot {
		t.Fatalf("expected invocation root %q, got %q", nestedRoot, fields["root"])
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "runecontext", "changes", fields["change_id"])); !os.IsNotExist(err) {
		t.Fatalf("expected dry-run not to persist nearest-ancestor change, got err=%v", err)
	}
}

func TestDryRunCloneRespectsFileLimit(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	original := dryRunCloneLimits
	dryRunCloneLimits.MaxFiles = 1
	t.Cleanup(func() { dryRunCloneLimits = original })
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--title", "Limit test", "--type", "feature", "--size", "small", "--bundle", "base", "--dry-run", "--path", projectRoot}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected invalid exit code on file limit, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "maximum file count") {
		t.Fatalf("expected max-file limit error, got %q", stderr.String())
	}
}

func TestChangeNewJSONOutputEnvelope(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--json", "--title", "JSON test", "--type", "feature", "--size", "small", "--bundle", "base", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	var payload struct {
		SchemaVersion int               `json:"schema_version"`
		Result        string            `json:"result"`
		Command       string            `json:"command"`
		ExitCode      int               `json:"exit_code"`
		FailureClass  string            `json:"failure_class"`
		Data          map[string]string `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON output, got err=%v payload=%q", err, stdout.String())
	}
	if payload.SchemaVersion != 1 || payload.Result != "ok" || payload.Command != "change_new" || payload.ExitCode != 0 {
		t.Fatalf("unexpected JSON envelope: %#v", payload)
	}
	if payload.Data["change_id"] == "" {
		t.Fatalf("expected change_id in JSON data, got %#v", payload.Data)
	}
}

func TestChangeInvalidJSONOutputEnvelope(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--json", "--type", "feature"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	var payload struct {
		SchemaVersion int               `json:"schema_version"`
		Result        string            `json:"result"`
		Command       string            `json:"command"`
		ExitCode      int               `json:"exit_code"`
		FailureClass  string            `json:"failure_class"`
		Data          map[string]string `json:"data"`
	}
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON error output, got err=%v payload=%q", err, stderr.String())
	}
	if payload.FailureClass != "usage" {
		t.Fatalf("expected failure_class=usage, got %q", payload.FailureClass)
	}
}

func TestExplainWarningWhenNotImplemented(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"status", "--explain", "--json", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	var payload struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON output: %v", err)
	}
	if payload.Data["explain"] != "true" {
		t.Fatalf("expected explain=true, got %q", payload.Data["explain"])
	}
	if payload.Data["explain_warning"] == "" {
		t.Fatalf("expected explain_warning when --explain not implemented, got %q", payload.Data)
	}
}

func TestMachineFlagsInterleavedAcrossArgs(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--title", "Interleaved flags", "--json", "--type", "feature", "--non-interactive", "--size", "small", "--bundle", "base", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success with interleaved machine flags, got %d (%s)", code, stderr.String())
	}
	var payload struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON output: %v", err)
	}
	if payload.Data["non_interactive"] != "true" {
		t.Fatalf("expected non_interactive=true, got %q", payload.Data["non_interactive"])
	}
}

func TestAssertUniqueKeysInOutput(t *testing.T) {
	if assertUniqueKeys([]line{{key: "a", value: "1"}, {key: "a", value: "2"}}) {
		t.Fatalf("expected duplicate keys to be rejected")
	}
	if !assertUniqueKeys([]line{{key: "a", value: "1"}, {key: "b", value: "2"}}) {
		t.Fatalf("expected unique keys to be accepted")
	}
	if duplicateKey, ok := firstDuplicateKey([]line{{key: "x", value: "1"}, {key: "x", value: "2"}}); !ok || duplicateKey != "x" {
		t.Fatalf("expected duplicate key x, got key=%q ok=%t", duplicateKey, ok)
	}
}

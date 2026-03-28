package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func TestRunStatusHumanOutputUsesSectionedAsciiLayout(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	firstID := runCLIChangeNewForTest(t, projectRoot, "Add cache invalidation")
	secondID := runCLIChangeNewForTest(t, projectRoot, "Revise cache invalidation")
	runCLIChangeClose(t, projectRoot, firstID, []string{"--verification-status", "skipped", "--superseded-by", secondID, "--closed-at", "2026-03-20", "--path", projectRoot})
	runCLIChangeClose(t, projectRoot, secondID, []string{"--verification-status", "passed", "--closed-at", "2026-03-21", "--path", projectRoot})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"status", projectRoot}, &stdout, &stderr); code != 0 {
		t.Fatalf("status command failed: %d (%s)", code, stderr.String())
	}
	out := stdout.String()
	for _, token := range []string{
		"RuneContext Status",
		"In Flight (0)",
		"Recently Completed (1)",
		"Replaced (1)",
		secondID,
		firstID,
	} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected human status output to contain %q, got:\n%s", token, out)
		}
	}
	if strings.Index(out, secondID) > strings.Index(out, firstID) {
		t.Fatalf("expected more recent closed entry to appear first, got:\n%s", out)
	}
	if strings.Contains(out, "|--") || strings.Contains(out, "`--") {
		t.Fatalf("expected default non-verbose output to avoid detail trees, got:\n%s", out)
	}
	if strings.Contains(out, "result=") || strings.Contains(out, "active_count=") {
		t.Fatalf("expected human renderer output, got key=value contract dump:\n%s", out)
	}
}

func TestRunStatusHistoryRecentUsesDefaultBoundedPreview(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	ids := []string{
		runCLIChangeNewForTest(t, projectRoot, "One"),
		runCLIChangeNewForTest(t, projectRoot, "Two"),
		runCLIChangeNewForTest(t, projectRoot, "Three"),
	}
	for i, id := range ids {
		date := []string{"2026-03-20", "2026-03-21", "2026-03-22"}[i]
		runCLIChangeClose(t, projectRoot, id, []string{"--verification-status", "passed", "--closed-at", date, "--path", projectRoot})
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"status", "--history-limit", "2", projectRoot}, &stdout, &stderr); code != 0 {
		t.Fatalf("status command failed: %d (%s)", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "showing 2 of 3 closed changes; use --history all to show more") {
		t.Fatalf("expected hidden-history hint, got:\n%s", out)
	}
	if strings.Contains(out, ids[0]) {
		t.Fatalf("expected oldest closed entry to be hidden by default recent mode, got:\n%s", out)
	}
}

func TestRunStatusHistoryAllShowsAllEntries(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	ids := []string{
		runCLIChangeNewForTest(t, projectRoot, "One"),
		runCLIChangeNewForTest(t, projectRoot, "Two"),
	}
	for i, id := range ids {
		date := []string{"2026-03-20", "2026-03-21"}[i]
		runCLIChangeClose(t, projectRoot, id, []string{"--verification-status", "passed", "--closed-at", date, "--path", projectRoot})
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"status", "--history", "all", projectRoot}, &stdout, &stderr); code != 0 {
		t.Fatalf("status command failed: %d (%s)", code, stderr.String())
	}
	out := stdout.String()
	for _, id := range ids {
		if !strings.Contains(out, id) {
			t.Fatalf("expected history all to include %q, got:\n%s", id, out)
		}
	}
	if strings.Contains(out, "showing ") {
		t.Fatalf("expected no hidden-history hint in all mode, got:\n%s", out)
	}
}

func TestRunStatusHistoryNoneHidesHistoricalSections(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	id := runCLIChangeNewForTest(t, projectRoot, "One")
	runCLIChangeClose(t, projectRoot, id, []string{"--verification-status", "passed", "--closed-at", "2026-03-20", "--path", projectRoot})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"status", "--history", "none", projectRoot}, &stdout, &stderr); code != 0 {
		t.Fatalf("status command failed: %d (%s)", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "Recently Completed (0)") {
		t.Fatalf("expected hidden closed section count, got:\n%s", out)
	}
	if !strings.Contains(out, "showing 0 of 1 closed changes; use --history all to show more") {
		t.Fatalf("expected hidden-history hint for none mode, got:\n%s", out)
	}
}

func TestRunStatusVerboseShowsRelationshipDetails(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	firstID := runCLIChangeNewForTest(t, projectRoot, "Add cache invalidation")
	secondID := runCLIChangeNewForTest(t, projectRoot, "Revise cache invalidation")
	runCLIChangeClose(t, projectRoot, firstID, []string{"--verification-status", "skipped", "--superseded-by", secondID, "--closed-at", "2026-03-20", "--path", projectRoot})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"status", "--verbose", projectRoot}, &stdout, &stderr); code != 0 {
		t.Fatalf("status command failed: %d (%s)", code, stderr.String())
	}
	out := stdout.String()
	for _, token := range []string{"|-- superseded by:", "`-- path:"} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected verbose tree details token %q, got:\n%s", token, out)
		}
	}
}

func TestRenderHumanStatusColorToggleIsDeterministic(t *testing.T) {
	projectRoot := fixtureRoot(t, "valid-project")
	absRoot, validator, loaded, err := loadProjectForCLI(projectRoot, true)
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	defer loaded.Close()

	summary, err := contracts.BuildProjectStatusSummary(validator, loaded)
	if err != nil {
		t.Fatalf("build status summary: %v", err)
	}

	ascii := renderHumanStatus(absRoot, loaded, summary, statusRenderOptions{color: false})
	if strings.Contains(ascii, "\x1b[") {
		t.Fatalf("expected ASCII-only rendering without ANSI escapes, got:\n%s", ascii)
	}
	colored := renderHumanStatus(absRoot, loaded, summary, statusRenderOptions{color: true})
	if !strings.Contains(colored, "\x1b[") {
		t.Fatalf("expected ANSI escapes when color is enabled, got:\n%s", colored)
	}
}

func TestBuildStatusSummaryProvidesRelationshipMetadataForRenderer(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, fixtureRoot(t, "valid-project"), root)
	absRoot, validator, loaded, err := loadProjectForCLI(root, true)
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	defer loaded.Close()
	if _, err := contracts.CloseChange(validator, loaded, "CHG-2026-001-a3f2-auth-gateway", contracts.ChangeCloseOptions{
		VerificationStatus: "skipped",
		ClosedAt:           time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		SupersededBy:       []string{"CHG-2026-002-b4c3-auth-revision"},
	}); err != nil {
		t.Fatalf("close change: %v", err)
	}
	loaded.Close()

	_, _, reloaded, err := loadProjectForCLI(absRoot, true)
	if err != nil {
		t.Fatalf("reload project: %v", err)
	}
	defer reloaded.Close()
	summary, err := contracts.BuildProjectStatusSummary(validator, reloaded)
	if err != nil {
		t.Fatalf("build summary: %v", err)
	}
	out := renderHumanStatus(absRoot, reloaded, summary, statusRenderOptions{color: false, verbose: true})
	for _, token := range []string{
		"depends on:",
		"related:",
		"superseded by:",
		"created:",
		"closed:",
	} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected rendered relationship/recency token %q, got:\n%s", token, out)
		}
	}
}

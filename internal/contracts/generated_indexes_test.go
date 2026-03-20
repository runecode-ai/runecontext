package contracts

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBuildGeneratedManifestMatchesGolden(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "traceability", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()

	manifest, err := index.BuildGeneratedManifest()
	if err != nil {
		t.Fatalf("build generated manifest: %v", err)
	}
	assertGeneratedArtifactValidAgainstSchema(t, v, "manifest.schema.json", "generated-manifest.yaml", manifest)
	assertGeneratedArtifactMatchesGolden(t, manifest, fixturePath(t, "generated-indexes", "golden", "traceability-manifest.yaml"))
}

func TestBuildGeneratedChangesByStatusIndexMatchesGolden(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "traceability", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()

	changeIndex, err := index.BuildGeneratedChangesByStatusIndex()
	if err != nil {
		t.Fatalf("build changes-by-status index: %v", err)
	}
	assertGeneratedArtifactValidAgainstSchema(t, v, "changes-by-status-index.schema.json", "generated-changes-by-status.yaml", changeIndex)
	assertGeneratedArtifactMatchesGolden(t, changeIndex, fixturePath(t, "generated-indexes", "golden", "traceability-changes-by-status.yaml"))
}

func TestBuildGeneratedBundlesIndexMatchesGolden(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()

	bundleIndex, err := index.BuildGeneratedBundlesIndex()
	if err != nil {
		t.Fatalf("build generated bundle index: %v", err)
	}
	assertGeneratedArtifactValidAgainstSchema(t, v, "bundles-index.schema.json", "generated-bundles-index.yaml", bundleIndex)
	assertGeneratedArtifactMatchesGolden(t, bundleIndex, fixturePath(t, "generated-indexes", "golden", "bundle-resolution-bundles.yaml"))
}

func TestBuildGeneratedIndexesDeterministic(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "traceability", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()

	firstManifest, err := index.BuildGeneratedManifest()
	if err != nil {
		t.Fatalf("build first manifest: %v", err)
	}
	secondManifest, err := index.BuildGeneratedManifest()
	if err != nil {
		t.Fatalf("build second manifest: %v", err)
	}
	if !reflect.DeepEqual(firstManifest, secondManifest) {
		t.Fatalf("expected deterministic manifest output\nfirst:  %#v\nsecond: %#v", firstManifest, secondManifest)
	}

	firstChanges, err := index.BuildGeneratedChangesByStatusIndex()
	if err != nil {
		t.Fatalf("build first changes-by-status index: %v", err)
	}
	secondChanges, err := index.BuildGeneratedChangesByStatusIndex()
	if err != nil {
		t.Fatalf("build second changes-by-status index: %v", err)
	}
	if !reflect.DeepEqual(firstChanges, secondChanges) {
		t.Fatalf("expected deterministic changes-by-status output\nfirst:  %#v\nsecond: %#v", firstChanges, secondChanges)
	}

	bundleIndex, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate bundle fixture project: %v", err)
	}
	defer bundleIndex.Close()

	firstBundles, err := bundleIndex.BuildGeneratedBundlesIndex()
	if err != nil {
		t.Fatalf("build first bundles index: %v", err)
	}
	secondBundles, err := bundleIndex.BuildGeneratedBundlesIndex()
	if err != nil {
		t.Fatalf("build second bundles index: %v", err)
	}
	if !reflect.DeepEqual(firstBundles, secondBundles) {
		t.Fatalf("expected deterministic bundles output\nfirst:  %#v\nsecond: %#v", firstBundles, secondBundles)
	}
}

func TestWriteGeneratedIndexesWritesStandardPaths(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(root)
	if err != nil {
		t.Fatalf("validate copied fixture project: %v", err)
	}
	defer index.Close()

	if err := index.WriteGeneratedIndexes(); err != nil {
		t.Fatalf("write generated indexes: %v", err)
	}

	paths := []struct {
		path   string
		schema string
	}{
		{path: filepath.Join(index.ContentRoot, "manifest.yaml"), schema: "manifest.schema.json"},
		{path: filepath.Join(index.ContentRoot, "indexes", "changes-by-status.yaml"), schema: "changes-by-status-index.schema.json"},
		{path: filepath.Join(index.ContentRoot, "indexes", "bundles.yaml"), schema: "bundles-index.schema.json"},
	}
	for _, item := range paths {
		data, err := os.ReadFile(item.path)
		if err != nil {
			t.Fatalf("read generated file %s: %v", item.path, err)
		}
		if err := v.ValidateYAMLFile(item.schema, item.path, data); err != nil {
			t.Fatalf("expected generated file to satisfy %s: %v", item.schema, err)
		}
	}
}

func TestGeneratedIndexSchemasRejectUnknownFields(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "traceability", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()

	manifest, err := index.BuildGeneratedManifest()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	manifestData, err := yaml.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	manifestValue, err := parseYAML(manifestData)
	if err != nil {
		t.Fatalf("parse manifest yaml: %v", err)
	}
	manifestMap := manifestValue.(map[string]any)
	manifestMap["unexpected"] = "value"
	if err := v.ValidateValue("manifest.schema.json", "generated-manifest.yaml", manifestMap); err == nil {
		t.Fatal("expected manifest schema to reject unknown fields")
	}

	changesByStatus, err := index.BuildGeneratedChangesByStatusIndex()
	if err != nil {
		t.Fatalf("build changes index: %v", err)
	}
	changesData, err := yaml.Marshal(changesByStatus)
	if err != nil {
		t.Fatalf("marshal changes index: %v", err)
	}
	changesValue, err := parseYAML(changesData)
	if err != nil {
		t.Fatalf("parse changes index yaml: %v", err)
	}
	changesMap := changesValue.(map[string]any)
	statuses := changesMap["statuses"].(map[string]any)
	proposed := statuses["proposed"].([]any)
	if len(proposed) == 0 {
		t.Fatal("expected fixture proposed status entries")
	}
	proposedEntry := proposed[0].(map[string]any)
	proposedEntry["unexpected"] = true
	if err := v.ValidateValue("changes-by-status-index.schema.json", "generated-changes-by-status.yaml", changesMap); err == nil {
		t.Fatal("expected changes-by-status schema to reject unknown entry fields")
	}

	bundleIndexProject, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate bundle fixture project: %v", err)
	}
	defer bundleIndexProject.Close()
	bundlesIndex, err := bundleIndexProject.BuildGeneratedBundlesIndex()
	if err != nil {
		t.Fatalf("build bundles index: %v", err)
	}
	bundlesData, err := yaml.Marshal(bundlesIndex)
	if err != nil {
		t.Fatalf("marshal bundles index: %v", err)
	}
	bundlesValue, err := parseYAML(bundlesData)
	if err != nil {
		t.Fatalf("parse bundles index yaml: %v", err)
	}
	bundlesMap := bundlesValue.(map[string]any)
	bundlesMap["unexpected"] = true
	if err := v.ValidateValue("bundles-index.schema.json", "generated-bundles-index.yaml", bundlesMap); err == nil {
		t.Fatal("expected bundles schema to reject unknown top-level fields")
	}
	delete(bundlesMap, "unexpected")
	bundles := bundlesMap["bundles"].([]any)
	if len(bundles) == 0 {
		t.Fatal("expected bundle entries")
	}
	bundleEntry := bundles[0].(map[string]any)
	bundleEntry["unexpected"] = "value"
	if err := v.ValidateValue("bundles-index.schema.json", "generated-bundles-index.yaml", bundlesMap); err == nil {
		t.Fatal("expected bundles schema to reject unknown bundle entry fields")
	}
	delete(bundleEntry, "unexpected")
	referencedPatterns := bundleEntry["referenced_patterns"].(map[string]any)
	projectPatterns := referencedPatterns["project"].(map[string]any)
	includes := projectPatterns["includes"].([]any)
	if len(includes) == 0 {
		t.Fatal("expected project includes entries")
	}
	firstInclude := includes[0].(map[string]any)
	firstInclude["unexpected"] = "value"
	if err := v.ValidateValue("bundles-index.schema.json", "generated-bundles-index.yaml", bundlesMap); err == nil {
		t.Fatal("expected bundles schema to reject unknown nested pattern fields")
	}
}

func TestBuildGeneratedChangesByStatusIndexRejectsUnknownLifecycleStatus(t *testing.T) {
	root := t.TempDir()
	index := &ProjectIndex{
		ContentRoot: root,
		Changes: map[string]*ChangeRecord{
			"CHG-2026-123-a1b2-unknown": {
				ID:         "CHG-2026-123-a1b2-unknown",
				Title:      "Unknown lifecycle",
				Type:       "feature",
				Status:     LifecycleStatus("unknown"),
				StatusPath: filepath.Join(root, "changes", "CHG-2026-123-a1b2-unknown", "status.yaml"),
			},
		},
	}
	_, err := index.BuildGeneratedChangesByStatusIndex()
	if err == nil || !strings.Contains(err.Error(), "unsupported lifecycle status") {
		t.Fatalf("expected unsupported lifecycle status failure, got %v", err)
	}
}

func TestGeneratedRelativeArtifactPathRejectsEscapes(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "outside.yaml")
	_, err := generatedRelativeArtifactPath(root, outside)
	if err == nil || !strings.Contains(err.Error(), "escapes RuneContext content root") {
		t.Fatalf("expected escape rejection, got %v", err)
	}
}

func TestBuildGeneratedChangesByStatusIndexRejectsPathOutsideContentRoot(t *testing.T) {
	root := t.TempDir()
	outsideRoot := t.TempDir()
	index := &ProjectIndex{
		ContentRoot: root,
		Changes: map[string]*ChangeRecord{
			"CHG-2026-124-a1b2-outside": {
				ID:         "CHG-2026-124-a1b2-outside",
				Title:      "Outside path",
				Type:       "feature",
				Status:     StatusProposed,
				StatusPath: filepath.Join(outsideRoot, "changes", "CHG-2026-124-a1b2-outside", "status.yaml"),
			},
		},
	}
	_, err := index.BuildGeneratedChangesByStatusIndex()
	if err == nil || !strings.Contains(err.Error(), "escapes RuneContext content root") {
		t.Fatalf("expected out-of-root path rejection, got %v", err)
	}
}

func TestGeneratedRelativeArtifactPathReturnsSlashCanonicalRelativePaths(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "changes", "CHG-2026-125-a1b2-canonical", "status.yaml")
	rel, err := generatedRelativeArtifactPath(root, target)
	if err != nil {
		t.Fatalf("generated relative path: %v", err)
	}
	want := "changes/CHG-2026-125-a1b2-canonical/status.yaml"
	if rel != want {
		t.Fatalf("expected %s, got %s", want, rel)
	}
}

func assertGeneratedArtifactValidAgainstSchema(t *testing.T, v *Validator, schema, path string, doc any) {
	t.Helper()
	data, err := yaml.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal generated artifact: %v", err)
	}
	if err := v.ValidateYAMLFile(schema, path, data); err != nil {
		t.Fatalf("expected generated artifact to satisfy schema %s: %v\n%s", schema, err, string(data))
	}
}

func assertGeneratedArtifactMatchesGolden(t *testing.T, doc any, goldenPath string) {
	t.Helper()
	data, err := yaml.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal generated artifact: %v", err)
	}
	goldenData, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("missing golden fixture %s\n%s", goldenPath, string(data))
		}
		t.Fatalf("read golden fixture %s: %v", goldenPath, err)
	}
	expected := normalizeResolutionValue(t, mustParseYAML(t, string(goldenData)))
	actual := normalizeResolutionValue(t, mustParseYAML(t, string(data)))
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("generated artifact mismatch\nexpected: %#v\nactual:   %#v\nactual_yaml:\n%s", expected, actual, string(data))
	}
}

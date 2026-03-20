package contracts

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	generatedManifestSchemaVersion    = 1
	generatedChangeIndexSchemaVersion = 1
	generatedBundleIndexSchemaVersion = 1
	generatedManifestRelativePath     = "manifest.yaml"
	generatedChangesIndexRelativePath = "indexes/changes-by-status.yaml"
	generatedBundlesIndexRelativePath = "indexes/bundles.yaml"
	generatedIndexesDirectoryRelative = "indexes"
)

type GeneratedManifest struct {
	SchemaVersion int                      `json:"schema_version" yaml:"schema_version"`
	Indexes       GeneratedManifestIndexes `json:"indexes" yaml:"indexes"`
	Counts        GeneratedManifestCounts  `json:"counts" yaml:"counts"`
	Standards     []string                 `json:"standards" yaml:"standards"`
	Bundles       []string                 `json:"bundles" yaml:"bundles"`
	Changes       []string                 `json:"changes" yaml:"changes"`
	Specs         []string                 `json:"specs" yaml:"specs"`
	Decisions     []string                 `json:"decisions" yaml:"decisions"`
}

type GeneratedManifestIndexes struct {
	ChangesByStatus string `json:"changes_by_status" yaml:"changes_by_status"`
	Bundles         string `json:"bundles" yaml:"bundles"`
}

type GeneratedManifestCounts struct {
	Standards int `json:"standards" yaml:"standards"`
	Bundles   int `json:"bundles" yaml:"bundles"`
	Changes   int `json:"changes" yaml:"changes"`
	Specs     int `json:"specs" yaml:"specs"`
	Decisions int `json:"decisions" yaml:"decisions"`
}

type GeneratedChangesByStatusIndex struct {
	SchemaVersion int                         `json:"schema_version" yaml:"schema_version"`
	Statuses      GeneratedChangeStatusGroups `json:"statuses" yaml:"statuses"`
}

type GeneratedChangeStatusGroups struct {
	Proposed    []GeneratedChangeStatusEntry `json:"proposed" yaml:"proposed"`
	Planned     []GeneratedChangeStatusEntry `json:"planned" yaml:"planned"`
	Implemented []GeneratedChangeStatusEntry `json:"implemented" yaml:"implemented"`
	Verified    []GeneratedChangeStatusEntry `json:"verified" yaml:"verified"`
	Closed      []GeneratedChangeStatusEntry `json:"closed" yaml:"closed"`
	Superseded  []GeneratedChangeStatusEntry `json:"superseded" yaml:"superseded"`
}

type GeneratedChangeStatusEntry struct {
	ID    string `json:"id" yaml:"id"`
	Title string `json:"title" yaml:"title"`
	Type  string `json:"type" yaml:"type"`
	Size  string `json:"size,omitempty" yaml:"size,omitempty"`
	Path  string `json:"path" yaml:"path"`
}

type GeneratedBundlesIndex struct {
	SchemaVersion int                    `json:"schema_version" yaml:"schema_version"`
	Bundles       []GeneratedBundleEntry `json:"bundles" yaml:"bundles"`
}

type GeneratedBundleEntry struct {
	ID                 string                          `json:"id" yaml:"id"`
	Path               string                          `json:"path" yaml:"path"`
	Extends            []string                        `json:"extends" yaml:"extends"`
	ResolvedParents    []string                        `json:"resolved_parents" yaml:"resolved_parents"`
	ReferencedPatterns GeneratedBundlePatternAspectSet `json:"referenced_patterns" yaml:"referenced_patterns"`
}

type GeneratedBundlePatternAspectSet struct {
	Project   GeneratedBundleAspectPatterns `json:"project" yaml:"project"`
	Standards GeneratedBundleAspectPatterns `json:"standards" yaml:"standards"`
	Specs     GeneratedBundleAspectPatterns `json:"specs" yaml:"specs"`
	Decisions GeneratedBundleAspectPatterns `json:"decisions" yaml:"decisions"`
}

type GeneratedBundleAspectPatterns struct {
	Includes []GeneratedBundlePattern `json:"includes" yaml:"includes"`
	Excludes []GeneratedBundlePattern `json:"excludes" yaml:"excludes"`
}

type GeneratedBundlePattern struct {
	Pattern string            `json:"pattern" yaml:"pattern"`
	Kind    BundlePatternKind `json:"kind" yaml:"kind"`
}

func (p *ProjectIndex) BuildGeneratedManifest() (*GeneratedManifest, error) {
	if p == nil {
		return nil, fmt.Errorf("project index is required")
	}
	manifest := &GeneratedManifest{
		SchemaVersion: generatedManifestSchemaVersion,
		Indexes: GeneratedManifestIndexes{
			ChangesByStatus: generatedChangesIndexRelativePath,
			Bundles:         generatedBundlesIndexRelativePath,
		},
		Standards: SortedKeys(p.Standards),
		Bundles:   sortedBundleIDs(p),
		Changes:   SortedKeys(p.Changes),
		Specs:     SortedKeys(p.Specs),
		Decisions: SortedKeys(p.Decisions),
	}
	manifest.Counts = GeneratedManifestCounts{
		Standards: len(manifest.Standards),
		Bundles:   len(manifest.Bundles),
		Changes:   len(manifest.Changes),
		Specs:     len(manifest.Specs),
		Decisions: len(manifest.Decisions),
	}
	return manifest, nil
}

func (p *ProjectIndex) BuildGeneratedChangesByStatusIndex() (*GeneratedChangesByStatusIndex, error) {
	if p == nil {
		return nil, fmt.Errorf("project index is required")
	}
	index := &GeneratedChangesByStatusIndex{
		SchemaVersion: generatedChangeIndexSchemaVersion,
		Statuses: GeneratedChangeStatusGroups{
			Proposed:    []GeneratedChangeStatusEntry{},
			Planned:     []GeneratedChangeStatusEntry{},
			Implemented: []GeneratedChangeStatusEntry{},
			Verified:    []GeneratedChangeStatusEntry{},
			Closed:      []GeneratedChangeStatusEntry{},
			Superseded:  []GeneratedChangeStatusEntry{},
		},
	}
	for _, changeID := range SortedKeys(p.Changes) {
		record := p.Changes[changeID]
		if record == nil {
			continue
		}
		statusPath, err := generatedRelativeArtifactPath(p.ContentRoot, record.StatusPath)
		if err != nil {
			return nil, fmt.Errorf("build generated changes index: %w", err)
		}
		entry := GeneratedChangeStatusEntry{
			ID:    record.ID,
			Title: record.Title,
			Type:  record.Type,
			Size:  record.Size,
			Path:  statusPath,
		}
		switch record.Status {
		case StatusProposed:
			index.Statuses.Proposed = append(index.Statuses.Proposed, entry)
		case StatusPlanned:
			index.Statuses.Planned = append(index.Statuses.Planned, entry)
		case StatusImplemented:
			index.Statuses.Implemented = append(index.Statuses.Implemented, entry)
		case StatusVerified:
			index.Statuses.Verified = append(index.Statuses.Verified, entry)
		case StatusClosed:
			index.Statuses.Closed = append(index.Statuses.Closed, entry)
		case StatusSuperseded:
			index.Statuses.Superseded = append(index.Statuses.Superseded, entry)
		default:
			return nil, fmt.Errorf("build generated changes index: change %q has unsupported lifecycle status %q", record.ID, record.Status)
		}
	}
	return index, nil
}

func (p *ProjectIndex) BuildGeneratedBundlesIndex() (*GeneratedBundlesIndex, error) {
	if p == nil {
		return nil, fmt.Errorf("project index is required")
	}
	index := &GeneratedBundlesIndex{SchemaVersion: generatedBundleIndexSchemaVersion, Bundles: []GeneratedBundleEntry{}}
	if p.Bundles == nil {
		return index, nil
	}
	for _, bundleID := range SortedKeys(p.Bundles.bundles) {
		bundle := p.Bundles.bundles[bundleID]
		if bundle == nil {
			continue
		}
		resolution, err := p.Bundles.Resolve(bundleID)
		if err != nil {
			return nil, err
		}
		bundlePath, err := generatedRelativeArtifactPath(p.ContentRoot, bundle.Path)
		if err != nil {
			return nil, fmt.Errorf("build generated bundles index: %w", err)
		}
		entry := GeneratedBundleEntry{
			ID:              bundle.ID,
			Path:            bundlePath,
			Extends:         append([]string(nil), bundle.Extends...),
			ResolvedParents: resolvedBundleParents(resolution, bundleID),
			ReferencedPatterns: GeneratedBundlePatternAspectSet{
				Project:   generatedBundleAspectPatterns(bundle, BundleAspectProject),
				Standards: generatedBundleAspectPatterns(bundle, BundleAspectStandards),
				Specs:     generatedBundleAspectPatterns(bundle, BundleAspectSpecs),
				Decisions: generatedBundleAspectPatterns(bundle, BundleAspectDecisions),
			},
		}
		index.Bundles = append(index.Bundles, entry)
	}
	return index, nil
}

func (p *ProjectIndex) WriteGeneratedIndexes() error {
	if p == nil {
		return fmt.Errorf("project index is required")
	}
	if strings.TrimSpace(p.ContentRoot) == "" {
		return fmt.Errorf("project index content root is required")
	}
	manifest, err := p.BuildGeneratedManifest()
	if err != nil {
		return err
	}
	changesByStatus, err := p.BuildGeneratedChangesByStatusIndex()
	if err != nil {
		return err
	}
	bundles, err := p.BuildGeneratedBundlesIndex()
	if err != nil {
		return err
	}
	indexDir := filepath.Join(p.ContentRoot, generatedIndexesDirectoryRelative)
	if err := os.MkdirAll(indexDir, 0o755); err != nil {
		return err
	}
	if err := writeGeneratedYAML(filepath.Join(p.ContentRoot, filepath.FromSlash(generatedManifestRelativePath)), manifest); err != nil {
		return err
	}
	if err := writeGeneratedYAML(filepath.Join(p.ContentRoot, filepath.FromSlash(generatedChangesIndexRelativePath)), changesByStatus); err != nil {
		return err
	}
	if err := writeGeneratedYAML(filepath.Join(p.ContentRoot, filepath.FromSlash(generatedBundlesIndexRelativePath)), bundles); err != nil {
		return err
	}
	return nil
}

func writeGeneratedYAML(path string, doc any) error {
	data, err := renderGeneratedYAML(doc)
	if err != nil {
		return err
	}
	return writeFileAtomically(path, data, 0o644)
}

func renderGeneratedYAML(doc any) ([]byte, error) {
	var buf bytes.Buffer
	if err := encodeYAMLDocument(&buf, doc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sortedBundleIDs(index *ProjectIndex) []string {
	if index == nil || index.Bundles == nil {
		return []string{}
	}
	return SortedKeys(index.Bundles.bundles)
}

func resolvedBundleParents(resolution *BundleResolution, bundleID string) []string {
	if resolution == nil {
		return []string{}
	}
	parents := make([]string, 0, len(resolution.Linearization))
	for _, id := range resolution.Linearization {
		if id == bundleID {
			continue
		}
		parents = append(parents, id)
	}
	return parents
}

func generatedBundleAspectPatterns(bundle *bundleDefinition, aspect BundleAspect) GeneratedBundleAspectPatterns {
	result := GeneratedBundleAspectPatterns{Includes: []GeneratedBundlePattern{}, Excludes: []GeneratedBundlePattern{}}
	if bundle == nil {
		return result
	}
	for _, rule := range bundle.Includes[aspect] {
		result.Includes = append(result.Includes, GeneratedBundlePattern{Pattern: rule.Pattern, Kind: rule.PatternKind})
	}
	for _, rule := range bundle.Excludes[aspect] {
		result.Excludes = append(result.Excludes, GeneratedBundlePattern{Pattern: rule.Pattern, Kind: rule.PatternKind})
	}
	return result
}

func generatedRelativeArtifactPath(root, targetPath string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("generated artifacts require a content root")
	}
	if strings.TrimSpace(targetPath) == "" {
		return "", fmt.Errorf("generated artifacts require a target path")
	}
	rel, err := filepath.Rel(root, targetPath)
	if err != nil {
		return "", fmt.Errorf("resolve relative path for %q: %w", targetPath, err)
	}
	rel = filepath.ToSlash(rel)
	if rel == "" || rel == "." {
		return "", fmt.Errorf("resolve relative path for %q: empty relative output", targetPath)
	}
	if strings.HasPrefix(rel, "../") || rel == ".." || filepath.IsAbs(rel) || strings.HasPrefix(rel, "/") {
		return "", fmt.Errorf("path %q escapes RuneContext content root", targetPath)
	}
	return rel, nil
}

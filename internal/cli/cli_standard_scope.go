package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func discoverCandidateStandards(index *contracts.ProjectIndex, scopePaths []string, focus string) []string {
	if index == nil {
		return nil
	}
	paths := contracts.SortedKeys(index.Standards)
	focus = strings.ToLower(strings.TrimSpace(focus))
	candidates := make([]string, 0, len(paths))
	for _, path := range paths {
		record := index.Standards[path]
		if record == nil || record.Status != contracts.StandardStatusActive {
			continue
		}
		if !pathMatchesAnyScope(path, scopePaths) {
			continue
		}
		if !standardMatchesFocus(path, record, focus) {
			continue
		}
		candidates = append(candidates, path)
	}
	return candidates
}

func normalizedScopePaths(items []string) ([]string, error) {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		normalized, err := normalizeScopePath(item)
		if err != nil {
			return nil, err
		}
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out, nil
}

func normalizeScopePath(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" {
		return "", nil
	}
	if hasParentTraversalSegment(trimmed) {
		return "", fmt.Errorf("invalid --scope-path %q: path traversal outside standards is not allowed", value)
	}
	normalized := canonicalScopePath(trimmed)
	if normalized == "" {
		return "", nil
	}
	normalized = ensureStandardsPrefix(normalized)
	if !isInStandardsNamespace(normalized) {
		return "", fmt.Errorf("invalid --scope-path %q: normalized path must stay under standards/", value)
	}
	return normalized, nil
}

func canonicalScopePath(value string) string {
	normalized := filepath.ToSlash(filepath.Clean(value))
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = strings.Trim(normalized, "/")
	if normalized == "" || normalized == "." || normalized == "runecontext" {
		return ""
	}
	return strings.TrimPrefix(normalized, "runecontext/")
}

func ensureStandardsPrefix(path string) string {
	if isInStandardsNamespace(path) {
		return path
	}
	return filepath.ToSlash(filepath.Join("standards", path))
}

func isInStandardsNamespace(path string) bool {
	return path == "standards" || strings.HasPrefix(path, "standards/")
}

func hasParentTraversalSegment(path string) bool {
	for _, segment := range strings.Split(filepath.ToSlash(path), "/") {
		if segment == ".." {
			return true
		}
	}
	return false
}

func pathMatchesAnyScope(path string, scopes []string) bool {
	if len(scopes) == 0 {
		return true
	}
	for _, scope := range scopes {
		if path == scope || strings.HasPrefix(path, scope+"/") {
			return true
		}
	}
	return false
}

func standardMatchesFocus(path string, record *contracts.StandardRecord, focus string) bool {
	if focus == "" {
		return true
	}
	if record == nil {
		return false
	}
	if strings.Contains(strings.ToLower(path), focus) {
		return true
	}
	if strings.Contains(strings.ToLower(record.ID), focus) {
		return true
	}
	return strings.Contains(strings.ToLower(record.Title), focus)
}

package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type assuranceBackfillResult struct {
	baselinePath   string
	historyPath    string
	adoptionCommit string
	commitCount    int
	importedAdded  bool
}

func executeAssuranceBackfill(root string) (assuranceBackfillResult, error) {
	context, err := newAssuranceEnableContext(root)
	if err != nil {
		return assuranceBackfillResult{}, err
	}
	if fmt.Sprint(context.rootCfg["assurance_tier"]) != "verified" {
		return assuranceBackfillResult{}, fmt.Errorf("assurance_tier must be verified before running backfill")
	}
	baselineEnvelope, baselineMap, err := loadAssuranceBaseline(context.baselinePath)
	if err != nil {
		return assuranceBackfillResult{}, err
	}
	adoptionCommit, err := baselineAdoptionCommit(baselineEnvelope)
	if err != nil {
		return assuranceBackfillResult{}, err
	}
	historyPath, commitCount, err := buildAndWriteImportedHistory(root, adoptionCommit)
	if err != nil {
		return assuranceBackfillResult{}, err
	}
	baselineUpdated, err := appendImportedEvidenceAndWriteBaseline(root, context.baselinePath, baselineMap, historyPath)
	if err != nil {
		return assuranceBackfillResult{}, err
	}
	return assuranceBackfillResult{
		baselinePath:   context.baselinePath,
		historyPath:    historyPath,
		adoptionCommit: adoptionCommit,
		commitCount:    commitCount,
		importedAdded:  baselineUpdated,
	}, nil
}

func buildAndWriteImportedHistory(root, adoptionCommit string) (string, int, error) {
	history, err := buildImportedGitHistory(root, adoptionCommit)
	if err != nil {
		return "", 0, err
	}
	historyPath, err := writeImportedGitHistory(root, adoptionCommit, history)
	if err != nil {
		return "", 0, err
	}
	return historyPath, len(history), nil
}

func appendImportedEvidenceAndWriteBaseline(root, baselinePath string, baselineMap map[string]any, historyPath string) (bool, error) {
	relativeHistoryPath, err := filepath.Rel(root, historyPath)
	if err != nil {
		return false, fmt.Errorf("resolve relative history path: %w", err)
	}
	baselineUpdated, err := appendImportedEvidenceToBaseline(baselineMap, relativeHistoryPath)
	if err != nil {
		return false, err
	}
	if !baselineUpdated {
		return false, nil
	}
	updatedBaseline, err := yaml.Marshal(baselineMap)
	if err != nil {
		return false, fmt.Errorf("marshal updated baseline: %w", err)
	}
	if err := writeAtomicFile(baselinePath, updatedBaseline, 0o644); err != nil {
		return false, fmt.Errorf("write updated baseline: %w", err)
	}
	return true, nil
}

func loadAssuranceBaseline(path string) (contracts.AssuranceEnvelope, map[string]any, error) {
	baselineData, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return contracts.AssuranceEnvelope{}, nil, fmt.Errorf("assurance baseline not found: %s", path)
		}
		return contracts.AssuranceEnvelope{}, nil, fmt.Errorf("read assurance baseline: %w", err)
	}
	var envelope contracts.AssuranceEnvelope
	if err := yaml.Unmarshal(baselineData, &envelope); err != nil {
		return contracts.AssuranceEnvelope{}, nil, fmt.Errorf("parse assurance baseline: %w", err)
	}
	var baselineMap map[string]any
	if err := yaml.Unmarshal(baselineData, &baselineMap); err != nil {
		return contracts.AssuranceEnvelope{}, nil, fmt.Errorf("parse baseline object: %w", err)
	}
	return envelope, baselineMap, nil
}

func baselineAdoptionCommit(envelope contracts.AssuranceEnvelope) (string, error) {
	value, ok := envelope.Value.(map[string]any)
	if !ok {
		return "", fmt.Errorf("assurance baseline value must be an object")
	}
	adoptionCommit := readOptionalString(value, "adoption_commit")
	if adoptionCommit == "" {
		return "", fmt.Errorf("assurance baseline adoption_commit is required for backfill")
	}
	return adoptionCommit, nil
}

func appendImportedEvidenceToBaseline(baseline map[string]any, historyPath string) (bool, error) {
	valueRaw, ok := baseline["value"]
	if !ok || valueRaw == nil {
		valueRaw = map[string]any{}
	}
	value, ok := valueRaw.(map[string]any)
	if !ok {
		return false, fmt.Errorf("assurance baseline value must be an object")
	}
	evidenceRaw, ok := value["imported_evidence"]
	if !ok || evidenceRaw == nil {
		evidenceRaw = make([]any, 0)
	}
	evidence, ok := evidenceRaw.([]any)
	if !ok {
		return false, fmt.Errorf("assurance baseline imported_evidence must be a list")
	}
	if importedEvidenceExists(evidence, historyPath) {
		return false, nil
	}
	evidence = append(evidence, map[string]any{
		"provenance": "imported_git_history",
		"path":       filepath.ToSlash(historyPath),
	})
	value["imported_evidence"] = evidence
	baseline["value"] = value
	return true, nil
}

func importedEvidenceExists(evidence []any, historyPath string) bool {
	for _, raw := range evidence {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if readOptionalString(entry, "provenance") != "imported_git_history" {
			continue
		}
		if filepath.ToSlash(readOptionalString(entry, "path")) == filepath.ToSlash(historyPath) {
			return true
		}
	}
	return false
}

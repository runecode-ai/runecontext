# Contracts, Resolution, And Change Workflow

This package provides the shared executable core for RuneContext's implemented
alpha.1, alpha.2, and alpha.3 semantics.

## Current Coverage

- JSON Schema validation for machine-readable YAML contracts
- restricted YAML profile checks for duplicate keys and anchors/aliases
- strict markdown parsing for `proposal.md` and `standards.md`
- strict YAML-frontmatter validation for `specs/*.md` and `decisions/*.md`
- project-level traceability checks across changes, bundles, specs, and decisions
- content-root-aware project validation that follows `runecontext.yaml` source settings
- embedded, git, and local-path source resolution with structured metadata,
  signed-tag verification support, and monorepo discovery
- deterministic bundle loading and evaluation with inheritance, precedence,
  diagnostics, and path-boundary enforcement
- standards validation, migration metadata checks, and canonical path-based
  standards reference enforcement
- change ID allocation, lifecycle validation, shaping/rendering helpers,
  status summaries, and fail-closed change mutation workflows

## Intentional Scope

- This package owns the canonical file-model, validation, resolution, and
  change-workflow semantics implemented so far.
- Thin CLI wrappers live in `internal/cli/`; adapters, context-pack generation,
  assurance artifact generation, and broader admin flows remain future work.
- Later alphas should continue building on this package rather than re-encoding
  contract rules ad hoc.

# Schemas

Hand-authored JSON Schema files for machine-readable RuneContext artifacts live here.

## Current v0.1.0-alpha.3 Schema Inventory

- `runecontext.schema.json`
  - root `runecontext.yaml` contract
- `bundle.schema.json`
  - `bundles/*.yaml` contract
- `change-status.schema.json`
  - `changes/*/status.yaml` contract
- `context-pack.schema.json`
  - generated context-pack contract
- `spec.schema.json`
  - YAML frontmatter contract for `specs/*.md`
- `decision.schema.json`
  - YAML frontmatter contract for `decisions/*.md`
- `standard.schema.json`
  - YAML frontmatter contract for `standards/**/*.md`

## Notes

- All v1 schemas target JSON Schema Draft 2020-12.
- Machine-readable contracts stay closed by default and fail closed on unknown `schema_version` values.
- `specs/*.md`, `decisions/*.md`, and `standards/**/*.md` use YAML frontmatter as their strict metadata layer; the markdown body remains hand-authored.
- The executable validation, resolution, and change-workflow core in
  `internal/contracts/` compiles and exercises these schemas against repository
  fixtures.
- The schema inventory now covers the authored alpha.1-alpha.3 contract
  surfaces; `context-pack.schema.json` is defined ahead of alpha.4 generation
  work so deterministic pack outputs can land without reopening the v1 shape.

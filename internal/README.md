# Internal

This directory holds shared Go packages for RuneContext implementation code.

## Current Packages

- `internal/cli/`
  - thin alpha.3 CLI entrypoints, parsing helpers, and stable line-oriented
    output for `validate`, `status`, and change commands
- `internal/contracts/`
  - shared validation, source-resolution, bundle, standards, and change-
    workflow implementation that the CLI and tests build on

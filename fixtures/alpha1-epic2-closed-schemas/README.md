# Alpha 1 Epic 2: Closed Schemas Test Fixtures

This directory contains test fixtures for schema validation, extensions handling, and closed-schema enforcement in v0.1.0-alpha.1 Epic 2.

## Fixture Organization

### Test Cases

#### 1. Closed Schema Validation

- **valid-runecontext.yaml**: Valid `runecontext.yaml` with no extensions; `allow_extensions: false` (default).
- **valid-bundle-embedded.yaml**: Valid `bundles/base.yaml` for embedded mode; closed schema.
- **valid-change-status.yaml**: Valid `changes/CHG-2026-001-a1b2-example/status.yaml`; closed schema.
- **valid-context-pack.yaml**: Valid context pack (fully closed, generated).

#### 2. Closed Schema Rejection

- **reject-unknown-field-runecontext.yaml**: Contains unknown top-level field `unknown_field`. Must fail validation.
- **reject-unknown-field-bundle.yaml**: Contains unknown top-level field `custom_metadata`. Must fail validation.
- **reject-unknown-field-status.yaml**: Contains unknown top-level field `internal_id`. Must fail validation.
- **reject-unknown-schema-version-bundle.yaml**: Contains `schema_version: 2` (unknown). Must fail closed.
- **reject-unknown-enum-type.yaml**: Contains unknown `type` value `x-internal` (not valid enum or x- prefix). Actually valid; see custom-type fixture instead.

#### 3. Extensions Without Opt-In (Must Reject)

- **reject-extensions-no-optin-bundle.yaml**: Contains `extensions` object but `runecontext.yaml` has `allow_extensions: false` (default). Must fail validation.
- **reject-extensions-no-optin-status.yaml**: Contains `extensions` object in status file; project has not opted in. Must fail validation.

#### 4. Extensions With Opt-In (Should Pass With Warning)

- **optin-with-extensions-runecontext.yaml**: `runecontext.yaml` with `allow_extensions: true`.
- **optin-with-extensions-bundle.yaml**: `bundles/custom.yaml` with `extensions: {io.runecode.custom_data: "example"}`. Requires `allow_extensions: true` in project config.
- **optin-with-extensions-status.yaml**: `changes/*/status.yaml` with `extensions: {dev.acme.internal_ref: "ABC-123"}`. Requires `allow_extensions: true`.

#### 5. Valid Extensions (Namespaced Keys)

- **valid-extensions-namespaced-bundle.yaml**: Extensions with properly namespaced keys:
  - `io.runecode.metadata: {...}`
  - `dev.acme.custom: "value"`
  - `org.example.config: {...}`

#### 6. Invalid Extensions (Typos, Violations)

- **reject-extensions-bad-key-bundle.yaml**: Extensions key `badKey` (not lowercase/namespaced). Must fail validation.
- **reject-extensions-bad-key-status.yaml**: Extensions key `X-CUSTOM` (uppercase, not namespaced). Must fail validation.
- **reject-extensions-leading-hyphen.yaml**: Extensions key `-invalid.key`. Must fail validation.
- **reject-extensions-trailing-dot.yaml**: Extensions key `io.runecode.`. Must fail validation.

#### 7. Custom Type Values (x- Prefix)

- **valid-custom-type-status.yaml**: `type: x-migration` (valid custom type in change status). Allowed without extensions opt-in.
- **valid-custom-type-epic.yaml**: `type: x-epic` (another valid custom type).

#### 8. YAML Profile Compliance

- **reject-anchors-and-aliases.yaml**: Contains YAML anchors (`&ref`) and aliases (`*ref`). Must fail validation.
- **reject-duplicate-keys.yaml**: Contains duplicate keys in an object. Must fail validation.
- **reject-implicit-coercion.yaml**: Contains `yes`/`no` without explicit `true`/`false`. Must fail validation.
- **reject-custom-tags.yaml**: Contains YAML custom tags (`!!str`, `!!timestamp`). Must fail validation.
- **reject-multiline-strings.yaml**: Contains multiline strings with `|` or `>` syntax. Must fail validation.
- **valid-utf8-only.yaml**: Valid file; all non-ASCII characters are UTF-8 encoded (no other encodings).

#### 9. Restricted Source Verification Enum

- **valid-source-verification-pinned.yaml**: Context pack with `source_verification: pinned_commit`. Valid.
- **valid-source-verification-signed-tag.yaml**: Context pack with `source_verification: verified_signed_tag` and `verified_signer_identity` field. Valid.
- **valid-source-verification-mutable.yaml**: Context pack with `source_verification: unverified_mutable_ref`. Valid.
- **valid-source-verification-local.yaml**: Context pack with `source_verification: unverified_local_source`. Valid.
- **reject-source-verification-unknown.yaml**: Context pack with unknown `source_verification` value. Must fail validation.

#### 10. Context Pack (Fully Closed)

- **valid-context-pack-full.yaml**: Complete valid context pack with all required fields.
- **reject-context-pack-with-extensions.yaml**: Context pack with `extensions` field (not allowed in v1). Must fail validation.
- **reject-context-pack-unknown-field.yaml**: Context pack with unknown field `metadata`. Must fail validation.

#### 11. Hashing and Canonicalization

- **valid-jcs-hash-computation.json**: Example JSON before and after JCS canonicalization with expected SHA256 hash.
  - Shows key sorting, number normalization, whitespace removal.
  - Includes expected hash output for parity testing.

## Test Expectations

### Validation Pass Cases
- All `valid-*.yaml` files must pass schema validation.
- All namespaced extension keys must be accepted when `allow_extensions: true`.

### Validation Fail Cases
- All `reject-*.yaml` files must fail schema validation with clear error messages.
- Unknown `schema_version` values must fail closed.
- Extensions without opt-in must fail validation.
- YAML profile violations (anchors, aliases, etc.) must fail validation.

### Warning Cases
- When extensions are present and `allow_extensions: true`, implementations must issue a visible warning.

## Usage in Implementation

1. **Unit Tests**: Use these fixtures to validate schema implementations across Go and TypeScript.
2. **Parity Tests**: Ensure local and remote implementations reject/accept the same fixtures.
3. **Integration Tests**: Verify that fixture-based validation is reliable in end-to-end workflows.
4. **CI/CD**: Include fixtures in automated test runs to catch regressions.

## Fixture Format

All fixtures use YAML or JSON as appropriate. JSON fixtures (for hashing examples) are provided for clarity but are not meant to be validated against schemas; they serve as documentation of expected canonical forms.

Files follow the restricted YAML profile defined in `schemas/MACHINE-READABLE-PROFILE.md`:
- No anchors or aliases
- No duplicate keys
- No custom tags
- UTF-8 only
- 2-space indentation
- Unix line endings

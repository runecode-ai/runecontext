# RuneContext ↔ RuneCode Compatibility Matrix

This repository and the RuneCode companion track follow the compatibility guidance established in
[docs/project_idea.md](docs/project_idea.md#runecontext-compatibility) and the alpha.8 value stream described in
[docs/implementation-plan/milestone-breakdown.md](docs/implementation-plan/milestone-breakdown.md#recommended-branch-cut-3-release-artifacts-compatibility-matrix-and-release-verification).
RuneCode uses the `runecontext_version` string from each project's root `runecontext.yaml` as the canonical
compatibility marker. Every release run should compare that value against its supported range before attempting
deeper validation or runtime wiring.

## Compatibility overview

| RuneCode release | Acceptable `runecontext_version` range | Adapter-pack compatibility | Notes |
| --- | --- | --- | --- |
| `v0.1.0-alpha.*` | `0.1.0-alpha.5` – `0.1.0-alpha.8` | `adapter-generic`, `adapter-codex`, `adapter-claude-code`, `adapter-opencode` (matching release tag) | RuneCode alpha builds target the matching RuneContext alpha releases. The `runecontext_version` field mirrors the release tag so RuneCode can gate upgrades and mixed-tree detection. |
| `v0.1.0` | `0.1.0` (future GA) | Same adapter packs plus any follow-up host-specific adapters | Compatibility freezes once RuneContext ships v0.1.0 GA; RuneCode will reject older `runecontext_version` values and expect clients to run `runectx upgrade` before further interaction. |

This table remains intentionally simple: every RuneCode release ships with a single supported `runecontext_version` range and the adapter packs emitted by the same GitHub Release. If a project reports a `runecontext_version` outside these ranges, `runectx validate`, `runectx doctor`, and RuneCode integration flows should fail fast and direct the user to upgrade.

## Release asset compatibility

- The `schema-bundle.tar.gz` file contains the authoritative `schemas/` directory for the release. RuneCode and other automation can use it to validate bundle/pack/input schemas without cloning the entire repository.
- The adapter packs (`adapter-generic.tar.gz`, `adapter-codex.tar.gz`, `adapter-claude-code.tar.gz`, and `adapter-opencode.tar.gz`) contain the host-specific prompt/skill workflows for each supported tool. They are versioned alongside the release tag and referenced in `runecontext_<tag>_release-manifest.json` under the `adapter_pack` kind so automation can discover them deterministically.
- Refer to `docs/release-process.md` and `docs/install-verify.md` for the signature, certificate, and attestation steps that tie the schema bundle, adapter packs, repo bundles, and optional `runectx` binaries to the same signed release.

RuneCode should mirror these release expectations so its own build/publish flow consumes adapter packs from the verified release graph rather than relying on ad hoc downloads. The compatibility matrix above is the authoritative source for RuneCode version gates and upgrade planning guidance used in the alpha.8 release/install story.

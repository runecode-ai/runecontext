# Codex Adapter Setup

This adapter is synced with:

```sh
runectx adapter sync codex --path <project-root>
```

The sync writes only to:

- `.runecontext/adapters/codex/managed/`
- `.runecontext/adapters/codex/sync-manifest.yaml`

No implicit network fetches occur during sync.

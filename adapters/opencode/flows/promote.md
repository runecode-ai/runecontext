# OpenCode Flow: promote

Use this conversational flow to advance promotion state explicitly.

## Inputs

- change ID
- action (`--accept` or `--complete`)
- optional targets (repeatable)
- optional project path

## Command Mapping

```sh
runectx promote <CHANGE_ID> [--accept|--complete] [--target <TYPE:PATH>] [--path <project-root>]
```

## Review Checkpoint

- Validate that targets come from discover output or explicit user input.
- Never mutate promotion state without explicit command execution.

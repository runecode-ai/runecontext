# OpenCode Flow: standard discover

Use this conversational flow to gather advisory standards candidates and optional
promotion handoff metadata.

## Inputs

- optional project path
- optional change ID
- optional scope paths (repeatable)
- optional focus text
- optional handoff confirmation and target

## Command Mapping

```sh
runectx standard discover [--path <project-root>] [--change <CHANGE_ID>] [--scope-path <path>] [--focus "<text>"] [--confirm-handoff] [--target <TYPE:PATH>]
```

## Candidate Data Rule

- Candidate promotion targets come from `candidate_promotion_target_*` output.
- Reuse emitted target values directly; do not invent hidden targets.

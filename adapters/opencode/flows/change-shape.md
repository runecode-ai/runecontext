# OpenCode Flow: change shape

Use this conversational flow to shape an existing change with explicit
`runectx` flags.

## Inputs

- change ID
- optional design summary
- optional verification summary
- optional tasks (repeatable)
- optional references (repeatable)
- optional project path

## Command Mapping

```sh
runectx change shape <CHANGE_ID> [--design "<text>"] [--verification "<text>"] [--task "<text>"] [--reference "<text>"] [--path <project-root>]
```

## Review Checkpoint

- Keep tasks/references user-visible in the conversation.
- Confirm the target `CHANGE_ID` before execution.

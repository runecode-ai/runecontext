# OpenCode Flow: change new

Use this conversational flow to create a new change while keeping all mutation
semantics in `runectx`.

## Inputs

- title
- type (`project|feature|bug|standard|chore`)
- optional size (`small|medium|large`)
- optional shape (`minimum|full`)
- optional bundle IDs
- optional description
- optional project path

## Command Mapping

```sh
runectx change new --title "<title>" --type <type> [--size <size>] [--shape <minimum|full>] [--bundle <bundle-id>] [--description "<text>"] [--path <project-root>]
```

## Review Checkpoint

- Show the full command before execution.
- Capture resulting change ID from command output.

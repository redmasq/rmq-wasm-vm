# TODO

This file is for development notes and deferred implementation details that are
better kept out of the source code.

Jira remains the source of truth for issue scope and status. When a note here
maps to a Jira item, prefer referencing the ticket key.

This snapshot was refreshed from Jira on 2026-03-23. Issues already marked
done were omitted.

## Open Jira Snapshot

### RWV-5 - Numeric Arithmatic Instructions for i32 and i64

- Type: Epic
- Status: In Progress
- Child tasks:
- RWV-21 - Implement `i32.rem_s` opcode
- RWV-22 - Implement `i32.rem_u` opcode
- RWV-23 - Implement `i64.rem_s` opcode
- RWV-24 - Implement `i64.rem_u` opcode

### RWV-15 - Retrofit for Golang conventions and minor improvements

- Type: Epic
- Status: To Do
- Child tasks:
- RWV-16 - Adjust types for `VMState` and `VMConfig` elements so they use specific typedefs where appropriate (`ON HOLD`)

### RWV-19 - Improve VM Infrastructure

- Type: Epic
- Status: To Do
- Child tasks:
- RWV-18 - Replace `VMConfig.FlatMemory` with a Memory Compositor (`ON HOLD`)
- RWV-20 - Improve pull request validation performance in GitHub Actions

### RWV-25 - Implement integer bitwise opcodes for i32 and i64

- Type: Epic
- Status: To Do
- Child tasks:
- RWV-28 - Implement `i32.and` opcode
- RWV-29 - Implement `i32.or` opcode
- RWV-30 - Implement `i32.xor` opcode
- RWV-31 - Implement `i64.and` opcode
- RWV-32 - Implement `i64.or` opcode
- RWV-33 - Implement `i64.xor` opcode

### RWV-26 - Implement integer unary numeric opcodes for i32 and i64

- Type: Epic
- Status: To Do
- Child tasks:
- RWV-34 - Implement `i32.clz` opcode
- RWV-35 - Implement `i32.ctz` opcode
- RWV-36 - Implement `i32.popcnt` opcode
- RWV-37 - Implement `i64.clz` opcode
- RWV-38 - Implement `i64.ctz` opcode
- RWV-39 - Implement `i64.popcnt` opcode

### RWV-27 - Implement integer shift and rotate opcodes for i32 and i64

- Type: Epic
- Status: To Do
- Child tasks:
- RWV-40 - Implement `i32.shl` opcode
- RWV-41 - Implement `i32.shr_s` opcode
- RWV-42 - Implement `i32.shr_u` opcode
- RWV-43 - Implement `i32.rotl` opcode
- RWV-44 - Implement `i32.rotr` opcode
- RWV-45 - Implement `i64.shl` opcode
- RWV-46 - Implement `i64.shr_s` opcode
- RWV-47 - Implement `i64.shr_u` opcode
- RWV-48 - Implement `i64.rotr` opcode
- RWV-49 - Implement `i64.rotl` opcode

### RWV-50 - Implement control flow opcodes, call-frame support, and entrypoint behavior

- Type: Epic
- Status: To Do
- Child tasks:
- RWV-51 - Add `block` support and retrofit `end` for structured block control
- RWV-52 - Add `CallStack` support for control flow and function execution
- RWV-53 - Add function-body execution support and retrofit `end` for function boundaries
- RWV-54 - Support Wasm start behavior and optionally prefer `_start` while retaining offset startup
- RWV-55 - Implement remaining control flow opcodes and retrofit `end` as needed

## Deferred / Backlog Notes

### Memory compositor follow-up items

- FlatSharedMemory supporting the shared flag according to standard.
- RingMemory compositor using VM context.
- Per-4 KiB-shard write ownership limits across threads.
- Default W^X behavior unless overridden by host configuration.
- Shard-level guarantee modes: None, Atomic, BeforeEndInstruction, and
  DirtyWithCommitSync.
- MMU-like assignment per execution context via ring 0 ring templates.

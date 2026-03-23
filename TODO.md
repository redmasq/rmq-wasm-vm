# TODO

This file is for development notes and deferred implementation details that are
better kept out of the source code.

Jira remains the source of truth for issue scope and status. When a note here
maps to a Jira item, prefer referencing the ticket key.

This snapshot was populated from a Jira RSS export provided on 2026-03-23.
Issues already marked done were omitted. No canceled issues appeared in that
export.

## Open Jira Snapshot

### RWV-19 - Improve VM Infrastructure

- Type: Epic
- Status: To Do
- Priority: Medium
- Parent: None
- Labels: None
- Created: 2025-08-04
- Updated: 2025-08-04
- Description: None

### RWV-18 - Replace VMConfig.FlatMemory with a Memory Compositor

- Type: Task
- Status: ON HOLD
- Priority: Medium
- Parent: RWV-19
- Labels: config, optimize, platform
- Created: 2025-08-04
- Updated: 2026-03-23
- Description:
- This will include a `FlatMemory` compositor which returns a 4 KiB shard of
  memory but otherwise has the same behavior as direct access.
- It must accept the logical address, the purpose (`read`, `write`,
  `execute`), and ring execution context. `FlatMemory` only uses the logical
  address, which is taken as the same as the backing memory.
- Not included in this ticket; add to backlog:
- `FlatSharedMemory`, supporting the shared flag according to standard.
- `RingMemory`, using VM context.
- Only up to one thread has write access to any 4 KiB shard.
- W^X by default unless overridden in host configuration.
- Guarantee modes by shard: `None`, `Atomic`, `BeforeEndInstruction`, and
  `DirtyWithCommitSync`.
- MMU-like assignment per execution context via ring templates defined in ring
  0.

### RWV-17 - Retrofit Trap handling to use advanced error types with description and enumeration

- Type: Task
- Status: In Progress
- Priority: Medium
- Parent: RWV-15
- Labels: onhold, optimize
- Created: 2025-07-19
- Updated: 2025-11-28
- Description: None

### RWV-16 - Adjust types for VMState and VMConfig elements so that they use specific typedefs where appropriate

- Type: Task
- Status: ON HOLD
- Priority: Medium
- Parent: RWV-15
- Labels: optimize
- Created: 2025-07-19
- Updated: 2025-08-05
- Description: None

### RWV-15 - Retrofit for Golang conventions and minor improvements

- Type: Epic
- Status: To Do
- Priority: Medium
- Parent: None
- Labels: None
- Created: 2025-07-19
- Updated: 2025-07-19
- Description: None

### RWV-5 - Numeric Arithmatic Instructions for i32 and i64

- Type: Epic
- Status: In Progress
- Priority: Medium
- Parent: None
- Labels: opcodes
- Created: 2025-07-16
- Updated: 2025-07-16
- Description: None

### RWV-3 - Handler and Unit Tests for div_s.i64

- Type: Task
- Status: To Do
- Priority: Medium
- Parent: RWV-5
- Labels: opcodes, unittests
- Created: 2025-07-16
- Updated: 2025-07-20
- Description: None

## Deferred / Backlog Notes

### Memory compositor follow-up items

- FlatSharedMemory supporting the shared flag according to standard.
- RingMemory compositor using VM context.
- Per-4 KiB-shard write ownership limits across threads.
- Default W^X behavior unless overridden by host configuration.
- Shard-level guarantee modes: None, Atomic, BeforeEndInstruction, and
  DirtyWithCommitSync.
- MMU-like assignment per execution context via ring 0 ring templates.

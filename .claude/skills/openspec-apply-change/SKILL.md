---
name: openspec-apply-change
description: Drive implementation of a change's task list to completion, checking off work as it lands and pausing on ambiguity or blockers. Trigger when the user wants to start, resume, or grind through implementation work on an existing change.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.0.2"
---

Turn a change's task list into working code, one task at a time, tracking progress as you go.

## Scope

Optional change name as input. Missing name is fine — infer it from the conversation, or fall back to prompting. Never invent a change name out of thin air when more than one candidate exists.

## Resolving which change to work on

- Explicit name given → use it directly.
- No name, but the conversation clearly points at one change → use that.
- No name, exactly one active change exists → auto-select it.
- Still ambiguous → run `openspec list --json`, then use **AskUserQuestion** to let the user pick from the list.

Whichever path is taken, state it out loud: "Using change: <name>" plus the override syntax (`/opsx:apply <other>`).

## Reading the workflow shape

Query the schema before touching anything:

```bash
openspec status --change "<name>" --json
```

Pull out `schemaName` (which workflow this change follows — e.g. "spec-driven") and locate which artifact actually holds the task checklist (usually `tasks`, but confirm rather than assume for non-default schemas).

## Pulling implementation guidance

```bash
openspec instructions apply --change "<name>" --json
```

The payload carries:
- paths to context files (proposal/specs/design/tasks for spec-driven; something else for other schemas)
- progress counters (total / complete / remaining)
- the task list with per-task status
- a dynamic instruction reflecting current state

Branch on the returned state:
- `"blocked"` → required artifacts are missing; point the user at openspec-continue-change.
- `"all_done"` → nothing left to implement; congratulate and point toward archiving.
- anything else → proceed into implementation.

## Loading context

Read every file named in `contextFiles`. Don't hardcode filenames — schemas vary in what they call their artifacts (spec-driven uses proposal/specs/design/tasks; others differ per the CLI output).

## Reporting where things stand

Before touching code, surface: which schema is active, "N/M tasks complete," a summary of what's left, and the CLI's current dynamic instruction.

## Working the task list

Cycle through pending tasks until the list is empty or something stops you:

- Announce which task is in flight.
- Make the minimal code change the task calls for — no scope creep.
- Flip its checkbox (`- [ ]` → `- [x]`) as soon as it's done.
- Move to the next one.

Stop and surface the situation instead of guessing when:
- the task's intent is genuinely unclear,
- the implementation surfaces a design problem worth revisiting artifacts for,
- something errors out or blocks progress,
- the user steps in mid-loop.

## Wrapping a session

Whether you finished the list or got interrupted, report: tasks closed this session, overall "N/M complete," and either a nudge toward archiving (all done) or an explanation of the pause plus a wait for direction.

## Sample transcripts

While working:
```
## Implementing: <change-name> (schema: <schema-name>)

Working on task 3/7: <task description>
[...implementation happening...]
✓ Task complete

Working on task 4/7: <task description>
[...implementation happening...]
✓ Task complete
```

Finished:
```
## Implementation Complete

**Change:** <change-name>
**Schema:** <schema-name>
**Progress:** 7/7 tasks complete ✓

### Completed This Session
- [x] Task 1
- [x] Task 2
...

All tasks complete! Ready to archive this change.
```

Interrupted:
```
## Implementation Paused

**Change:** <change-name>
**Schema:** <schema-name>
**Progress:** 4/7 tasks complete

### Issue Encountered
<description of the issue>

**Options:**
1. <option 1>
2. <option 2>
3. Other approach

What would you like to do?
```

## Operating rules

- Don't stop between tasks unless done or genuinely blocked.
- Load context files before writing any code.
- Ambiguous task → ask before guessing.
- Design issues surfaced mid-implementation → flag artifact updates, don't silently improvise.
- Every change should be scoped tightly to its task — resist drive-by refactors.
- Checkbox flips happen immediately, not batched at the end.
- Errors and unclear requirements are pause conditions, not things to paper over.
- Derive filenames from `contextFiles`; never assume fixed names.

## Fits into the fluid workflow model

This isn't phase-locked:
- Can run before every artifact exists (as long as tasks do), mid-implementation, or interleaved with other actions.
- If work in progress exposes a design gap, it's fine to suggest updating proposal/design/specs before continuing — the loop isn't rigid about ordering.

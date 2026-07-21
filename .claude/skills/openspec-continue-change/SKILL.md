---
name: openspec-continue-change
description: Advance an in-progress OpenSpec change by drafting whichever artifact comes next in its schema. Trigger when the user wants to pick a change back up or move it one artifact further along.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.0.2"
---

Draft the next unlocked artifact for a change, one artifact per invocation.

## Scope

Change name is optional. Infer it from conversation if possible; otherwise prompt from the active list. Never assume when more than one candidate exists.

## Picking the change

No name given → run `openspec list --json`, sort by recency, and hand the top 3-4 to **AskUserQuestion**. For each option surface: name, schema (default to "spec-driven" if unset), a rough status like "2/5 tasks" or "no tasks yet", and how recently it moved. Tag the most recently touched one "(Recommended)" — it's the likely target — but the choice is always the user's, never auto-selected.

## Reading current state

```bash
openspec status --change "<name>" --json
```

Pull `schemaName`, the `artifacts` array (each with a `status` of `done` / `ready` / `blocked`), and the `isComplete` flag.

## Branching on status

**Everything done (`isComplete: true`)** — congratulate, print final status with the schema name, and point toward implementing (`/opsx:apply`) or archiving. Stop here.

**Something is `ready`** — take the first ready artifact and fetch its build instructions:
```bash
openspec instructions <artifact-id> --change "<name>" --json
```
The payload gives you `context` and `rules` (constraints that shape your writing — never copied into the file), `template` (the skeleton to fill in), `instruction` (schema-specific guidance), `outputPath`, and `dependencies` (already-completed artifacts worth reading first).

Read the dependency files, fill in the template using `instruction` as guidance, respect `context`/`rules` without echoing them, and write to `outputPath`. Report what got created and what it unlocked. Stop — one artifact per call, no chaining.

**Nothing ready, nothing done** — this signals a broken schema graph; show the raw status and flag it rather than guessing a fix.

## Confirming progress

```bash
openspec status --change "<name>"
```

## What to report back

Artifact just created, active schema, progress as N/M, what's newly unlocked, and an invitation to keep going ("ask me to continue, or tell me what's next").

## Artifact content by schema

For the default `spec-driven` schema (proposal → specs → design → tasks):
- `proposal.md` — Why / What Changes / Capabilities / Impact. The Capabilities list matters: each entry needs a matching spec file later.
- `specs/<capability>/spec.md` — one file per capability named in the proposal (capability name, not change name).
- `design.md` — the technical decisions and approach.
- `tasks.md` — checkbox breakdown of implementation work.

Any other schema: follow whatever `instruction` says rather than assuming this shape.

## Operating notes

- One artifact per invocation, no exceptions.
- Read dependencies before writing anything.
- Don't skip ahead or reorder the schema's sequence.
- Genuinely unclear intent → ask before drafting.
- Confirm the file actually landed on disk before reporting progress.
- `context` and `rules` steer you; they are never file content — strip any `<context>`, `<rules>`, `<project_context>` blocks out of what you write.

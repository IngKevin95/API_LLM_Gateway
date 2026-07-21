---
name: openspec-ff-change
description: Generate every artifact a change needs to reach implementation-ready, back to back, without pausing between them. Trigger when the user wants to skip the step-by-step ritual and get straight to tasks.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.0.2"
---

Push a change from zero to implementation-ready in one pass — every required artifact, no stopping in between.

## Input

A change name (kebab-case) or a plain description of the work. If neither is clear, use **AskUserQuestion** open-ended: "What do you want to build or fix?" Derive a kebab-case name from the answer (e.g. "add user authentication" → `add-user-auth`). Don't proceed on a guess — get an actual answer first.

## Setting up the change

```bash
openspec new change "<name>"
```
Scaffolds `openspec/changes/<name>/`. If a change with that name already exists, suggest `/opsx:continue` on it instead of creating a duplicate.

## Mapping the build order

```bash
openspec status --change "<name>" --json
```
Extract `applyRequires` (the artifact IDs that must be `done` before implementation can start) and the full `artifacts` list with statuses and dependencies.

## Working through the queue

Track progress with **TodoWrite** as you go. Process artifacts in dependency order — anything with `status: "ready"` first.

For each ready artifact:
```bash
openspec instructions <artifact-id> --change "<name>" --json
```
This returns `context` and `rules` (writing constraints — never copied into the output), `template` (the file skeleton), `instruction` (what this artifact type needs), `outputPath`, and `dependencies` (finished artifacts worth reading first). Read those dependencies, fill the template per `instruction`, respect `context`/`rules` silently, write to `outputPath`, and report tersely: "created `<artifact-id>`".

Re-check status after each write:
```bash
openspec status --change "<name>" --json
```
Keep looping until every ID in `applyRequires` shows `status: "done"`.

If an artifact genuinely can't be drafted without more input, use **AskUserQuestion** to unstick it, then keep going — but default to a reasonable judgment call over stopping, since the whole point here is momentum.

## Final check

```bash
openspec status --change "<name>"
```

## What to report

Change name and location, each artifact created with a one-line description, confirmation that implementation-readiness is reached ("all required artifacts done"), and a nudge toward `/opsx:apply` or "ask me to implement" to start on tasks.

## Drafting guidance

Follow `instruction` per artifact type — the schema, not this skill, defines what belongs in each file. Always read dependency artifacts before drafting something that depends on them. `context` and `rules` are inputs that shape your writing, not text that belongs in the output — strip out any `<context>`, `<rules>`, or `<project_context>` blocks before saving.

## Operating notes

- Every artifact in `apply.requires` gets created — this isn't a partial run.
- Read dependencies before each new artifact, every time.
- Prefer a reasonable default over stalling on ambiguity; ask only when it's truly blocking.
- Existing change with the same name → redirect to continuing it.
- Confirm each file actually landed before moving to the next.

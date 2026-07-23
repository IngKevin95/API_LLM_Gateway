---
name: openspec-new-change
description: Scaffold a fresh OpenSpec change and stop right after showing the first artifact's template. Trigger when the user wants to start a new feature, fix, or modification step by step rather than all at once.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.0.2"
---

Scaffold a new change and hand back the first artifact's template — nothing gets written yet.

## Input

A kebab-case name or a description of the work. If neither is clear, ask open-ended via **AskUserQuestion**: "What do you want to build or fix?" Derive a kebab-case name from whatever comes back (e.g. "add user authentication" → `add-user-auth`). Don't move forward without a real answer.

## Choosing a schema

Default: omit `--schema` entirely and let the CLI pick its default workflow.

Only deviate when the user:
- names a specific schema explicitly → pass `--schema <name>`
- asks "what workflows are there" → run `openspec schemas --json` and let them choose from the list

Otherwise, don't second-guess the default.

## Scaffolding

```bash
openspec new change "<name>"
```
Add `--schema <name>` only if one was explicitly requested. This creates `openspec/changes/<name>/` seeded with the chosen schema.

If a change with this name already exists, suggest continuing it (`/opsx:continue`) instead of re-scaffolding.

## Checking what's next

```bash
openspec status --change "<name>"
```
Shows which artifacts are still needed and which are unlocked (dependencies satisfied).

## Surfacing the first artifact

Find the first artifact marked `ready` in the status output — for the default spec-driven schema this is `proposal`, but don't hardcode the name — then:
```bash
openspec instructions <first-artifact-id> --change "<name>"
```
This returns the template and context for that artifact. Show it, don't fill it in.

## Stop here

Don't draft anything yet. Wait for the user's next move.

## What to report

Change name and path, the schema in play and its artifact sequence, current progress (0/N), the first artifact's template, and an invitation: "describe what this change is about and I'll draft it, or ask me to continue."

## Operating notes

- No artifacts get created in this pass — only the scaffold and a template preview.
- Don't advance past the first artifact's template under any circumstance.
- Invalid name (not kebab-case) → ask for a valid one before scaffolding.
- Existing change with the same name → redirect to continuing it.
- Non-default schema requested → always pass `--schema` explicitly, never assume the CLI remembers.

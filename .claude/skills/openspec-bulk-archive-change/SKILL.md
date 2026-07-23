---
name: openspec-bulk-archive-change
description: Close out several finished changes together in one pass, resolving overlapping spec edits by checking what's actually shipped. Trigger when the user has multiple parallel changes ready to retire at once.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.0.2"
---

Archive a batch of finished changes in one operation, resolving any overlapping delta specs by inspecting the codebase rather than guessing.

## Input

None needed upfront — this flow always starts by asking which changes to include.

## Gathering candidates

```bash
openspec list --json
```

No active changes at all → tell the user and exit.

## Choosing the batch

**AskUserQuestion** with multi-select, offering:
- each active change (with its schema shown)
- an "All changes" shortcut

Any count works — one is fine, but the point of this skill is handling several at once. Never preselect anything.

## Collecting per-change status

For every change chosen, gather three things:

1. **Artifact completeness** — `openspec status --change "<name>" --json`, noting `schemaName` and which artifacts are `done` vs not.
2. **Task tally** — read `tasks.md`, count `- [ ]` vs `- [x]`. Missing file → record "No tasks".
3. **Delta specs** — inventory `openspec/changes/<name>/specs/`, and within each capability file pull out every `### Requirement: <name>` line.

## Spotting overlaps

Build a capability → changes-touching-it map:

```
auth -> [change-a, change-b]   (2+ entries = overlap)
api  -> [change-c]             (only one = fine)
```

Any capability claimed by 2+ selected changes needs resolution before archiving.

## Resolving an overlap

Per conflicting capability:

1. Read each change's delta spec for that capability to see what it claims.
2. Grep the codebase for evidence each claim actually landed (files, functions, tests).
3. Decide:
   - Only one side implemented → sync that one only.
   - Both implemented → apply oldest-first, newest wins on overlap.
   - Neither implemented → skip the sync for that capability, flag it.
4. Note the reasoning (what evidence was found) alongside the decision — this feeds the summary table.

### Worked examples

Single winner:
```
Conflict: specs/auth/spec.md claimed by [add-oauth, add-jwt]
add-oauth -> found src/auth/oauth.ts implementing the flow
add-jwt   -> no matching implementation found
=> sync add-oauth only
```

Both landed:
```
Conflict: specs/api/spec.md claimed by [add-rest-api (2026-01-10), add-graphql (2026-01-15)]
Both have matching source files.
=> apply add-rest-api first, then add-graphql (chronological, later wins)
```

## Presenting the batch

Render one table covering every selected change:

```
| Change               | Artifacts | Tasks | Specs   | Conflicts | Status |
|---------------------|-----------|-------|---------|-----------|--------|
| schema-management   | Done      | 5/5   | 2 delta | None      | Ready  |
| project-config      | Done      | 3/3   | 1 delta | None      | Ready  |
| add-oauth           | Done      | 4/4   | 1 delta | auth (!)  | Ready* |
| add-verify-skill    | 1 left    | 2/5   | None    | None      | Warn   |
```

Annotate any conflict resolution below the table, and list warnings for incomplete changes separately.

## Confirming before executing

Single **AskUserQuestion** confirmation covering the whole batch, e.g.:
- "Archive all N changes"
- "Archive only the ready ones, skip incomplete"
- "Cancel"

If incomplete changes are in the mix, spell out that archiving them carries the noted warnings.

## Running the archive

Process changes honoring the conflict-resolution order determined above. Per change:

1. If it has delta specs, run the sync-specs merge logic (respecting resolved ordering for conflicts).
2. Move the folder:
   ```bash
   mkdir -p openspec/changes/archive
   mv openspec/changes/<name> openspec/changes/archive/YYYY-MM-DD-<name>
   ```
3. Record success / failure (with error) / skipped (user opted out).

## Final report

```
## Bulk Archive Complete

Archived N changes:
- <change-1> -> archive/YYYY-MM-DD-<change-1>/
- <change-2> -> archive/YYYY-MM-DD-<change-2>/

Spec sync summary:
- N delta specs synced to main specs
- No conflicts (or: M conflicts resolved)
```

Partial runs list Skipped and Failed sections separately, each with a reason. Nothing selected because no active changes exist:

```
## No Changes to Archive

No active changes found. Use `/opsx:new` to create a new change.
```

## Operating notes

- Batch size is flexible — never force a minimum beyond one.
- Selection always goes through the user, never auto-picked.
- Surface capability conflicts up front and resolve them against real code evidence.
- Chronological ordering (older first) is the tiebreaker when both sides of a conflict are implemented.
- Skip syncing only when nothing is actually implemented — and say so.
- One confirmation covers the entire batch, not one per change.
- Track and report every outcome bucket: archived, skipped, failed.
- Preserve `.openspec.yaml` through the move.
- Archive path always uses today's date: `YYYY-MM-DD-<name>`.
- A naming collision at the target fails just that one change; keep going with the rest.

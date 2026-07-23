---
name: openspec-archive-change
description: Retire a finished change into the archive once its work is done. Trigger when the user is ready to close out a change after implementation wraps up.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.0.2"
---

Move a finished change out of the active workspace and into `openspec/changes/archive/`.

## Input

Change name is optional — infer from context if not given; if genuinely unclear, prompt.

## Step 1: Pick the change

No name supplied? Run `openspec list --json` and hand the list to **AskUserQuestion**. Show only non-archived changes, tagged with their schema. Never guess — the user picks.

## Step 2: Sanity-check artifacts

```bash
openspec status --change "<name>" --json
```

Read `schemaName` and each artifact's status. Anything not `done`? Warn, list what's incomplete, and confirm via **AskUserQuestion** before moving on — a warning, not a hard stop.

## Step 3: Sanity-check tasks

Open the tasks file (usually `tasks.md`) and tally `- [ ]` against `- [x]`. Incomplete tasks found → warn with the count and confirm via **AskUserQuestion**. No tasks file at all → skip this check silently.

## Step 4: Check whether specs need syncing

Look under `openspec/changes/<name>/specs/`. Nothing there → skip straight to archiving.

Delta specs present → diff each against its main spec counterpart at `openspec/specs/<capability>/spec.md`, work out what would change (additions/modifications/removals/renames), and present a combined summary.

Offer choices depending on state:
- Changes pending: "Sync now (recommended)" vs "Archive without syncing"
- Already in sync: "Archive now" / "Sync anyway" / "Cancel"

Choosing sync means running the same logic as openspec-sync-specs. Either way, archiving proceeds afterward.

## Step 5: Move the directory

```bash
mkdir -p openspec/changes/archive
```

Target folder name: `YYYY-MM-DD-<change-name>` using today's date.

Name collision at the target → stop with an error (suggest a rename or different date) rather than overwrite. Otherwise:

```bash
mv openspec/changes/<name> openspec/changes/archive/YYYY-MM-DD-<name>
```

## Step 6: Report the outcome

Summarize: change name, schema, final archive path, whether specs got synced, and any warnings surfaced along the way.

### Success template

```
## Archive Complete

**Change:** <change-name>
**Schema:** <schema-name>
**Archived to:** openspec/changes/archive/YYYY-MM-DD-<name>/
**Specs:** ✓ Synced to main specs (or "No delta specs" or "Sync skipped")

All artifacts complete. All tasks complete.
```

## Operating notes

- Selection is never automatic — always prompt.
- Rely on `openspec status --json` (the artifact graph), don't hand-check completeness.
- Warnings inform, they don't block — the user decides whether to proceed.
- `.openspec.yaml` travels with the directory during the move — don't strip it out.
- Always give a clear closing summary.
- Sync requests route through the openspec-sync-specs agent-driven merge logic.
- Whenever delta specs exist, run the sync assessment and show the summary before asking anything.

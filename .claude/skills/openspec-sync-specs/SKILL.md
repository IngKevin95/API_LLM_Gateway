---
name: openspec-sync-specs
description: Merge a change's delta specs into the main spec files without archiving the change. Trigger when the user wants specs updated mid-flight, ahead of or separate from archiving.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.0.2"
---

Fold delta specs into the main spec tree by hand — intelligent merge, not a file copy, and no archiving involved.

## Input

A change name. If missing, run `openspec list --json` and ask which one via **AskUserQuestion**.

## Reading the delta

Inventory `openspec/changes/<name>/specs/` — one subfolder per capability touched. Each `spec.md` inside is written in delta form, not full-requirement form:

```
## ADDED Requirements
### Requirement: New Thing
The system SHALL ...

#### Scenario: Happy path
- **WHEN** ...
- **THEN** ...

## MODIFIED Requirements
### Requirement: Existing Thing
[full replacement text for this requirement]

## REMOVED Requirements
### Requirement: Old Thing
**Reason**: [why]

## RENAMED Requirements
- FROM: `### Requirement: Old Name`
- TO: `### Requirement: New Name`
```

No delta specs found anywhere under the change → say so and stop.

## Locating the target

For each capability folder in the delta, check `openspec/specs/<capability>/spec.md`.

**Doesn't exist yet** — create it fresh with a `## Purpose` section (infer from the delta's intent, mark TBD if genuinely unclear) and a `## Requirements` section built entirely from the ADDED block.

**Already exists** — read it in full before changing anything. This is the merge target.

## Applying each delta type

- **ADDED** — append the new `### Requirement:` block(s) to the main file's Requirements section. If a requirement of that name already exists, treat it as an implicit MODIFIED instead of duplicating.
- **MODIFIED** — find the matching requirement by name in the main file and merge in the delta's changes: new scenarios get added, changed scenarios get updated, everything not mentioned stays put. This is not a wholesale block replacement.
- **REMOVED** — delete the matching requirement block from the main file. Drop the `**Reason**` line — that's changelog context for reviewers, not spec content.
- **RENAMED** — rename the heading in place; keep the requirement's body as-is unless a MODIFIED block for the same requirement says otherwise.

The core discipline is surgical editing: touch only what the delta names. Add a scenario without rewriting its parent requirement. Adjust one line of a requirement without re-deriving the whole block from scratch.

## Worked example

Delta says:
```
## MODIFIED Requirements
### Requirement: User Login
#### Scenario: Username login
- **WHEN** a user submits a username and password
- **THEN** they are authenticated
```

Main spec currently has "User Login" with only an email-based scenario. The merge adds the username scenario alongside the existing email one — it does not touch the email scenario, and it does not require the delta to have repeated it.

## Doing this for every touched capability

Repeat capability by capability. Nothing here touches `tasks.md`, `design.md`, or the change folder itself — this operation is scoped to specs only.

## Reporting back

```
## Specs Synced: <change-name>

auth:
- Added requirement: "New Feature"
- Modified requirement: "Existing Feature" (added 1 scenario)

billing:
- Created new spec file
- Added requirement: "Another Feature"

Main specs updated. The change stays active — archive separately when implementation is complete.
```

## Operating notes

- Never archive as part of this — that's a separate, explicit step (`/opsx:archive`).
- Read the current main spec before editing it, every time, even on a second sync of the same change.
- Never touch a requirement the delta doesn't mention.
- Running sync twice on an unchanged delta should be a no-op — idempotency matters here.
- Ambiguous merge target (e.g. a requirement renamed and modified in ways that don't obviously reconcile) — ask rather than guess.
- Narrate what's changing as you apply it, not just in the final summary.

---
name: openspec-onboard
description: Teach the full OpenSpec cycle by running it once, live, against a real small task in the user's own codebase. Trigger when a user is new to OpenSpec and wants a guided first pass rather than reading docs.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.0.2"
---

Run one complete change cycle end to end, narrating each step, using a real piece of the user's own codebase rather than a toy example.

## Scope

This is a teaching pass, not a shortcut — every phase below happens, even for a tiny task, because the point is to build the muscle memory for the whole rhythm. Roughly 15-20 minutes of real work.

## Checking readiness

```bash
openspec status --json 2>&1 || echo "NOT_INITIALIZED"
```
Not initialized → tell the user to run `openspec init` first and come back afterward. Stop here in that case.

## Setting the stage

Lay out what's coming before diving in: pick a small real task, explore it briefly, stand up a change container, build proposal → specs → design → tasks, implement, archive. Naming the shape up front means the user isn't surprised by any of the phases that follow.

## Finding real work to do

Scan the codebase for small, genuine opportunities rather than inventing a toy problem:
- `TODO` / `FIXME` / `HACK` / `XXX` comments
- catch blocks that swallow errors, or risky operations with no try/catch at all
- functions in `src/` with no matching coverage in the test directories
- `: any` / `as any` in TypeScript
- stray `console.log`, `console.debug`, `debugger` statements outside debug code
- input handlers with no validation

Cross-reference with recent activity:
```bash
git log --oneline -10 2>/dev/null || echo "No git history"
```

Surface 3-4 concrete candidates, each with a file:line location, a rough size estimate (files touched, lines), and a one-line reason it's a good pick — plus an open "something else, tell me what you want to work on" option. Nothing found in the scan → just ask directly what small thing they've been meaning to fix.

Let the user choose freely, including describing their own task instead of any suggestion.

**If what they pick is clearly too big** (a major feature, multi-day scope): say so plainly, and offer to slice it down to the smallest useful piece, swap to a smaller suggestion, or proceed anyway if they'd rather push through the larger scope. This guardrail bends to the user — if they insist, go with what they picked.

## A taste of explore mode

Before creating anything, spend a minute or two actually looking at the relevant code — read the files involved, sketch an ASCII diagram if the shape benefits from one, note anything worth flagging. Present this as a live demonstration of what `/opsx:explore` is for (thinking before committing to a direction), and mention it's available any time, not just here.

Wait for the user to acknowledge before moving on to creating the change.

## Standing up the change container

Explain briefly: a change is the folder that holds all the thinking for one piece of work — proposal, design, specs, tasks — living at `openspec/changes/<name>/`.

```bash
openspec new change "<derived-name>"
```

Show the resulting empty structure:
```
openspec/changes/<name>/
├── proposal.md    ← why we're doing this
├── design.md      ← how we'll build it
├── specs/         ← detailed requirements
└── tasks.md       ← implementation checklist
```

## Drafting the proposal

Frame it: the proposal is the elevator pitch — why this, what changes, at a high level. Draft the content but hold off on saving:

```
## Why
[1-2 sentences on the problem or opportunity]

## What Changes
[bullet points of what's different afterward]

## Capabilities
### New Capabilities
- <capability-name>: [brief description]

## Impact
- src/path/to/file.ts: [what changes]
```

Ask if this captures the intent before saving. Once confirmed, pull the template and write it for real:
```bash
openspec instructions proposal --change "<name>" --json
```
Save to `openspec/changes/<name>/proposal.md`, and note this document can always be revisited as understanding evolves.

## Drafting specs

Frame it: specs pin down *what* gets built in testable terms — requirement plus scenario, WHEN/THEN/AND, readable almost like a test case. A task this size probably needs just one spec file.

```bash
mkdir -p openspec/changes/<name>/specs/<capability-name>
```

Draft:
```
## ADDED Requirements
### Requirement: <Name>
<description of expected behavior>

#### Scenario: <name>
- **WHEN** <trigger>
- **THEN** <expected outcome>
- **AND** <additional outcome if needed>
```
Save to `openspec/changes/<name>/specs/<capability>/spec.md`.

## Drafting design

Frame it: design captures *how* — the technical approach and tradeoffs. For a task this small, brief is fine; not everything needs a deep design discussion.

```
## Context
[current state, briefly]

## Goals / Non-Goals
**Goals:** [what this achieves]
**Non-Goals:** [explicitly out of scope]

## Decisions
### Decision 1: [the key call]
[approach and why]
```
Save to `openspec/changes/<name>/design.md`.

## Breaking into tasks

Frame it: tasks turn the plan into a checklist that drives implementation — small, clearly worded, in a sensible order.

```
## 1. [file or category]
- [ ] 1.1 [specific task]
- [ ] 1.2 [specific task]

## 2. Verify
- [ ] 2.1 [verification step]
```

Ask if they're ready to implement before moving on. Save to `openspec/changes/<name>/tasks.md`.

## Implementing

Frame it: each task gets implemented and checked off in turn, with light narration tying the work back to specs and design where it's natural ("the spec says X, so this does Y") — not a line-by-line lecture.

For each task: announce it, make the change, flip `- [ ]` to `- [x]` in tasks.md, note completion briefly. When every task is done, confirm the full checklist and point toward archiving as the last step.

## Archiving

Frame it: archiving moves a finished change from `openspec/changes/` into `openspec/changes/archive/YYYY-MM-DD-<name>/` — it becomes part of the project's decision history, findable later to explain why something was built a certain way.

```bash
openspec archive "<name>"
```
Confirm the new path and note the code lives in the codebase while the record lives in the archive.

## Wrapping up

Recap the full rhythm just walked: explore, new, proposal (why), specs (what), design (how), tasks, apply, archive — and note this same rhythm scales from a small fix to a major feature.

Leave a command reference behind:

| Command | What it does |
|---|---|
| `/opsx:explore` | Think through problems before or during work |
| `/opsx:new` | Start a new change, step through artifacts |
| `/opsx:ff` | Fast-forward: create all artifacts at once |
| `/opsx:continue` | Continue working on an existing change |
| `/opsx:apply` | Implement tasks from a change |
| `/opsx:verify` | Verify implementation matches artifacts |
| `/opsx:archive` | Archive a completed change |

Close with an invitation to try `/opsx:new` or `/opsx:ff` on something they actually want to build.

## If the user wants to bail early

Someone signaling they need to stop, pause, or disengage doesn't get pushback. Reassure them the change is saved at `openspec/changes/<name>/`, that `/opsx:continue <name>` picks artifact drafting back up and `/opsx:apply <name>` jumps straight to implementation if tasks already exist, and that nothing is lost by stepping away.

## If the user just wants the command list

Someone asking to skip straight to the reference gets exactly that — the same command table as above, framed as a quick-reference rather than a tutorial recap — and a nudge toward `/opsx:new` or `/opsx:ff` to actually get moving.

## Operating notes

- Explain before doing, do the thing, show the result, pause for acknowledgment — at the transitions that matter: after the explore demo, after the proposal draft, after tasks are drafted, after archiving.
- Narration during implementation stays light — teaching, not lecturing.
- Don't skip a phase because the task is small; the point of this run-through is the full rhythm.
- Never pressure a user who wants to stop.
- Always use a real task from their actual codebase — no simulated examples.
- Scope guardrail nudges toward smaller tasks but yields to what the user actually wants.

---
name: openspec-verify-change
description: Check a change's implementation against its own specs, tasks, and design before archiving. Trigger when the user wants a pre-archive sanity check rather than a blind sign-off.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.0.2"
---

Cross-check what got built against what the change actually promised, across three angles: is it done, is it right, does it hang together.

## Input

A change name, or infer one from context. Genuinely unclear → run `openspec list --json`, surface changes that have a `tasks.md`, tag ones with unchecked boxes "(In Progress)", and let **AskUserQuestion** decide. Never auto-select.

## Loading the artifacts

```bash
openspec status --change "<name>" --json
```
Note `schemaName` and which artifacts actually exist — this determines how much of the verification below is even possible.

```bash
openspec instructions apply --change "<name>" --json
```
Read every file in the returned `contextFiles`.

## The three dimensions

**Completeness** — did everything that was supposed to get built, get built.

**Correctness** — does what got built actually do what the spec says.

**Coherence** — does it hang together with the design and the rest of the codebase.

Each dimension collects issues at one of three severities: CRITICAL (blocks archiving), WARNING (should get fixed), SUGGESTION (nice to have). When genuinely unsure which bucket an issue belongs in, round down in severity rather than up — a false CRITICAL is more disruptive than a missed one.

## Checking completeness

Parse `tasks.md` if present: count `- [x]` against `- [ ]`. Every unchecked box is a CRITICAL, with a recommendation to either finish it or confirm it's actually done and just needs the box checked.

For each capability's delta spec under `openspec/changes/<name>/specs/`, pull every `### Requirement:` and search the codebase for evidence it exists. A requirement with no findable implementation is CRITICAL — "Requirement not found: <name>", with a recommendation naming what to build.

## Checking correctness

For each requirement with implementation evidence, look at the actual code against what the requirement describes. Note file:line references. A plausible-but-divergent implementation is a WARNING, not a CRITICAL — flag the mismatch, cite the location, and recommend reviewing it against the requirement text.

For each `#### Scenario:` under a requirement, check whether the corresponding condition is handled in code and whether a test exercises it. An uncovered scenario is a WARNING with a recommendation to add the missing test or handling.

## Checking coherence

If `design.md` exists, pull out its stated decisions (headings or lines like "Decision:", "Approach:", "Architecture:") and check the implementation actually followed them. A contradiction is a WARNING — either the code drifted or the design doc is stale, and the recommendation should say which seems more likely. No `design.md` → skip this check and say so explicitly rather than silently passing it.

Skim new code for consistency with existing project conventions — naming, file layout, style. Meaningful deviations are SUGGESTIONs with a pointer to the pattern being broken. This is not a style nitpick pass; only flag things that would actually confuse the next reader.

## Adjusting to what exists

Not every change has all three artifact types, and the check should shrink to match:

| Available | What gets verified |
|---|---|
| tasks.md only | Completeness only |
| tasks.md + specs | Completeness + Correctness |
| tasks.md + specs + design.md | All three dimensions |

State plainly which checks ran and which were skipped, and why — never silently pass over a dimension.

## The report

```
## Verification Report: <change-name>

### Summary
| Dimension    | Status            |
|--------------|-------------------|
| Completeness | 4/5 tasks, 3 reqs |
| Correctness  | 2/3 reqs covered  |
| Coherence    | 1 issue           |

### CRITICAL
- Task incomplete: "Add rate limiting middleware" — recommendation: finish or confirm done and check the box.

### WARNING
- Scenario not covered: "Expired token retry" — recommendation: add a test in src/auth/__tests__/token.test.ts.

### SUGGESTION
- src/auth/session.ts doesn't follow the repo's error-wrapping convention used elsewhere in src/auth/.

### Assessment
1 critical issue found. Fix before archiving.
```

Final line changes with outcome: critical issues present → name the count and say fix before archiving; only warnings → note them but say it's ready to archive with noted improvements; everything clean → say so plainly.

## Operating notes

- Every issue needs a specific, actionable recommendation — no "consider reviewing this" without a target.
- Cite file:line wherever there's a concrete location to point to.
- Uncertain severity → round down (SUGGESTION over WARNING, WARNING over CRITICAL).
- Skipped checks get named, not silently dropped.
- This never modifies files — it's a read-only assessment.

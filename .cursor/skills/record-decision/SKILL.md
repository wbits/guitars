---
name: record-decision
description: >-
  Captures product and architecture decisions from the conversation into
  .agents/decisions.md and related project docs. Use when the user says record
  decision, /record-decision, memorize this, remember this for the project,
  save this decision, or asks to persist a design choice for future sessions.
---

# Record decision

Persist decisions from the current conversation so future agent sessions treat them as project memory.

## Trigger phrases

- `record decision` / `/record-decision`
- `memorize this` / `remember this for the project`
- `save this decision` / `add this to decisions`

## Workflow

1. **Extract** — Summarize the decision(s) from the conversation. If ambiguous, ask one clarifying question before writing.
2. **Classify** each item:

   | Kind | Destination |
   |------|-------------|
   | Fixed choice (do not reverse lightly) | `.agents/decisions.md` |
   | Planned work, not yet fixed | `.agents/backlog.md` |
   | Multi-step feature design | `.agents/plans/<topic>.md` (create or extend) |
   | Current status shift | `.agents/CONTEXT.md` (brief bullet under Recent focus / Not started) |
   | Webapp-only UI/UX choice | `guitars-webapp/.agents/decisions.md` (sibling repo) |
   | Cross-cutting (API + webapp) | `.agents/decisions.md` here; one-line pointer in webapp `.agents/decisions.md` if needed |

3. **Write** — Edit files in place. Match existing tone: short bullets, tables where useful, links to plans.
4. **Confirm** — Reply with what was recorded and which files changed.
5. **Do not commit** unless the user explicitly asks.

## Decision entry template

Add under the right `##` section in `decisions.md`:

```markdown
- **Short title** — one or two sentences: what we chose and why (cost, scope, trust, etc.).
```

For tier/feature tables, extend existing tables rather than duplicating sections.

## Rules

- Prefer **decisions** over **plans**: if the user only agreed directionally, use `backlog.md` or a plan file; use `decisions.md` when the choice is settled.
- Never store secrets, API keys, or production tokens.
- Keep entries concise; link to [plans/](plans/) for implementation detail.
- After substantive decision batches, add one line to `.agents/CONTEXT.md` if it changes what “not started” vs “decided” means.
- Related repo: [`wbits/guitars-webapp`](https://github.com/wbits/guitars-webapp) — open that tree when the decision touches the webapp.

## Example

**User:** “Photo analysis should only run for tier 2 BYOK owners.”

**Agent writes to `.agents/decisions.md`:**

```markdown
## Photo analysis (guitars)

- **Generation is tier-2 (BYOK) only** — vision analysis on upload uses the collection owner's API key; operator does not pay for per-upload vision on tier 1.
- **Search uses stored metadata for everyone** — once written, tags/summary are searchable by viewer assistant without extra LLM cost.
- **Human fields stay authoritative** — analysis is advisory; never silently overwrite brand, model, year, etc.
```

**Agent adds to `.agents/backlog.md`:** implementation checklist items.

**Agent replies:** summary + file list; no commit.

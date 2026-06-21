# Decisions

Fixed choices — do not reverse lightly without discussion.

## Architecture

- **DDD layout** — domain invariants in `internal/guitarcollection/domain/`, not in HTTP handlers.
- **Single Guitar aggregate** per bounded context; market logs are separate.
- **Prices** as integer minor units (cents) in JSON; `EUR` and `USD` only.
- **Webapp is a separate repo** — [`wbits/guitars-webapp`](https://github.com/wbits/guitars-webapp).

## Auth

- Production: Cognito JWT; collection scoped to token `sub`.
- Admins: Cognito group `guitars-admins`.
- Clients never send `owner` on POST/PUT; API assigns it.
- Local dev: bearer token via Secrets Manager / LocalStack.

## MCP

- **REST ≠ MCP.** REST routes stay; MCP is an adapter layer.
- **Phase 1:** local stdio server in `mcp/` (this repo).
- **Phase 2:** `/mcp` on API Gateway + Node Lambda, same Cognito auth.
- Zod contract in `mcp/src/contracts/guitar.ts` mirrors Go domain — keep in sync.
- Market crawl via MCP dispatches GitHub Actions `crawl.yml` (no REST trigger).

## guitars-assistant

- **Two profiles:** **viewer** (read-only, collection page) and **curator** (owner manage + MCP power users).
- **Instructions** in `.agents/assistants/` (`viewer.md`, `curator.md`, `shared.md`) — not in MCP server or `.cursor/mcp.json`.
- **`.agents/AGENTS.md`** is for coding agents on this repo; product assistant persona is separate.
- **One tool core** in `mcp/src/tools/`; viewer vs curator = which tools are registered per session.
- **Viewer delivery:** webapp chat only (filters/explains; should drive gallery UI when possible).
- **Viewer chat default closed** — collection assistant shows only the “Ask about this collection” trigger until the user opens it; never auto-expand on page load.
- **v1 gallery filter is client-side** — assistant returns `filter` + `matchingIds`; webapp applies `filterGuitars` locally. No list API query params in v1.
- **Curator delivery:** webapp owner chat + hosted MCP (Phase 2) for Cursor etc.
- **Confirm before write** for create/update and market crawl; web research never auto-writes descriptions.

## guitars-assistant — hosting tiers

| Tier | Status | LLM billing | Rate limit |
|------|--------|-------------|------------|
| **1 — hosted** | Done | Operator (`ASSISTANT_LLM_API_KEY` on Lambda) | Strict daily cap per Cognito `sub` (default 10/day) |
| **2 — BYOK** | Done | Owner API key in settings (encrypted) | Relaxed on operator side; abuse caps only |

Tier 2 BYOK key is used for **curator assistant chat** and **photo analysis generation** (same encrypted owner key; separate opt-in toggles). Tier 1 and tier 2 BYOK are both shipped.

## Photo analysis (vision metadata)

Implemented 2026-06-21 — SQS queue, DLQ, and worker Lambda.

- **Generation is tier-2 (BYOK) only** — auto analysis of guitar photos on upload runs with the collection owner's API key (encrypted in settings). Tier 1 does not trigger operator-paid vision jobs.
- **Owner opt-in** — BYOK key plus an explicit “analyze photos on upload” toggle; upload succeeds even if analysis fails.
- **Search uses stored metadata for everyone** — once `GuitarAnalysis` (or equivalent) exists, viewer assistant and gallery filtering may use tags/summary; no extra vision cost at query time.
- **Human fields stay authoritative** — brand, model, year, price, etc. are curator-entered; AI output is advisory (suggest/accept), never silent overwrite.
- **Async worker** — SQS queue + worker Lambda; API enqueues and returns 202. Re-run only when the cover picture selection changes (`coverPictureIndex` or cover URL), not when other gallery photos change.
- **Cover photo only** — vision analyzes `pictures[coverPictureIndex]` (one image per guitar).
- **Trust** — label AI-detected metadata; show confidence where useful.
- **Sync analyze for add-from-photo** — `POST /me/analyze-photo` runs vision inline during the add-guitar wizard; async worker still handles post-upload re-analysis on existing guitars.

## Collection visibility

- **`hiddenInCollection` (default false)** — owner can hide a guitar from gallery listings while keeping it in their account; open by id still works.
- **Public collections never show hidden guitars** — `GET /collections/{userId}/guitar` filters them out; owner list uses `GET /guitar?includeHidden=true` when needed.
- **Hide/show endpoints** — `POST /guitar/{id}/hide` and `/show`; owner-only.

## Add guitar (BYOK photo-first)

- **Photo-first wizard when BYOK is usable** — upload or URL → analyze → review form; manual form remains available.
- **Price is always human-entered** — AI never guesses price; required before create.
- **Inspect-before-add default** — photo-first create can leave “Add to collection gallery” unchecked; API creates then hides when requested (`hiddenInCollection`).

## Tag search and similarity

- **Similar guitars = shared AI tags** — tag cloud on detail page; `/guitars/similar` and `/collections/{userId}/similar` filter by selected tags client-side.
- **Source guitar stays in similar results** — the guitar you navigated from is included when its tags match.
- **Collection assistant tag hints** — local tag matching in chat complements LLM filter parsing (no extra API in v1).

## Agent session memory

- **Record decisions via skill** — when the user says `/record-decision` (or similar), follow [`.cursor/skills/record-decision/SKILL.md`](../.cursor/skills/record-decision/SKILL.md) and persist to `.agents/decisions.md`, backlog, plans, or `CONTEXT.md`. Do not commit unless asked.
- **`/cpd` skill** — when the user says `/cpd`, follow [`.cursor/skills/cpd/SKILL.md`](../.cursor/skills/cpd/SKILL.md): commit → push → deploy API and/or webapp. Deploy vars live in gitignored `.agents/config/cpd.env` (see [cpd.env.example](config/cpd.env.example)); `/cpd` is explicit permission to ship.

## Git

- Commits and deploy only on user request.
- Do not commit secrets, production tokens, or crawler credentials.

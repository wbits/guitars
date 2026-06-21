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
| **2 — BYOK** | Planned | Owner API key in settings (encrypted) | Relaxed on operator side; abuse caps only |

Tier 2 BYOK key is intended for **curator assistant chat** and **photo analysis generation** (same encrypted owner key; separate opt-in toggles).

Start with tier 1 only. Tier 2 is documented in [plans/guitars-assistant.md](plans/guitars-assistant.md) but not built yet.

## Photo analysis (vision metadata)

Discussed 2026-06-21 — not implemented.

- **Generation is tier-2 (BYOK) only** — auto analysis of guitar photos on upload runs with the collection owner's API key (encrypted in settings). Tier 1 does not trigger operator-paid vision jobs.
- **Owner opt-in** — BYOK key plus an explicit “analyze photos on upload” toggle; upload succeeds even if analysis fails.
- **Search uses stored metadata for everyone** — once `GuitarAnalysis` (or equivalent) exists, viewer assistant and gallery filtering may use tags/summary; no extra vision cost at query time.
- **Human fields stay authoritative** — brand, model, year, price, etc. are curator-entered; AI output is advisory (suggest/accept), never silent overwrite.
- **Async worker** — analyze after save; re-run only when picture set changes (hash diff), not on every field edit.
- **Trust** — label AI-detected metadata; show confidence where useful.

Implementation belongs in API worker + optional curator UI; see [backlog.md](backlog.md).

## Agent session memory

- **Record decisions via skill** — when the user says `/record-decision` (or similar), follow [`.cursor/skills/record-decision/SKILL.md`](../.cursor/skills/record-decision/SKILL.md) and persist to `.agents/decisions.md`, backlog, plans, or `CONTEXT.md`. Do not commit unless asked.

## Git

- Commits and deploy only on user request.
- Do not commit secrets, production tokens, or crawler credentials.

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
- **Curator delivery:** webapp owner chat + hosted MCP (Phase 2) for Cursor etc.
- **Confirm before write** for create/update and market crawl; web research never auto-writes descriptions.

## Git

- Commits and deploy only on user request.
- Do not commit secrets, production tokens, or crawler credentials.

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

## Agent documentation

- **`AGENTS.md`** at repo root → **`.agents/`** for API, crawler, MCP, infra.
- **`guitars-webapp`** has its own `.agents/` for React-only concerns.
- Session notes in `.agents/sessions/YYYY-MM-DD.md`.

## Git

- Commits and deploy only on user request.
- Do not commit secrets, production tokens, or crawler credentials.

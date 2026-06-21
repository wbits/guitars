# Context — last updated 2026-06-21

## What this project is

Go Lambda HTTP API (GuitarCollection) for guitars.com — guitar CRUD, user profiles, collections directory, market logs, picture presign, and weekly marketplace crawler.

## Live & repos

| | |
|---|---|
| **API** | API Gateway + Lambda (this repo) |
| **Webapp** | [`wbits/guitars-webapp`](https://github.com/wbits/guitars-webapp) — React on S3 + CloudFront |
| **GitHub (this repo)** | `wbits/guitars` |
| **Branch** | `master` |
| **Documentation** | `.agents/` |

## Recent focus

- **guitars-assistant tier 1:** shipped — `POST /assistant/chat`, rate limits, webapp viewer chat (closed by default, voice input), client-side gallery filters
- **Photo analysis + tier 2 BYOK:** direction decided; not implemented — see [decisions.md](decisions.md)
- **`/record-decision` skill** — [`.cursor/skills/record-decision/`](../.cursor/skills/record-decision/) for persisting design choices to `.agents/`
- MCP Phase 1 in `mcp/`; agent docs in `.agents/`

## Not started yet

- guitars-assistant tier 2 (owner BYOK)
- Photo analysis on upload (tier-2 BYOK; decision in [decisions.md](decisions.md))
- Curator webapp chat; Phase 2 hosted MCP — see [plans/guitars-assistant.md](plans/guitars-assistant.md)

## Quick verify

```bash
make test
make mcp-test   # requires: cd mcp && npm install first
```

Full setup: [runbook.md](runbook.md).

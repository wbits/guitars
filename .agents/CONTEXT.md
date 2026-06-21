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

- MCP Phase 1 in `mcp/`; agent docs consolidated in `.agents/`
- **guitars-assistant** design: viewer (read-only) + curator (manage) profiles; instructions in `.agents/assistants/`

## Not started yet

- guitars-assistant implementation (webapp chat, Assistant Lambda, new MCP tools) — see [plans/guitars-assistant.md](plans/guitars-assistant.md)
- Phase 2 hosted MCP on API Gateway (Streamable HTTP Lambda)
- Wire MCP in Cursor: `make mcp-build` + `.cursor/mcp.json`

## Quick verify

```bash
make test
make mcp-test   # requires: cd mcp && npm install first
```

Full setup: [runbook.md](runbook.md).

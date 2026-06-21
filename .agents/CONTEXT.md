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

- **Collection visibility** — `hiddenInCollection`, hide/show API, webapp list toggle and detail controls
- **BYOK photo-first add guitar** — `POST /me/analyze-photo`, wizard in webapp, optional seed analysis on create
- **Tag search & similar guitars** — tag cloud, similar routes, collection chat tag matching
- **Agent skills** — `/record-decision`, `/cpd` (commit-push-deploy)
- **guitars-assistant** tier 1 + tier 2 BYOK shipped; async photo analysis worker + sync analyze-photo for add flow
- MCP Phase 1 in `mcp/`; agent docs in `.agents/`

## Not started yet

- Curator webapp chat; Phase 2 hosted MCP — see [plans/guitars-assistant.md](plans/guitars-assistant.md)
- Assistant search over stored analysis metadata (tags/summary at query time)
- Security hardening backlog — see [backlog.md](backlog.md)

## Quick verify

```bash
make test
make mcp-test   # requires: cd mcp && npm install first
```

Full setup: [runbook.md](runbook.md).

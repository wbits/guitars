# AGENTS.md — guitars (GuitarCollection API)

Read this file at the start of a session. This repo is the **GuitarCollection API** — Go Lambda on API Gateway, DynamoDB, Cognito, market crawler, and MCP server.

## Project in one sentence

**guitars** — AWS Lambda-backed HTTP API for a personal guitar collection, with weekly marketplace price crawling and a local MCP adapter for AI agents.

## Read first (order)

| # | File | Why |
|---|------|-----|
| 1 | [`.agents/CONTEXT.md`](CONTEXT.md) | Current status, open work |
| 2 | [`.agents/architecture.md`](architecture.md) | System design, DDD layout, MCP |
| 3 | [`.agents/api-contract.md`](api-contract.md) | HTTP endpoints and payloads |
| 4 | [`.agents/decisions.md`](decisions.md) | Fixed choices |
| 5 | [`.agents/runbook.md`](runbook.md) | Local dev, deploy, crawler, MCP |
| 6 | [`internal/guitarcollection/domain/`](../internal/guitarcollection/domain/) | Go domain source of truth |

After substantive changes, update `.agents/CONTEXT.md` briefly.

## Record decisions (command)

When the user says **`record decision`**, **`/record-decision`**, **`memorize this`**, or **`remember this for the project`**, follow [`.cursor/skills/record-decision/SKILL.md`](../.cursor/skills/record-decision/SKILL.md) and write to `.agents/decisions.md` (and backlog/plans as needed). Do not commit unless asked.

## Commit push deploy (command)

When the user says **`/cpd`**, **`cpd`**, or **`commit push deploy`**, follow [`.cursor/skills/cpd/SKILL.md`](../.cursor/skills/cpd/SKILL.md). Load deploy vars from `.agents/config/cpd.env` (see [cpd.env.example](config/cpd.env.example)). `/cpd` is explicit permission to commit, push, and deploy.

## Related repos

| Repo | Role |
|------|------|
| **`wbits/guitars`** (this repo) | API, crawler, MCP, infra |
| [`wbits/guitars-webapp`](https://github.com/wbits/guitars-webapp) | React static site (S3 + CloudFront) |

The webapp mirrors the API contract in zod at `src/domain/` — keep in sync when changing payloads.

## Folder structure

```
guitars/
├── AGENTS.md
├── cmd/
│   ├── lambda/           # API Gateway handler entry
│   └── crawler/          # market crawl CLI
├── internal/
│   ├── guitarcollection/ # DDD: domain, application, infrastructure, http
│   ├── marketcrawler/
│   └── userprofile/
├── mcp/                  # local MCP stdio server for AI agents
├── .agents/
│   └── assistants/       # guitars-assistant viewer/curator prompts (product, not repo dev)
├── template.yaml         # SAM: API, Lambda, DynamoDB, Cognito
└── .github/workflows/    # ci.yml, crawl.yml
```

## Rules for agents

1. **Minimal diffs** — match existing Go patterns; test-first where the repo already does.
2. **Domain invariants** live in `internal/guitarcollection/domain/` — not in handlers.
3. **Webapp changes** belong in `guitars-webapp` — link or note contract updates in `api-contract.md`.
4. **Tests** — run `make test` for Go changes; `make mcp-test` for MCP.
5. **Commits / deploy** — only when the user explicitly asks.
6. **Update context** — after substantive sessions: `.agents/CONTEXT.md` and optionally `.agents/sessions/YYYY-MM-DD.md`.
7. **Secrets** — never commit tokens, `env.local.json` secrets, or real values in example configs.

## Common tasks

| Task | Action |
|------|--------|
| New/changed API endpoint | domain → application → `interfaces/http/` → `api-contract.md` → note webapp zod sync |
| Market crawler change | `internal/marketcrawler/`, `cmd/crawler/` |
| MCP tool change | `mcp/src/tools/`, `mcp/README.md`, rebuild with `make mcp-build` |
| Deploy API | `S3_BUCKET=… STACK_NAME=… make deploy` — see [runbook.md](runbook.md) |
| Phase 2 hosted MCP | SAM template + new Lambda — see [plans/mcp-server.md](plans/mcp-server.md) (Phase 1: [`mcp/README.md`](../../mcp/README.md)) |
| guitars-assistant | Prompts in [assistants/](assistants/); plan [plans/guitars-assistant.md](plans/guitars-assistant.md); tools in `mcp/src/tools/` |

## Open work

See [`.agents/backlog.md`](backlog.md).

# Backlog

## guitars-assistant

Plan: [plans/guitars-assistant.md](plans/guitars-assistant.md). Instructions: [assistants/](assistants/).

- [x] Tier 1: `POST /assistant/chat` + DynamoDB rate limit + rule/LLM parse
- [x] Webapp viewer chat on collection pages (`guitars-webapp`)
- [x] Client-side `filterGuitars` (no list API query params)
- [ ] Optional: set `ASSISTANT_LLM_API_KEY` in production for natural-language beyond rules
- [x] Tier 2: owner BYOK in webapp profile settings + encrypted storage on API
- [ ] Production: set `ASSISTANT_BYOK_ENCRYPTION_KEY` (32-byte base64) to enable BYOK in AWS
- [x] Photo analysis on upload (tier-2 BYOK only) — async vision job → `GuitarAnalysis` metadata
- [x] Sync `POST /me/analyze-photo` for BYOK add-from-photo wizard
- [x] Collection visibility (`hiddenInCollection`, hide/show)
- [x] Tag-based similar guitars + collection chat tag matching (webapp)
- [ ] Assistant search over analysis metadata (tags + summary; cheap at query time)
- [ ] Curator webapp chat + hosted MCP (Phase 2)
- [ ] `research_guitar`, `presign_upload` MCP tools
- [ ] Cursor rule/skill for `assistants/curator.md`

## MCP — Phase 2 (hosted)

- [ ] API Gateway route `POST /mcp` (Streamable HTTP)
- [ ] Node.js Lambda with MCP SDK
- [ ] Reuse Cognito JWT authorizer
- [ ] User docs: connect any MCP client with Cognito token

## API / crawler

- [ ] `POST /admin/market-crawl` REST endpoint (alternative to GitHub Actions trigger)

## Security hardening (review 2026-06-21)

- [ ] Close legacy **ownerless guitar** claim path — guitars with empty `owner` are writable by any authenticated user until claimed
- [ ] Restrict **market-log presign** — `POST /upload/presign` with `purpose: market-log` should be crawler/admin-only, not any signed-in user
- [ ] Consider redacting **owner email** on `GET /collections` (or expose username/displayName only)
- [ ] Validate **analyze-photo URLs** — allowlist CDN/S3 origins where possible before sending to vision API
- [ ] Document that **CORS `*`** is acceptable because auth is bearer-only (no cookie sessions)

## Webapp (track in guitars-webapp)

- [ ] UI to toggle `marketCrawlEnabled` (API endpoint exists)
- [ ] Add-guitar review form: clearer validation when price missing

## MCP Phase 1 — done

See [`mcp/`](../mcp/) and [`mcp/README.md`](../mcp/README.md).

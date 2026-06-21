# Backlog

## guitars-assistant

Plan: [plans/guitars-assistant.md](plans/guitars-assistant.md). Instructions: [assistants/](assistants/).

- [x] Tier 1: `POST /assistant/chat` + DynamoDB rate limit + rule/LLM parse
- [x] Webapp viewer chat on collection pages (`guitars-webapp`)
- [x] Client-side `filterGuitars` (no list API query params)
- [ ] Optional: set `ASSISTANT_LLM_API_KEY` in production for natural-language beyond rules
- [ ] Tier 2: owner BYOK in webapp settings
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

## Webapp (track in guitars-webapp)

- [ ] UI to toggle `marketCrawlEnabled` (API endpoint exists)

## MCP Phase 1 — done

See [`mcp/`](../mcp/) and [`mcp/README.md`](../mcp/README.md).

# Backlog

## guitars-assistant

Plan: [plans/guitars-assistant.md](plans/guitars-assistant.md). Instructions: [assistants/](assistants/).

- [ ] `search_collection` tool (brand, price range, year filters)
- [ ] Viewer read-only tools (`list_collection_guitars`, public `get_guitar` if needed)
- [ ] Assistant Lambda + `POST /assistant/chat` (or webapp-only backend)
- [ ] Webapp viewer chat on collection pages (`guitars-webapp`)
- [ ] Webapp curator chat on owner dashboard (`guitars-webapp`)
- [ ] `research_guitar` fetch-only tool + confirm-before-update flow
- [ ] `presign_upload` MCP tool
- [ ] Cursor rule/skill pointing power users at `assistants/curator.md`

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

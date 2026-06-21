# guitars-mcp

Local MCP server (stdio) for the GuitarCollection API. Lets AI agents list, create, and update guitars in your collection.

## Prerequisites

- Node 20+
- `GUITARS_API_BASE_URL` and `GUITARS_BEARER_TOKEN` (Cognito ID token or dev bearer token)
- Optional: `gh` CLI + auth for `trigger_market_crawl`

## Build

```bash
cd mcp
npm install
npm run build
```

From repo root:

```bash
npm run mcp:build
```

## Tools

| Tool | Description |
|------|-------------|
| `list_guitars` | List your collection |
| `get_guitar` | Get one guitar by id |
| `create_guitar` | Add a guitar (`price` in major units, e.g. `1999.00`) |
| `update_guitar` | Full replace by id |
| `trigger_market_crawl` | Dispatch GitHub Actions crawl workflow |

## Cursor configuration

Copy [`.agents/config/mcp.json.example`](../.agents/config/mcp.json.example) to `.cursor/mcp.json` in this repo (or merge into your user MCP config):

```json
{
  "mcpServers": {
    "guitars": {
      "command": "node",
      "args": ["${workspaceFolder}/mcp/dist/index.js"],
      "env": {
        "GUITARS_API_BASE_URL": "https://your-api.example.com",
        "GUITARS_BEARER_TOKEN": "your-token-here"
      }
    }
  }
}
```

Do **not** commit real tokens.

### Environment variables

| Variable | Required | Purpose |
|----------|----------|---------|
| `GUITARS_API_BASE_URL` | Yes | API base URL (no trailing slash) |
| `GUITARS_BEARER_TOKEN` | Yes | Bearer token for API auth |
| `GITHUB_REPO` | No | Default `wbits/guitars` for crawl dispatch |
| `GH_TOKEN` / `GITHUB_TOKEN` | No | Used by `gh` when triggering crawls |

## Test

```bash
cd mcp && npm test
```

## Architecture notes

Tool handlers and the API client are transport-agnostic so they can be reused for Phase 2 (hosted MCP on API Gateway). See [`.agents/plans/mcp-server.md`](../.agents/plans/mcp-server.md).

# Runbook

Operations for the **guitars** API, crawler, and MCP server.

## Prerequisites

- Go 1.22+
- Docker + Docker Compose (LocalStack)
- AWS SAM CLI + AWS CLI
- Node 20+ (for `mcp/` only)
- `gh` CLI (optional — MCP crawl trigger or manual workflow dispatch)

## API — local development

```bash
make test
make localstack-up
make localstack-init
make api
```

API at <http://127.0.0.1:3000>:

```bash
curl -H "Authorization: Bearer local-dev-token" http://127.0.0.1:3000/guitar
curl -H "Authorization: Bearer local-dev-token" -H "Content-Type: application/json" \
  -d '{"collectionUserId":"local-dev-user","message":"Fenders under 2000 euro"}' \
  http://127.0.0.1:3000/assistant/chat
```

## Viewer assistant (tier 1)

Hosted on the API Lambda. Without `ASSISTANT_LLM_API_KEY`, uses rule-based parsing only.

| Variable | Default | Purpose |
|----------|---------|---------|
| `ASSISTANT_USAGE_TABLE` | (empty → in-memory) | DynamoDB daily counter table |
| `ASSISTANT_DAILY_LIMIT` | `10` | Max messages per Cognito `sub` per UTC day |
| `ASSISTANT_LLM_API_KEY` | (empty) | OpenAI-compatible API key (operator account) |
| `ASSISTANT_LLM_BASE_URL` | OpenAI | Compatible API base URL |
| `ASSISTANT_LLM_MODEL` | `gpt-4o-mini` | Chat model |

Tier 2 (owner BYOK) is planned, not implemented. See [plans/guitars-assistant.md](plans/guitars-assistant.md).

Webapp: chat on `/collections/:userId` in **guitars-webapp**.

## API — tests, lint, CI

| Command | Effect |
|---------|--------|
| `make test` | Go unit tests |
| `make test-cover` | Coverage profile (`coverage.out`) |
| `make lint` | golangci-lint |

CI (`.github/workflows/ci.yml`) on push/PR to `master`:

| Step | Tool | Purpose |
|------|------|---------|
| Static analysis | golangci-lint | Bugs, style, common Go issues |
| Unit tests | `go test` | Full suite with race detector |
| Coverage gate | octocov | ≥ 80% coverage on new/changed lines |

The first CI run on a branch establishes a baseline; subsequent pushes enforce the diff-coverage threshold.

## API — deploy to AWS

```bash
S3_BUCKET=your-bucket STACK_NAME=guitars make deploy
```

Provisions API Gateway, Lambda, DynamoDB, Cognito, Secrets Manager. Rotate the bearer token in Secrets Manager after deploy; Lambda refreshes within five minutes.

## Admin role and market crawl opt-in

Admins: Cognito group **`guitars-admins`** (from `template.yaml`). JWT includes `cognito:groups`; `GET /me` exposes `isAdmin`.

Assign admins in AWS Console → Cognito → Groups → guitars-admins. Local bearer auth: `LOCAL_DEV_ADMIN_GROUPS=guitars-admins`.

Each collection has `marketCrawlEnabled` (default **false**). Crawler only scans opted-in collections.

```http
PATCH /collections/{userId}/market-crawl
Authorization: Bearer <admin token>
Content-Type: application/json

{"marketCrawlEnabled": true}
```

Clear all market logs for a collection (admin):

```http
DELETE /collections/{userId}/market-log
Authorization: Bearer <admin token>
```

→ `{"deletedCount": 42}`

See [api-contract.md](api-contract.md) for full endpoint list.

## Market crawler

Workflow: [`.github/workflows/crawl.yml`](../.github/workflows/crawl.yml). Reverb only in CI (`-skip-ebay -skip-marktplaats`).

| Trigger | When |
|---------|------|
| Schedule | Sunday 06:00 UTC |
| Manual | Actions → Market crawl → Run workflow |
| Push | Changes under `cmd/crawler/` or `internal/marketcrawler/` |

### GitHub configuration

| Kind | Name | Purpose |
|------|------|---------|
| Variable | `GUITARS_API_URL` | Production API base URL |
| Variable | `COGNITO_CLIENT_ID` | Cognito app client ID |
| Variable | `COGNITO_REGION` | e.g. `eu-central-1` |
| Secret | `COGNITO_CRAWLER_USERNAME` | Crawler Cognito user |
| Secret | `COGNITO_CRAWLER_PASSWORD` | Must match Cognito password |
| Secret | `REVERB_API_TOKEN` | Reverb token with `public` scope |
| Secret | `EBAY_CLIENT_ID` / `EBAY_CLIENT_SECRET` | Optional; local runs |

### Reverb API token

1. [reverb.com](https://reverb.com) → My Profile → API & Integrations → Generate New Token
2. Enable **`public`** scope
3. Store in `REVERB_API_TOKEN` GitHub secret (never commit)

Local:

```bash
REVERB_API_TOKEN=your-token make crawl
```

Password from Secrets Manager locally:

```bash
COGNITO_PASSWORD_SECRET_ID=guitars/crawler-cognito-password make crawl
```

Crawler account defaults to `info@wbits.net`. Override with `MARKET_CRAWLER_EMAIL` / `MARKET_CRAWLER_USER_ID` on the API Lambda.

## MCP server

Phase 1 (local stdio): [`mcp/README.md`](../mcp/README.md).

```bash
cd mcp && npm install    # first time
make mcp-build
make mcp-test
```

| Variable | Required | Purpose |
|----------|----------|---------|
| `GUITARS_API_BASE_URL` | Yes | API base URL (no trailing slash) |
| `GUITARS_BEARER_TOKEN` | Yes | Cognito ID token or dev bearer |
| `GITHUB_REPO` | No | Default `wbits/guitars` for crawl dispatch |

Cursor: [`.agents/config/mcp.json.example`](config/mcp.json.example).

Phase 2 (hosted): [plans/mcp-server.md](plans/mcp-server.md).

## Related

Webapp dev/deploy: [`guitars-webapp` runbook](https://github.com/wbits/guitars-webapp/blob/master/.agents/runbook.md).

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
```

## API — tests & lint

| Command | Effect |
|---------|--------|
| `make test` | Go unit tests |
| `make test-cover` | Coverage profile |
| `make lint` | golangci-lint |

CI: `.github/workflows/ci.yml` on push/PR to `master`.

## API — deploy to AWS

```bash
S3_BUCKET=your-bucket STACK_NAME=guitars make deploy
```

Stack provisions API Gateway, Lambda, DynamoDB, Cognito, secrets. See [`README.md`](../README.md) for admin/crawler configuration.

## Market crawler

Local:

```bash
REVERB_API_TOKEN=your-token make crawl
```

GitHub Actions: `.github/workflows/crawl.yml` — weekly + manual dispatch. Secrets/vars documented in [`README.md`](../README.md#market-crawler-github-actions).

## MCP server

```bash
cd mcp && npm install    # first time
make mcp-build
make mcp-test
```

### Environment variables

| Variable | Required | Purpose |
|----------|----------|---------|
| `GUITARS_API_BASE_URL` | Yes | API base URL (no trailing slash) |
| `GUITARS_BEARER_TOKEN` | Yes | Cognito ID token or dev bearer |
| `GITHUB_REPO` | No | Default `wbits/guitars` for crawl dispatch |

Cursor config: [`.agents/config/mcp.json.example`](config/mcp.json.example) — see [`mcp/README.md`](../mcp/README.md).

## Related

Webapp dev/deploy: [`guitars-webapp` runbook](https://github.com/wbits/guitars-webapp/blob/master/.agents/runbook.md).

# Guitars

AWS Lambda API for the **guitars.com** guitar collection — API Gateway, DynamoDB, Cognito, weekly market crawler, and a local MCP adapter for AI agents.

**Related:** React webapp at [`wbits/guitars-webapp`](https://github.com/wbits/guitars-webapp).

## Quick start

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

## Documentation

| Audience | Start here |
|----------|------------|
| **AI agents** | [`AGENTS.md`](AGENTS.md) → [`.agents/AGENTS.md`](.agents/AGENTS.md) |
| **API contract** | [`.agents/api-contract.md`](.agents/api-contract.md) |
| **Architecture** | [`.agents/architecture.md`](.agents/architecture.md) |
| **Dev, deploy, CI, crawler, MCP** | [`.agents/runbook.md`](.agents/runbook.md) |
| **MCP server** | [`mcp/README.md`](mcp/README.md) |
| **Current status** | [`.agents/CONTEXT.md`](.agents/CONTEXT.md) |

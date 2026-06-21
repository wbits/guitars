# guitars-assistant

End-user AI assistants for guitars.com — **separate from** [`.agents/AGENTS.md`](../AGENTS.md), which documents how **coding agents** work on this repo.

## Two audiences

| Profile | User | Intent | Mutations |
|---------|------|--------|-----------|
| **viewer** | Someone browsing a collection | Explore, filter, compare, ask questions | None |
| **curator** | Collection owner (or admin) | Manage, enrich, trigger jobs | Create, update, crawl, upload |

Example prompts:

- **Viewer:** “Show all Fenders in this collection”, “Guitars between €500 and €1000”
- **Curator:** “Add this guitar to my collection”, “Find documentation and update the description”, “Run a market scan for this guitar”

## Where instructions live

| Layer | Location | Contains |
|-------|----------|----------|
| Persona & policy | [`viewer.md`](viewer.md), [`curator.md`](curator.md), [`shared.md`](shared.md) | Role, guardrails, tone |
| Callable actions | [`mcp/src/tools/`](../../mcp/src/tools/) | Tool handlers + short MCP descriptions |
| Repo dev agents | [`.agents/AGENTS.md`](../AGENTS.md) | How to develop the API/MCP (not product behavior) |
| Client wiring | `.cursor/mcp.json`, webapp assistant API | Connection only — not behavior |

MCP has **no system prompt**. Tool descriptions in `mcp/src/tools/` say what each tool does; persona lives in the markdown files above.

## How each surface loads instructions

| Surface | Profile | Instructions | Tools |
|---------|---------|--------------|-------|
| Webapp chat on collection page | viewer | `viewer.md` + `shared.md` | Read-only tool set |
| Webapp owner dashboard chat | curator | `curator.md` + `shared.md` | Read + write tool set |
| Cursor / MCP client (Phase 2) | curator | Cursor rule or skill → `curator.md` | Hosted MCP, curator tools |

Implementation plan: [`.agents/plans/guitars-assistant.md`](../plans/guitars-assistant.md).

## Hosting tiers

| Tier | Status | Who pays for LLM |
|------|--------|------------------|
| **1 — hosted** | Implemented | Operator (`ASSISTANT_LLM_API_KEY` on Lambda) |
| **2 — BYOK** | Planned | Collection owner (optional settings key) |

Tier 1 uses strict daily rate limits per authenticated user (default 10/day).

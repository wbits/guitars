# API contract

Server-side reference for the GuitarCollection HTTP API. Implementation: `internal/guitarcollection/interfaces/http/`.

All requests:

- `Accept: application/json`
- `Authorization: Bearer <token>`
- `Content-Type: application/json` (when body present)

Errors: `{ "error": "message" }` with appropriate HTTP status.

## Guitars

| Method | Path | Description |
|--------|------|-------------|
| GET | `/guitar` | List guitars owned by authenticated user (hidden omitted by default; see query) |
| GET | `/guitar/{id}` | Single guitar |
| POST | `/guitar` | Create (triggers photo analysis when owner opted in) |
| PUT | `/guitar/{id}` | Full replace (re-analyzes when pictures change) |
| DELETE | `/guitar/{id}` | Delete |
| POST | `/guitar/{id}/hide` | Hide from collection listings (owner only) |
| POST | `/guitar/{id}/show` | Show in collection listings (owner only) |
| POST | `/guitar/{id}/analyze` | Re-run photo analysis (owner only) |

Go domain: [`internal/guitarcollection/domain/guitar.go`](../internal/guitarcollection/domain/guitar.go).

### Example POST/PUT body

```json
{
  "brand": "Fender",
  "typeName": "Stratocaster",
  "buildYear": 1996,
  "priceAmount": 199900,
  "priceCurrency": "EUR",
  "pictures": ["https://example.com/front.jpg"],
  "coverPictureIndex": 0,
  "serialNumber": "SN-12345",
  "description": "1996 sunburst"
}
```

### POST/PUT body

Clients **do not send** `id` or `owner`; the API assigns them.

| Field | Type | Notes |
|-------|------|-------|
| `brand` | string | Required |
| `typeName` | string | Required |
| `buildYear` | int | 1800 … current year + 1 |
| `priceAmount` | int | **Minor units** (cents) |
| `priceCurrency` | `"EUR"` \| `"USD"` | |
| `pictures` | string[] | Absolute http(s) URLs |
| `coverPictureIndex` | int | Index into `pictures` for thumbnail |
| `serialNumber` | string? | |
| `color` | string? | |
| `country` | string? | |
| `factory` | string? | |
| `description` | string? | Optional HTML |

### Response

Same fields plus `id` and optional `owner` (Cognito `sub`).

| Field | Type | Notes |
|-------|------|-------|
| `hiddenInCollection` | boolean? | When `true`, omitted from public and default owner list views; owner can still open by id |

### GET `/guitar` query

| Param | Values | Notes |
|-------|--------|-------|
| `includeHidden` | `true` / `1` | Owner list includes hidden guitars; default omits them |

Public collection list (`GET /collections/{userId}/guitar`) never includes hidden guitars. Owners can always read hidden guitars by id; other users receive 404.

Optional `analysis` object (AI-detected, advisory):

| Field | Type | Notes |
|-------|------|-------|
| `status` | `"pending"` \| `"ready"` \| `"failed"` | |
| `visualSummary` | string? | Present when `ready` |
| `tags` | string[]? | Lowercase kebab-case visual tags |
| `confidence` | number? | 0–1 when `ready` |
| `failureReason` | string? | When `failed` |
| `analyzedAt` | string? | RFC3339 when `ready` |

Photo analysis runs after create/update when the owner has BYOK configured and `photoAnalysisEnabled: true`. Re-runs only when the cover picture selection changes (`coverPictureIndex` or the URL at that index). Vision analyzes the cover photo only.

### Legacy records

Guitars without `owner` are hidden from list until updated by an authenticated user (backfills ownership).

## Current user

| Method | Path | Description |
|--------|------|-------------|
| GET | `/me` | Profile + `isAdmin` + assistant BYOK status |
| PATCH | `/me` | Update username and/or photo analysis opt-in |
| PUT | `/me/assistant-byok` | Store encrypted owner LLM API key (tier 2) |
| DELETE | `/me/assistant-byok` | Remove owner LLM API key |
| POST | `/me/reanalyze-collection` | Re-run photo analysis for all owned guitars with pictures (BYOK required) |

### GET `/me` — assistant fields

| Field | Type | Notes |
|-------|------|-------|
| `assistantByokConfigured` | boolean | True when an encrypted key is stored |
| `assistantLlmBaseUrl` | string? | Optional OpenAI-compatible base URL |
| `assistantLlmModel` | string? | Optional model name |
| `photoAnalysisEnabled` | boolean | True when owner opted in and BYOK is configured |

### PATCH `/me` body

```json
{
  "username": "my-handle",
  "photoAnalysisEnabled": true
}
```

`photoAnalysisEnabled` requires a configured assistant API key (400 otherwise). Clearing BYOK disables photo analysis.

### POST `/me/reanalyze-collection`

Enqueues cover-photo analysis for every owned guitar with a valid cover image. Requires a stored assistant API key. Does not require `photoAnalysisEnabled` (explicit manual action). Returns immediately; vision runs in a background worker.

Response **202**:

```json
{
  "total": 5,
  "queued": 4,
  "skipped": 1
}
```

| Status | Meaning |
|--------|---------|
| 202 | Jobs enqueued (vision runs asynchronously) |
| 400 | No assistant API key configured, or BYOK credentials need re-saving |
| 503 | Photo analysis or queue not configured |

### POST `/guitar/{id}/analyze`

Enqueues cover-photo analysis for one guitar (owner only). Returns **202** with the guitar object and `analysis.status: "pending"`.

### PUT `/me/assistant-byok` body

```json
{
  "apiKey": "sk-…",
  "baseUrl": "https://api.openai.com/v1",
  "model": "gpt-4o-mini"
}
```

Returns the same shape as GET `/me`. Requires `ASSISTANT_BYOK_ENCRYPTION_KEY` on the server (503 if unset).

## Collections

| Method | Path | Description |
|--------|------|-------------|
| GET | `/collections` | Public directory of collection owners |
| GET | `/collections/{userId}/guitar` | Another user's guitars |
| PATCH | `/collections/{userId}/market-crawl` | Admin: toggle `marketCrawlEnabled` |
| DELETE | `/collections/{userId}/market-log` | Admin: clear all market logs |

## Market logs

| Method | Path | Description |
|--------|------|-------------|
| GET | `/guitar/{id}/market-log` | List observations |
| POST | `/guitar/{id}/market-log` | Append (owner or crawler account) |

Populated by `cmd/crawler` (weekly GitHub Actions). Crawler writes only when owner has `marketCrawlEnabled: true`.

## Assistant (viewer)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/assistant/chat` | Natural language → filter + matching guitar ids |

**Tier 1 (hosted):** operator LLM key (`ASSISTANT_LLM_API_KEY`); strict daily cap per caller (`ASSISTANT_DAILY_LIMIT`, default 10/day).

**Tier 2 (owner BYOK):** when the caller browses **their own** collection and has stored a key via `/me/assistant-byok`, the server uses the owner's decrypted key and `ASSISTANT_BYOK_DAILY_LIMIT` (default 200/day). Visitors on that collection still use tier 1.

### POST `/assistant/chat` body

```json
{
  "collectionUserId": "cognito-sub-or-local-dev-user",
  "message": "Show Fenders under 1000 euro"
}
```

### Response

```json
{
  "message": "Filtering for Fender, under €1000. Showing 2 of 5.",
  "matchingIds": ["id-1", "id-2"],
  "filter": {
    "brand": "Fender",
    "maxPriceMajor": 1000
  }
}
```

| Status | Meaning |
|--------|---------|
| 429 | Daily assistant quota exceeded (`ASSISTANT_DAILY_LIMIT`, default 10/day) |
| 503 | Assistant not configured |

Filter fields use **major** price units in JSON. Optional `tag` and `searchText` match stored AI analysis metadata (`GuitarAnalysis` table). Gallery filtering in the webapp mirrors the same shape. When `ASSISTANT_LLM_API_KEY` is unset, the server uses rule-based parsing only.

## Uploads

| Method | Path | Description |
|--------|------|-------------|
| POST | `/upload/presign` | Presigned URL for guitar picture upload |

## Client mirror

The React webapp mirrors this contract in zod at [`guitars-webapp/src/domain/`](https://github.com/wbits/guitars-webapp/tree/master/src/domain). MCP uses [`mcp/src/contracts/guitar.ts`](../mcp/src/contracts/guitar.ts). Update both when changing payloads.

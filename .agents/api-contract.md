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
| GET | `/guitar` | List guitars owned by authenticated user |
| GET | `/guitar/{id}` | Single guitar |
| POST | `/guitar` | Create |
| PUT | `/guitar/{id}` | Full replace |
| DELETE | `/guitar/{id}` | Delete |

Go domain: [`internal/guitarcollection/domain/guitar.go`](../internal/guitarcollection/domain/guitar.go).

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

### Legacy records

Guitars without `owner` are hidden from list until updated by an authenticated user (backfills ownership).

## Current user

| Method | Path | Description |
|--------|------|-------------|
| GET | `/me` | Profile + `isAdmin` |
| PATCH | `/me` | Update username |

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

## Uploads

| Method | Path | Description |
|--------|------|-------------|
| POST | `/upload/presign` | Presigned URL for guitar picture upload |

## Client mirror

The React webapp mirrors this contract in zod at [`guitars-webapp/src/domain/`](https://github.com/wbits/guitars-webapp/tree/master/src/domain). MCP uses [`mcp/src/contracts/guitar.ts`](../mcp/src/contracts/guitar.ts). Update both when changing payloads.

# Guitars

A small AWS-Lambda-backed API that exposes my personal guitar collection.

## Endpoints

All endpoints are protected by a shared bearer token. Send it as
`Authorization: Bearer <token>` on every request.

| Method | Path             | Description                       |
| ------ | ---------------- | --------------------------------- |
| GET    | `/guitar`        | list guitars owned by the authenticated user |
| GET    | `/guitar/{id}`   | retrieve a single guitar          |
| POST   | `/guitar`        | add a new guitar                  |
| PUT    | `/guitar/{id}`   | replace an existing guitar        |
| DELETE | `/guitar/{id}`   | remove a guitar                   |

The JSON body for POST/PUT looks like:

```json
{
  "serialNumber": "SN-12345",
  "pictures": ["https://example.com/front.jpg"],
  "description": "1996 sunburst",
  "brand": "Fender",
  "typeName": "Stratocaster",
  "buildYear": 1996,
  "priceAmount": 199900,
  "priceCurrency": "EUR"
}
```

Responses include `"owner"` (the Cognito user id). Clients do not send `owner`
in POST/PUT bodies; the API assigns it from the authenticated user. Listing
returns only guitars owned by the caller. Legacy guitars without an owner are
hidden from the list until the next update, which backfills ownership.

Prices are stored in **minor units** (cents). `199900` therefore means
EUR&nbsp;1999,00. The only currencies currently accepted are `EUR` and `USD`.

## Architecture

The project follows a small Domain-Driven Design layout, built test-first.

```
cmd/lambda/                       # lambda entry point (provided.al2)
internal/guitarcollection/
    domain/                       # Guitar aggregate, Money VO, repository port
    application/                  # use-case service
    infrastructure/
        persistence/              # in-memory + DynamoDB adapters
        auth/                     # bearer-token authenticator (Secrets Manager)
    interfaces/http/              # API Gateway proxy adapter
template.yaml                     # SAM/CloudFormation: API Gateway + Lambda + DynamoDB + Secret
docker-compose.yml                # LocalStack for local development
scripts/localstack-init.sh        # creates the local table + secret
```

The bounded context is **GuitarCollection** and the single aggregate is
**Guitar**. All invariants live in `domain/guitar.go`; the application service
in `application/service.go` orchestrates use-cases without owning business
rules.

A future scraping subsystem (eBay, Marktplaats, Reverb) can be added as a
separate package inside `internal/` consuming the same `domain.Repository`
port, with its own application service.

## Storage

`DynamoDB` with a single table (`Guitars`) keyed by `id` (string). A single
table is sufficient because the model has no relationships. If we later add a
"market listings" aggregate from the scrapers, a separate table is the
preferred next step rather than overloading this one.

## Authentication

Requests must carry a bearer token. The expected value is stored in AWS
Secrets Manager (secret name `guitars/bearer-token` by default) and cached in
the Lambda for five minutes between fetches. A future iteration is expected to
graduate to OAuth.

## Running locally with LocalStack

Prerequisites: Go 1.22+, Docker (with Compose), the [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html),
and the AWS CLI.

```bash
# 1. Run the unit tests
make test

# 2. Boot LocalStack (DynamoDB + Secrets Manager)
make localstack-up

# 3. Create the Guitars table and the bearer-token secret
make localstack-init

# 4. Compile the lambda binary and start the API locally
make api
```

The API is then reachable at <http://127.0.0.1:3000>. Example:

```bash
curl -H "Authorization: Bearer local-dev-token" http://127.0.0.1:3000/guitar
```

## Tests

```bash
make test
```

The suite covers every layer in isolation:

- domain entity invariants and the `Money` value object,
- application service use-cases through a fake repository,
- the in-memory and DynamoDB repository implementations,
- the bearer authenticator (with token caching and TTL refresh),
- the API Gateway adapter (full CRUD lifecycle plus auth failure modes).

## Continuous integration

Every push and pull request to `master` runs [`.github/workflows/ci.yml`](.github/workflows/ci.yml):

| Step | Tool | Purpose |
| ---- | ---- | ------- |
| Static analysis | [golangci-lint](https://golangci-lint.run/) | Bugs, style, and common Go issues (free, no signup) |
| Unit tests | `go test` | Full test suite with race detector |
| Coverage gate | [octocov](https://github.com/k1LoW/octocov) | **≥ 80% coverage on new/changed code** |

This is a free SonarCloud-style setup: lint + test coverage with a diff
coverage gate, without a paid SonarCloud subscription. octocov stores the
previous report as a GitHub Actions artifact and compares your branch against
it. Pull requests get a summary comment with coverage on changed lines.

Run the same checks locally:

```bash
make test-cover   # writes coverage.out (cmd/ entrypoints excluded)
make lint         # requires golangci-lint on PATH
```

The first CI run on a branch establishes a baseline; subsequent pushes enforce
the 80% diff-coverage threshold on lines you change.

## Admin role and market crawl opt-in

Administrators belong to the Cognito user pool group **`guitars-admins`**
(provisioned by `template.yaml`). Group membership is included in JWT access
and ID tokens as the `cognito:groups` claim. The API exposes `isAdmin` on
`GET /me` so clients can show admin UI when needed.

**Assigning admins:** add users to the `guitars-admins` group in the AWS
Console (Cognito → User pools → guitars → Groups → guitars-admins → Add user).
That manual assignment is the recommended approach for a small, fixed admin set.
No self-service promotion path exists in the app.

For local bearer-token auth, set `LOCAL_DEV_ADMIN_GROUPS=guitars-admins` to
simulate admin membership.

Each user collection has a `marketCrawlEnabled` flag on the user profile
(default **`false`**). The scheduled market crawler only scans collections
where an admin has enabled this flag. Admins toggle it with:

```http
PATCH /collections/{userId}/market-crawl
Authorization: Bearer <admin token>
Content-Type: application/json

{"marketCrawlEnabled": true}
```

Admins can wipe every market log for a collection and start fresh:

```http
DELETE /collections/{userId}/market-log
Authorization: Bearer <admin token>
```

Response: `{"deletedCount": 42}`

`GET /collections` includes `marketCrawlEnabled` for each owner. Owners can
always append market logs to their own guitars; the crawler account may only
write to guitars whose owner has opted in.

## Market crawler (GitHub Actions)

The [Market crawl](.github/workflows/crawl.yml) workflow searches Reverb and
uploads price observations to the API. eBay and Marktplaats are currently skipped
in GitHub Actions (`-skip-ebay` and `-skip-marktplaats`); Reverb alone is enough
for production crawls today.

| Trigger | When |
| ------- | ---- |
| Schedule | Every Sunday 06:00 UTC |
| Manual | Actions → Market crawl → Run workflow |
| Push | Changes under `cmd/crawler/` or `internal/marketcrawler/` |

Configure in the GitHub repo:

| Kind | Name | Value |
| ---- | ---- | ----- |
| Variable | `GUITARS_API_URL` | `https://guitars.brouwers.club` |
| Variable | `COGNITO_CLIENT_ID` | Cognito app client ID |
| Variable | `COGNITO_REGION` | `eu-central-1` (optional) |
| Secret | `COGNITO_CRAWLER_USERNAME` | `info@wbits.net` |
| Secret | `COGNITO_CRAWLER_PASSWORD` | Must match the Cognito user password exactly |
| Secret | `EBAY_CLIENT_ID` | eBay production app client ID (optional; local runs only until CI enables eBay) |
| Secret | `EBAY_CLIENT_SECRET` | eBay production app client secret (optional; local runs only until CI enables eBay) |

The crawler account (`info@wbits.net` by default) may append market logs to
guitars in collections where `marketCrawlEnabled` is true. When a listing
includes a photo, the crawler center-crops a 256×256 JPEG thumbnail, uploads
it to the CDN under `images/market-logs/`, and stores the URL as
`listingImageUrl` on the market log. It discovers collections via
`GET /collections` and skips owners with the flag disabled.
Set `MARKET_CRAWLER_EMAIL` and `MARKET_CRAWLER_USER_ID` on the API Lambda if you use
a different crawler account.

To read the current password from AWS Secrets Manager (if you use it as the
source of truth):

```bash
aws secretsmanager get-secret-value \
  --secret-id guitars/crawler-cognito-password \
  --query SecretString --output text --region eu-central-1
```

Copy that value into the `COGNITO_CRAWLER_PASSWORD` GitHub secret. When you
rotate the password, update **both** Cognito and the GitHub secret (and Secrets
Manager if you use it locally).

Local runs can load the password from Secrets Manager instead:

```bash
COGNITO_PASSWORD_SECRET_ID=guitars/crawler-cognito-password make crawl
```

## Deploying to AWS

```bash
S3_BUCKET=your-bucket STACK_NAME=guitars make deploy
```

The CloudFormation stack provisions the API Gateway, the Lambda, the DynamoDB
table and the Secrets Manager secret. After deployment, rotate the bearer
token by updating the secret value; the Lambda will pick up the new value
within five minutes.

# Guitars

A small AWS-Lambda-backed API that exposes my personal guitar collection.

## Endpoints

All endpoints are protected by a shared bearer token. Send it as
`Authorization: Bearer <token>` on every request.

| Method | Path             | Description                       |
| ------ | ---------------- | --------------------------------- |
| GET    | `/guitar`        | list every guitar in the collection |
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

## Market crawler (GitHub Actions)

The [Market crawl](.github/workflows/crawl.yml) workflow searches Reverb (and
optionally eBay/Marktplaats) and uploads price observations to the API.

| Trigger | When |
| ------- | ---- |
| Schedule | Every Sunday 06:00 UTC |
| Manual | Actions → Market crawl → Run workflow |
| Push | Changes under `cmd/crawler/` only |

Configure in the GitHub repo:

| Kind | Name | Value |
| ---- | ---- | ----- |
| Variable | `GUITARS_API_URL` | `https://guitars.brouwers.club` |
| Variable | `COGNITO_CLIENT_ID` | Cognito app client ID |
| Variable | `COGNITO_REGION` | `eu-central-1` (optional) |
| Secret | `COGNITO_CRAWLER_USERNAME` | `info@wbits.net` |
| Secret | `AWS_ACCESS_KEY_ID` | IAM user with `secretsmanager:GetSecretValue` on `guitars/crawler-cognito-password` |
| Secret | `AWS_SECRET_ACCESS_KEY` | Matching secret key |

The crawler password lives in **AWS Secrets Manager** as
`guitars/crawler-cognito-password` (single source of truth with Cognito). The
workflow reads it at runtime — you no longer need `COGNITO_CRAWLER_PASSWORD` in
GitHub.

To rotate the password:

```bash
NEW_PW='your-new-password'
aws cognito-idp admin-set-user-password \
  --user-pool-id eu-central-1_J8DZBZWRu \
  --username info@wbits.net \
  --password "$NEW_PW" --permanent --region eu-central-1
aws secretsmanager put-secret-value \
  --secret-id guitars/crawler-cognito-password \
  --secret-string "$NEW_PW" --region eu-central-1
```

Local runs can use the same secret:

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

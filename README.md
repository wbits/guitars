# guitars-webapp

A small React + TypeScript single-page app for browsing and managing a guitar
collection. It is a thin UI over the **GuitarCollection** HTTP API at
[`wbits/guitars`](https://github.com/wbits/guitars) — a Go AWS Lambda fronted
by API Gateway that exposes five endpoints under `/guitar` (list, get, create,
replace, delete) and authenticates every request with a bearer token.

The app is shipped as a static bundle (Vite build) and deployed to S3 +
CloudFront.

## Stack

- Vite + React 18 + TypeScript (strict)
- react-router-dom v6
- TanStack Query (React Query v5) for data fetching, caching, mutations
- react-hook-form + zod for form state and validation
- Tailwind CSS (no UI kit)
- Native `fetch` wrapped in a tiny typed `apiClient`
- Vitest + @testing-library/react

## Project layout

```
src/
├── api/         fetch wrapper + typed CRUD helpers and React Query hooks
├── components/  AuthGate, GuitarForm, MoneyInput, PictureGallery, ErrorBanner
├── domain/      zod schema + TS types mirroring the API contract
├── lib/         token resolution, money conversion + formatting
├── pages/       routes: GuitarList, GuitarNew, GuitarView, GuitarEdit, Settings
├── App.tsx      shell layout + nav
└── main.tsx     React Query provider + router + AuthGate wiring
```

## Routes

| Path                  | Purpose                                        |
| --------------------- | ---------------------------------------------- |
| `/`                   | Redirects to `/guitars`                        |
| `/guitars`            | Table of guitars, sorted by brand              |
| `/guitars/new`        | Create form                                    |
| `/guitars/:id`        | Detail view + picture gallery + delete modal   |
| `/guitars/:id/edit`   | Edit form (prefilled)                          |
| `/settings`           | Paste a runtime bearer token (sessionStorage)  |

All routes except `/settings` are wrapped in `<AuthGate>`, which blocks the UI
with a friendly message linking to `/settings` whenever no token is configured.

## Local development

Prerequisites: Node 20+, npm 10+.

```sh
cp .env.example .env.local   # then fill in the values
npm install
npm run dev                  # http://localhost:5173
```

Required environment variables (all `VITE_`-prefixed so they are inlined into
the bundle):

| Variable                      | Purpose                                                                                         |
| ----------------------------- | ----------------------------------------------------------------------------------------------- |
| `VITE_GUITARS_API_BASE_URL`   | Base URL of the GuitarCollection API (no trailing slash).                                       |
| `VITE_GUITARS_BEARER_TOKEN`   | Optional bearer token baked into the build. Anything set here is publicly visible — see below.  |

Useful scripts:

| Command            | Effect                                            |
| ------------------ | ------------------------------------------------- |
| `npm run dev`      | Vite dev server with HMR.                         |
| `npm run build`    | `tsc --noEmit` + `vite build` (writes to `dist`). |
| `npm run preview`  | Serves the production build locally.              |
| `npm test`         | Runs the Vitest suite once.                       |
| `npm run test:watch` | Vitest in watch mode.                           |
| `npm run lint`     | `tsc --noEmit`.                                   |

The `Makefile` mirrors the most common scripts (`install`, `dev`, `test`,
`lint`, `build`) and adds `deploy` / `invalidate` for the static-hosting flow.

## Security: bearer-token trade-off

The GuitarCollection API requires `Authorization: Bearer <token>` on every
request. Because this app is a static SPA — there is no server we control that
can hold a secret — we offer **two** ways to supply the token, both
implemented in `src/lib/token.ts`:

1. **Build-time env var** (`VITE_GUITARS_BEARER_TOKEN`).
   If set when `npm run build` runs, the value is inlined into the JavaScript
   bundle. **Anyone who downloads the bundle can read it.** This is fine for a
   personal deployment with a token that only you use, but it is not safe for
   shared deployments.

2. **Runtime override** (`/settings`).
   Pasting a token on the Settings page stores it in `sessionStorage` (not
   `localStorage`) under the key `guitars:bearerToken`. It lives only for the
   current browser tab and is **never** sent to any server other than the API.
   When both sources are present, the runtime value wins.

For shared deploys prefer leaving `VITE_GUITARS_BEARER_TOKEN` blank and asking
each user to paste their own token at `/settings`. A future v2 should retire
both options and use a proper auth flow (Amazon Cognito, OAuth, or similar)
where the static site only ever sees a short-lived, per-user credential.

## Building and deploying

The app is hosted as a static bundle behind CloudFront. The included
`template.yaml` is a SAM/CloudFormation template that provisions:

- An S3 bucket (private, with Origin Access Control).
- A CloudFront distribution.
- A CloudFront Function that rewrites SPA paths (anything without a file
  extension) to `/index.html`, so deep links survive a refresh.
- 403/404 → 200 + `/index.html` error mappings as a belt-and-braces SPA
  fallback.

Outputs include the bucket name and distribution ID.

One-time stack creation (example):

```sh
sam deploy \
  --template-file template.yaml \
  --stack-name guitars-webapp \
  --capabilities CAPABILITY_IAM \
  --resolve-s3 \
  --parameter-overrides BucketName=my-guitars-webapp-bucket
```

After the stack exists, build and upload:

```sh
export VITE_GUITARS_API_BASE_URL=https://your-api.execute-api.eu-west-1.amazonaws.com/Prod
export VITE_GUITARS_BEARER_TOKEN=     # leave blank for runtime-only tokens
export BUCKET=my-guitars-webapp-bucket
export DIST=E1ABCDEFGHIJKL

make build
make deploy
make invalidate
```

## Continuous deployment

`.github/workflows/deploy.yml` runs on every push to `master`:

1. `npm ci`
2. `npx vitest run`
3. `npm run build`
4. `aws s3 sync dist/ s3://$BUCKET --delete`
5. `aws cloudfront create-invalidation --paths "/*"`

AWS credentials are obtained via OIDC — **no static access keys are stored in
the repo**. You must configure:

| Kind              | Name                       | Purpose                                                                |
| ----------------- | -------------------------- | ---------------------------------------------------------------------- |
| Secret            | `AWS_DEPLOY_ROLE_ARN`      | IAM role the workflow assumes (must trust `token.actions.githubusercontent.com`). |
| Secret            | `GUITARS_BEARER_TOKEN`     | Exposed to the build as `VITE_GUITARS_BEARER_TOKEN`. Optional.         |
| Variable          | `GUITARS_API_BASE_URL`     | Exposed to the build as `VITE_GUITARS_API_BASE_URL`. Required.         |
| Variable          | `GUITARS_BUCKET`           | Name of the S3 bucket (from `template.yaml` outputs).                  |
| Variable          | `GUITARS_DISTRIBUTION_ID`  | CloudFront distribution ID (from `template.yaml` outputs).             |
| Variable (opt.)   | `AWS_REGION`               | Defaults to `eu-west-1` if unset.                                      |

## CORS dependency on the API side

The deployed site lives at a CloudFront domain and calls the API at a
different origin (`*.execute-api.<region>.amazonaws.com`). For the browser to
allow those requests the API must respond to preflight `OPTIONS` and include
the appropriate `Access-Control-Allow-*` headers. The current
[`wbits/guitars`](https://github.com/wbits/guitars) API does not configure
CORS, so until that follow-up is shipped the deployed UI will be blocked by
the browser when calling a remote API. Local development via `npm run dev` is
not affected if you call a local SAM-Local endpoint that returns CORS headers.

## Out of scope (v1)

- No authentication beyond the bearer token.
- No CORS work on the API itself (tracked as a follow-up).
- No image uploads — pictures are referenced by absolute URL only.
- No price-comparison / scraper integration.

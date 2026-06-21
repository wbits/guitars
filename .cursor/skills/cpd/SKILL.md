---
name: cpd
description: >-
  Commit, push, and deploy changed repos (guitars API and/or guitars-webapp).
  Use when the user says /cpd, cpd, commit push deploy, or asks to ship changes
  to production.
---

# Commit push deploy (`/cpd`)

Ship pending work: commit → push → deploy. Covers **both** repos when needed.

## Trigger phrases

- `/cpd`
- `cpd`
- `commit push deploy`
- `commit, push, and deploy`

## Repos

| Repo | Path (typical) | Deploy when |
|------|----------------|-------------|
| **guitars** (API) | workspace root or `../guitars` | Go/API/infra/MCP changes |
| **guitars-webapp** | `../guitars-webapp` (sibling) | React/UI changes |

If the sibling webapp path is missing, ask once for its location.

## Config

Load deploy env from **guitars** repo:

```
.agents/config/cpd.env
```

If missing, copy from [`.agents/config/cpd.env.example`](../../.agents/config/cpd.env.example), fill in values, and **do not commit** `cpd.env`.

## Workflow

Run repos **in parallel** where independent (status/diff/log). Commit/push/deploy **sequentially per repo**.

### 1. Detect scope

For each repo, run `git status` and `git diff`. Include a repo only if it has staged/unstaged/untracked changes to ship.

If **no changes** anywhere, say so and stop.

### 2. Pre-flight (per changed repo)

| Repo | Checks |
|------|--------|
| guitars | `make test` (skip only if user said so or change is docs-only) |
| guitars-webapp | `npm test` |

### 3. Commit (per changed repo)

Follow the user's **Git Safety Protocol**:

1. In parallel: `git status`, `git diff`, `git log -3 --oneline`
2. Stage relevant files — **never** secrets (`.env`, `cpd.env`, tokens, credentials)
3. Commit with HEREDOC message (1–2 sentences, focus on *why*)
4. `git status` to verify

Do **not** amend unless user rules allow.

### 4. Push (per changed repo)

```bash
git push -u origin HEAD
```

Use `required_permissions: ["all"]` when the sandbox blocks network/git write.

### 5. Deploy

Source config first:

```bash
set -a && source .agents/config/cpd.env && set +a
```

**API** (guitars repo — deploy if API/infra/MCP changed):

```bash
cd <guitars-root>
S3_BUCKET="$GUITARS_S3_BUCKET" STACK_NAME="$GUITARS_STACK_NAME" AWS_REGION="$GUITARS_AWS_REGION" make deploy
```

**Webapp** (guitars-webapp — deploy if webapp changed):

```bash
cd <guitars-webapp-root>
export VITE_GUITARS_API_BASE_URL VITE_COGNITO_REGION VITE_COGNITO_USER_POOL_ID VITE_COGNITO_CLIENT_ID
make build
BUCKET="$WEBAPP_BUCKET" make deploy
DIST="$WEBAPP_DIST" make invalidate
```

Skip deploy steps for repos that had **no** changes. If only one repo changed, only deploy that one.

### 6. Report

Reply with:

- Commits created (hash + message per repo)
- Push result
- Deploy result (or CI note if push-to-master auto-deploys webapp and user prefers that)
- Anything skipped and why

## Options (user may specify)

| Flag / phrase | Effect |
|---------------|--------|
| `api only` / `webapp only` | Limit to one repo |
| `no deploy` / `commit push` | Stop after push |
| `skip tests` | Skip pre-flight tests |

## Rules

- `/cpd` **is** explicit permission to commit, push, and deploy — unlike normal sessions.
- Never commit `cpd.env`, `.env.local`, or bearer tokens.
- Never force-push to `master`.
- If deploy fails, report the error; do not re-commit unless fixing deploy blockers.

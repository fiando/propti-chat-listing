# Deployment Guide

## Deployment model

This repository deploys to the existing production projects:

- Frontend: existing Vercel project for `https://propti.id`
- Backend: existing AWS SAM stack `propti-backend` in `ap-southeast-1`

Preferred deployment path for both is GitHub Actions. In normal operation, you deploy by pushing the relevant changes to `main`.

## Env-file workflow

Keep two layers of env files:

### 1. Committed templates

Use these as the source of truth for required variables:

- `frontend/.env.production.example`
- `frontend/.env.development.example`
- `backend/.env.production.example`
- `backend/.env.development.example`

### 2. Real local env files

Use these for real values when preparing local deploy work:

- `frontend/.env.production`
- `frontend/.env.development`
- `backend/.env.production`
- `backend/.env.development`

These local files are intentionally gitignored.

## Frontend deployment (Vercel)

### Production runtime variables

Use these in both the committed frontend production template and the real Vercel project environment:

| Variable | Purpose |
|--------|-------------|
| `NEXTAUTH_URL` | Production frontend URL for NextAuth (`https://propti.id`) |
| `NEXTAUTH_SECRET` | NextAuth signing secret |
| `GOOGLE_CLIENT_ID` | Google OAuth web client ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth web client secret |
| `NEXT_PUBLIC_API_URL` | Backend API URL (`https://api.propti.id/v1`) |
| `NEXT_PUBLIC_GOOGLE_MAPS_API_KEY` | Google Maps API key |

### GitHub production secrets for frontend deploys

| Secret | Purpose |
|--------|-------------|
| `NEXTAUTH_URL` | Passed to build and deploy steps |
| `NEXTAUTH_SECRET` | Passed to build and deploy steps |
| `GOOGLE_CLIENT_ID` | Passed to build and deploy steps |
| `GOOGLE_CLIENT_SECRET` | Passed to build and deploy steps |
| `NEXT_PUBLIC_API_URL` | Used for frontend production builds |
| `NEXT_PUBLIC_GOOGLE_MAPS_API_KEY` | Used for frontend production builds |
| `VERCEL_TOKEN` | Authenticates the GitHub Actions Vercel deploy |
| `VERCEL_ORG_ID` | Targets the existing Vercel organization |
| `VERCEL_PROJECT_ID` | Targets the existing Vercel project |

### Preferred deploy flow

```bash
gh run list --workflow deploy-frontend.yml --limit 5
gh secret list --env production
cd frontend && npm run test:deploy-config
cd frontend && npm run lint
cd frontend && npm run build
```

Then push your frontend changes to `main` and let the `Deploy Frontend` workflow publish them. Use `workflow_dispatch` only when you need to rerun the workflow without a new commit.

### Custom domain checks

- Production domain: `https://propti.id`
- Ensure `NEXTAUTH_URL` matches the production domain.
- Ensure Google OAuth authorized origins and callback URLs include the production domain.

## Backend deployment (AWS SAM)

### Production secret inputs

Use these in the committed backend production template and GitHub production secrets:

| Variable | Purpose |
|--------|-------------|
| `JWT_SECRET` | JWT signing secret |
| `OPENAI_API_KEY` | OpenAI API key for AI parsing |
| `GOOGLE_MAPS_API_KEY` | Google Maps API key |
| `DOKU_CLIENT_ID` | DOKU client ID |
| `DOKU_SECRET_KEY` | DOKU secret key |
| `AWS_ROLE_ARN` | GitHub Actions OIDC role ARN |

### Runtime values injected by SAM

These come from `backend/template.yaml` and do not need to be committed in local env templates unless a local tool specifically needs them:

- `DYNAMODB_LISTINGS_TABLE`
- `DYNAMODB_USERS_TABLE`
- `DYNAMODB_TRANSACTIONS_TABLE`
- `DYNAMODB_MODERATIONS_TABLE`
- `S3_MEDIA_BUCKET`
- `DOKU_ENV`
- `LOG_LEVEL`

### Preferred deploy flow

```bash
gh run list --workflow deploy-backend.yml --limit 5
cd backend && go test ./...
cd backend && sam build
```

Then push your backend changes to `main` and let the `Deploy Backend` workflow publish them. Use `workflow_dispatch` only when you need to rerun the workflow without a new commit.

### Important backend notes

- The production stack reuses previously stored secret parameter values when they are omitted from repeated `sam deploy` runs.
- `CAPABILITY_NAMED_IAM` is required because the SAM template creates a named IAM role.
- The GitHub Actions backend deploy uses `sam deploy --resolve-s3` so SAM can manage an artifact bucket for Lambda package uploads.
- The backend custom domain remains `https://api.propti.id`.
- The current SAM mapping keeps non-production stages on `DOKU_ENV=sandbox` and sets `production` to `DOKU_ENV=production`.

## Deployment troubleshooting

- If `gh secret list --env production` returns `no secrets found`, re-add the required production environment secrets before rerunning either deploy workflow.
- The latest verified frontend workflow failure in this repo stopped on `Missing required frontend secret: NEXTAUTH_URL`.
- The latest verified backend workflow failure in this repo stopped on `Credentials could not be loaded, please check your action inputs: Could not load credentials from any providers`, which is consistent with a missing `AWS_ROLE_ARN` production secret.

## AWS OIDC setup for GitHub Actions

Use `AWS_ROLE_ARN` for the GitHub Actions deploy workflow. If you ever need to recreate the role/provider, follow the existing OIDC setup process already used by the repository.

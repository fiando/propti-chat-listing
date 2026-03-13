# Deployment Guide

## Deployment model

This repository deploys to the existing production projects:

- Frontend: existing Vercel project for `https://propti.id`
- Backend: existing AWS SAM stack `propti-backend` in `ap-southeast-1`

Preferred deployment path for both is GitHub Actions. Direct CLI deploys are only appropriate when the corresponding project credentials and secrets already exist.

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

Then deploy through the GitHub Actions workflow `Deploy Frontend`.

### Direct CLI deploy

```bash
cd frontend
npm install
npm run build
vercel env ls production
vercel --prod
```

Only use direct CLI deploy after confirming that the existing Vercel project already contains the required production env values or you are explicitly passing them during deploy.

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
| `MIDTRANS_SERVER_KEY` | Midtrans server key |
| `AWS_ROLE_ARN` | GitHub Actions OIDC role ARN |

### Runtime values injected by SAM

These come from `backend/template.yaml` and do not need to be committed in local env templates unless a local tool specifically needs them:

- `DYNAMODB_LISTINGS_TABLE`
- `DYNAMODB_USERS_TABLE`
- `DYNAMODB_TRANSACTIONS_TABLE`
- `DYNAMODB_MODERATIONS_TABLE`
- `S3_MEDIA_BUCKET`
- `MIDTRANS_ENV`
- `LOG_LEVEL`

### Preferred deploy flow

```bash
gh run list --workflow deploy-backend.yml --limit 5
cd backend && go test ./...
cd backend && sam build
```

Then deploy through the GitHub Actions workflow `Deploy Backend`.

### Manual deploy from an AWS-authenticated machine

```bash
cd backend
go test ./...
sam build
sam deploy \
  --no-confirm-changeset \
  --no-fail-on-empty-changeset \
  --resolve-s3 \
  --stack-name propti-backend \
  --region ap-southeast-1 \
  --capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM \
  --parameter-overrides \
    Stage=production \
    ApiCustomDomainName=api.propti.id \
    ApiCustomDomainCertificateArn=arn:aws:acm:ap-southeast-1:260317865867:certificate/a6625ab7-0527-4dbf-aa1e-f22a39a33e98 \
    JWTSecret=your-jwt-secret \
    OpenAIAPIKey=your-openai-api-key \
    GoogleMapsAPIKey=your-google-maps-api-key \
    MidtransServerKey=your-midtrans-server-key
```

### Important backend notes

- The production stack reuses previously stored secret parameter values when they are omitted from repeated `sam deploy` runs.
- `CAPABILITY_NAMED_IAM` is required because the SAM template creates a named IAM role.
- The backend custom domain remains `https://api.propti.id`.
- The current SAM mapping sets `MIDTRANS_ENV=sandbox` even when `Stage=production`; treat that as current behavior unless you intentionally change the template.

## Deployment troubleshooting

- If `gh secret list --env production` returns `no secrets found`, re-add the required production environment secrets before rerunning either deploy workflow.
- The latest verified frontend workflow failure in this repo stopped on `Missing required frontend secret: NEXTAUTH_URL`.
- The latest verified backend workflow failure in this repo stopped on `Credentials could not be loaded, please check your action inputs: Could not load credentials from any providers`, which is consistent with a missing `AWS_ROLE_ARN` production secret.

## AWS OIDC setup for GitHub Actions

Use `AWS_ROLE_ARN` for the GitHub Actions deploy workflow. If you ever need to recreate the role/provider, follow the existing OIDC setup process already used by the repository.

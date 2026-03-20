# AGENTS.md

## Repeatable deployment notes

Use these steps when you need to redeploy production without rediscovering the setup.

### Deployment reminder

Production deployment now relies on GitHub Actions. The normal release path is to push the relevant changes to `main` and let the workflows deploy automatically.

### Shared env-file workflow

Use committed example files as the source of truth for required variables:

- `frontend/.env.production.example`
- `frontend/.env.development.example`
- `backend/.env.production.example`
- `backend/.env.development.example`

Use real local env files only for local secret handling and keep them untracked:

- `frontend/.env.production`
- `frontend/.env.development`
- `backend/.env.production`
- `backend/.env.development`

These local files are ignored by git so they can be used for later deploy work without committing secrets.

### Frontend (`frontend/`) -> Vercel

Preferred path: push frontend changes to `main` so the `Deploy Frontend` GitHub Actions workflow runs automatically. Use `workflow_dispatch` only for manual reruns.

Committed frontend templates should include:
- `NEXTAUTH_URL`
- `NEXTAUTH_SECRET`
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- `NEXT_PUBLIC_API_URL`
- `NEXT_PUBLIC_GOOGLE_MAPS_API_KEY`

Required GitHub production secrets:
- `NEXTAUTH_URL`
- `NEXTAUTH_SECRET`
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- `NEXT_PUBLIC_API_URL`
- `NEXT_PUBLIC_GOOGLE_MAPS_API_KEY`
- `VERCEL_TOKEN`
- `VERCEL_ORG_ID`
- `VERCEL_PROJECT_ID`

Production defaults in this repo:
- `NEXTAUTH_URL=https://propti.id`
- `NEXT_PUBLIC_API_URL=https://api.propti.id`

Helpful checks:

```bash
gh run list --workflow deploy-frontend.yml --limit 5
gh secret list --env production
cd frontend && npm run test:deploy-config
cd frontend && npm run lint
cd frontend && npm run build
cd frontend && vercel env ls production
```

If `gh secret list --env production` returns `no secrets found`, restore the frontend secrets above before rerunning the workflow. The latest failed frontend workflow in this repo stopped immediately on missing `NEXTAUTH_URL`.

Current linked Vercel project details in this repo:
- `frontend/.vercel/project.json`
- `projectName=frontend`
- `orgId=team_hcVAY7JajK1yJwfF1wZjForP`
- `projectId=prj_NWTrClnRGeb5iYaPuGY3mNwd6hTf`

The latest verified production deploy path is GitHub Actions on `main`, targeting `https://propti.id`.

### Backend (`backend/`) -> AWS SAM

Preferred path: push backend changes to `main` so the `Deploy Backend` GitHub Actions workflow runs automatically. Use `workflow_dispatch` only for manual reruns.

Committed backend templates should include:
- `JWT_SECRET`
- `OPENAI_API_KEY`
- `GOOGLE_MAPS_API_KEY`
- `DOKU_CLIENT_ID`
- `DOKU_SECRET_KEY`
- `AWS_ROLE_ARN`

Required GitHub production secrets:
- `AWS_ROLE_ARN`
- `JWT_SECRET`
- `OPENAI_API_KEY`
- `GOOGLE_MAPS_API_KEY`
- `DOKU_CLIENT_ID`
- `DOKU_SECRET_KEY`

SAM injects these runtime values during deploy, so they do not need to be authored in committed env templates unless local tooling specifically needs them:
- `DYNAMODB_LISTINGS_TABLE`
- `DYNAMODB_USERS_TABLE`
- `DYNAMODB_TRANSACTIONS_TABLE`
- `DYNAMODB_MODERATIONS_TABLE`
- `S3_MEDIA_BUCKET`
- `DOKU_ENV`
- `LOG_LEVEL`

Notes:
- The production stack already stores the secret parameters, so omitting them in repeated `sam deploy` runs reuses the previous values.
- `CAPABILITY_NAMED_IAM` is required because the SAM template creates a named IAM role.
- The current SAM mapping should keep non-production stages on `DOKU_ENV=sandbox` and set `production` to `DOKU_ENV=production`.

Helpful checks:

```bash
aws cloudformation describe-stacks --stack-name propti-backend --region ap-southeast-1
gh run list --workflow deploy-backend.yml --limit 5
cd backend && go test ./...
cd backend && sam build
```

If `gh secret list --env production` returns `no secrets found`, restore `AWS_ROLE_ARN` and the backend secrets above before rerunning the workflow. The latest failed backend workflow in this repo stopped at AWS credential setup.

The latest verified production deploy path is GitHub Actions on `main`, targeting `https://api.propti.id`.

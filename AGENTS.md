# AGENTS.md

## Repeatable deployment notes

Use these steps when you need to redeploy production without rediscovering the setup.

### Skills-first reminder

When a future agent needs to deploy or redeploy production, load these repo-local skills first:

- `.github/skills/production-deploy/SKILL.md`
- `.github/skills/deploy-frontend/SKILL.md`
- `.github/skills/deploy-backend/SKILL.md`

Use the combined skill for triage and end-to-end redeploys, then the per-surface skills for the exact frontend or backend procedure.

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

Preferred path: run the GitHub Actions workflow `Deploy Frontend`.

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

If `gh secret list --env production` returns `no secrets found`, restore the frontend secrets above before rerunning `Deploy Frontend`. The latest failed frontend workflow in this repo stopped immediately on missing `NEXTAUTH_URL`.

Direct CLI deploy is only safe after the auth and public runtime variables above exist either in the Vercel project env or are passed explicitly during deploy.

Manual CLI fallback that was verified from this repo on 2026-03-13:

```bash
cd frontend
vercel whoami
vercel env ls production
npm run test:deploy-config
npm run lint
npm run build
vercel deploy --prod --yes
vercel inspect <deployment-url>
```

Current linked Vercel project details in this repo:
- `frontend/.vercel/project.json`
- `projectName=frontend`
- `orgId=team_hcVAY7JajK1yJwfF1wZjForP`
- `projectId=prj_NWTrClnRGeb5iYaPuGY3mNwd6hTf`

The latest verified production redeploy from local CLI produced a Ready deployment aliased to `https://propti.id`.

### Backend (`backend/`) -> AWS SAM

Preferred path: run the GitHub Actions workflow `Deploy Backend`.

Committed backend templates should include:
- `JWT_SECRET`
- `OPENAI_API_KEY`
- `GOOGLE_MAPS_API_KEY`
- `MIDTRANS_SERVER_KEY`
- `AWS_ROLE_ARN`

Required GitHub production secrets:
- `AWS_ROLE_ARN`
- `JWT_SECRET`
- `OPENAI_API_KEY`
- `GOOGLE_MAPS_API_KEY`
- `MIDTRANS_SERVER_KEY`

SAM injects these runtime values during deploy, so they do not need to be authored in committed env templates unless local tooling specifically needs them:
- `DYNAMODB_LISTINGS_TABLE`
- `DYNAMODB_USERS_TABLE`
- `DYNAMODB_TRANSACTIONS_TABLE`
- `DYNAMODB_MODERATIONS_TABLE`
- `S3_MEDIA_BUCKET`
- `MIDTRANS_ENV`
- `LOG_LEVEL`

Manual deploy from an AWS-authenticated machine:

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
    ApiCustomDomainCertificateArn=arn:aws:acm:ap-southeast-1:260317865867:certificate/a6625ab7-0527-4dbf-aa1e-f22a39a33e98
```

Notes:
- The production stack already stores the secret parameters, so omitting them in repeated `sam deploy` runs reuses the previous values.
- `CAPABILITY_NAMED_IAM` is required because the SAM template creates a named IAM role.
- The current SAM mapping keeps `MIDTRANS_ENV=sandbox` even for `production`; document that behavior unless you intentionally change the template.

Helpful checks:

```bash
aws cloudformation describe-stacks --stack-name propti-backend --region ap-southeast-1
gh run list --workflow deploy-backend.yml --limit 5
cd backend && go test ./...
cd backend && sam build
```

If `gh secret list --env production` returns `no secrets found`, restore `AWS_ROLE_ARN` and the backend secrets above before rerunning `Deploy Backend`. The latest failed backend workflow in this repo stopped at AWS credential setup.

Manual CLI fallback that was verified from this repo on 2026-03-13:

```bash
aws sts get-caller-identity
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
    ApiCustomDomainCertificateArn=arn:aws:acm:ap-southeast-1:260317865867:certificate/a6625ab7-0527-4dbf-aa1e-f22a39a33e98
aws cloudformation describe-stacks --stack-name propti-backend --region ap-southeast-1
```

The latest verified local SAM redeploy completed successfully and preserved the custom API domain `https://api.propti.id`.

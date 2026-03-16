---
name: Production deploying
description: Coordinates production redeploys for this repo. Use when asked to deploy or redeploy latest changes, choose between GitHub Actions and local CLI fallback, verify prerequisites, and confirm frontend/backend production status.
---

## Goal

Redeploy the latest production changes for this repository without rediscovering the setup.

## Deployment surfaces

- Frontend: Next.js app in `frontend/`, deployed to Vercel, production domain `https://propti.id`
- Backend: Go/SAM app in `backend/`, deployed to AWS, production API domain `https://api.propti.id`

## Default path

Prefer GitHub Actions first:

- `Deploy Frontend` from `.github/workflows/deploy-frontend.yml`
- `Deploy Backend` from `.github/workflows/deploy-backend.yml`

Check first:

```bash
git --no-pager status --short
git rev-parse --abbrev-ref HEAD
gh workflow list
gh run list --limit 10
gh secret list --env production
```

## If GitHub Actions are blocked

Use local CLI fallback when GitHub production secrets are incomplete but the local machine is already authenticated.

Frontend fallback:

1. Load `deploy-frontend`
2. Verify local Vercel auth with `vercel whoami`
3. Validate and deploy from `frontend/`

Backend fallback:

1. Load `deploy-backend`
2. Verify local AWS auth with `aws sts get-caller-identity`
3. Validate and deploy from `backend/`

## Required prechecks

- Confirm repo state and current `main` commit
- Check recent failed runs before retrying, so you do not repeat a known secret/auth failure blindly
- Validate locally before any production CLI deploy:
  - Frontend: `npm run test:deploy-config && npm run lint && npm run build`
  - Backend: `go test ./... && sam build`

## Current known failure modes

- Frontend GitHub Action fails immediately if `VERCEL_TOKEN` is missing
- Backend GitHub Action fails at AWS credential setup if `AWS_ROLE_ARN` is missing or invalid
- If `gh secret list --env production` shows no secrets, restore production secrets before using the workflow path

## Verification

After deploying:

```bash
cd frontend && vercel ls frontend --prod --yes | head -n 20
cd frontend && vercel inspect <deployment-url>
aws cloudformation describe-stacks --stack-name propti-backend --region ap-southeast-1
```

Successful state:

- Vercel deployment is `Ready` and aliased to `https://propti.id`
- CloudFormation stack `propti-backend` finishes `UPDATE_COMPLETE`
- Backend custom domain remains `https://api.propti.id`

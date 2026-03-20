---
name: Production deploying
description: Use when asked how production deployment works in this repo, how to verify GitHub Actions deployment status, or how pushes to main trigger frontend and backend releases
---

## Goal

Explain and verify the production deployment flow for this repository.

## Deployment surfaces

- Frontend: Next.js app in `frontend/`, deployed to Vercel, production domain `https://propti.id`
- Backend: Go/SAM app in `backend/`, deployed to AWS, production API domain `https://api.propti.id`

## Default path

Production deployment relies on GitHub Actions:

- `Deploy Frontend` from `.github/workflows/deploy-frontend.yml`
- `Deploy Backend` from `.github/workflows/deploy-backend.yml`

If you push changes to `main`, GitHub Actions deploys the affected surface automatically:

- changes under `frontend/**` trigger the frontend deploy workflow
- changes under `backend/**` trigger the backend deploy workflow

`workflow_dispatch` exists for manual reruns, but the normal release path is still GitHub Actions on `main`.

Check first:

```bash
git --no-pager status --short
git rev-parse --abbrev-ref HEAD
gh workflow list
gh run list --limit 10
gh secret list --env production
```

## Required prechecks

- Confirm repo state and current `main` commit
- Check recent failed runs before retrying, so you do not repeat a known secret/auth failure blindly
- Validate locally before pushing when the changes are deployment-sensitive:
  - Frontend: `npm run test:deploy-config && npm run lint && npm run build`
  - Backend: `go test ./... && sam build`

## Current known failure modes

- Frontend GitHub Action fails immediately if `VERCEL_TOKEN` is missing
- Backend GitHub Action fails at AWS credential setup if `AWS_ROLE_ARN` is missing or invalid
- If `gh secret list --env production` shows no secrets, restore production secrets before using the workflow path

## Verification

After deploying:

```bash
gh run list --workflow deploy-frontend.yml --limit 5
gh run list --workflow deploy-backend.yml --limit 5
aws cloudformation describe-stacks --stack-name propti-backend --region ap-southeast-1
```

Successful state:

- Frontend workflow succeeds and the Vercel deployment is `Ready` at `https://propti.id`
- CloudFormation stack `propti-backend` finishes `UPDATE_COMPLETE`
- Backend custom domain remains `https://api.propti.id`

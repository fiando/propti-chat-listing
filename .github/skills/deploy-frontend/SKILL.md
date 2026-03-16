---
name: Frontend deploying
description: Deploys the frontend in this repo to Vercel production. Use when redeploying `frontend/`, checking required auth/runtime env, validating locally, or verifying the Vercel deployment and alias for `https://propti.id`.
---

## App

- Directory: `frontend/`
- Platform: Vercel
- Production domain: `https://propti.id`
- Linked project file: `frontend/.vercel/project.json`

Current linked identifiers:

- `projectName=frontend`
- `orgId=team_hcVAY7JajK1yJwfF1wZjForP`
- `projectId=prj_NWTrClnRGeb5iYaPuGY3mNwd6hTf`

## Preferred workflow path

```bash
gh run list --workflow deploy-frontend.yml --limit 5
gh secret list --env production
```

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

If `VERCEL_TOKEN` or project IDs are missing, the GitHub Actions path will fail. Use local CLI fallback instead if the machine is already authenticated with Vercel.

## Local validation

```bash
cd frontend
vercel whoami
vercel env ls production
npm run test:deploy-config
npm run lint
npm run build
```

Production defaults used by this repo:

- `NEXTAUTH_URL=https://propti.id`
- `NEXT_PUBLIC_API_URL=https://api.propti.id`

## Local CLI deploy

```bash
cd frontend
vercel deploy --prod --yes
```

Capture the returned deployment URL, then verify:

```bash
cd frontend
vercel inspect <deployment-url>
vercel ls frontend --prod --yes | head -n 20
```

Success criteria:

- Status is `Ready`
- Aliases include `https://propti.id`
- The deployment appears as the newest production deployment for `frontend`

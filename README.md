# Propti - Jual Beli Properti Semudah Chat WhatsApp

**Propti** is an Indonesian real estate platform that makes listing and finding properties as easy as sending a WhatsApp message. Users can paste their informal property listing text and AI automatically structures it into a professional listing.

## Features

- 🤖 **AI Text Parsing** – Paste WhatsApp-style listing text, AI extracts all property details
- 🏠 **Free Listing** – 1 listing/month with up to 3 photos (free tier)
- 💎 **Premium Features** – Unlimited listings, featured placement, promotion
- 📍 **Location Search** – Find properties by city, district, or nearby
- 💳 **Midtrans Payments** – Support for all Indonesian payment methods
- 🔒 **Google OAuth** – Simple sign-in with Google

## Business Model

| Feature | Free | Premium (Rp 49k/bln) |
|---------|------|----------------------|
| Listings/month | 1 | Unlimited |
| Photos per listing | 3 | Unlimited |
| Featured listing | – | Rp 50-100k / 7 hari |
| Promotion listing | – | Rp 25-50k / 7 hari |

## Architecture

```
Frontend (Vercel)          Backend (AWS Lambda)
Next.js 14              ←→  Go 1.24 + API Gateway
Tailwind + Shadcn/UI        DynamoDB + S3
NextAuth (Google OAuth)     OpenAI GPT-4 mini
                            Midtrans (payment)
                            Google Maps API
```

## Project Structure

```
propti/
├── backend/          # Go Lambda functions (AWS SAM)
│   ├── cmd/          # Lambda entry points
│   ├── internal/     # Models, services, handlers, repository
│   └── template.yaml # AWS SAM infrastructure
├── frontend/         # Next.js 14 application
│   ├── app/          # App Router pages
│   ├── components/   # React components
│   ├── hooks/        # Custom React hooks
│   └── lib/          # API client, utils
└── docs/             # Documentation
```

## Quick Start

See [docs/LOCAL_SETUP.md](docs/LOCAL_SETUP.md) for setup instructions.
See [docs/BRAND_GUIDELINES.md](docs/BRAND_GUIDELINES.md) for the Propti brand system.

### Backend
```bash
cd backend
make build
sam local start-api
```

### Frontend
```bash
cd frontend
cp .env.local.example .env.local
npm install
npm run dev
```

## Deployment

### Frontend (Vercel)

Preferred path: push frontend changes to `main`. GitHub Actions runs `Deploy Frontend` automatically for `frontend/**` changes.

Required production secrets:
- `NEXTAUTH_URL`
- `NEXTAUTH_SECRET`
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- `NEXT_PUBLIC_API_URL`
- `NEXT_PUBLIC_GOOGLE_MAPS_API_KEY`
- `VERCEL_TOKEN`
- `VERCEL_ORG_ID`
- `VERCEL_PROJECT_ID`

Env-file workflow:
- committed templates: `frontend/.env.production.example`, `frontend/.env.development.example`
- local ignored files: `frontend/.env.production`, `frontend/.env.development`

Helpful checks:

```bash
gh run list --workflow deploy-frontend.yml --limit 5
gh secret list --env production
cd frontend && npm run test:deploy-config
cd frontend && npm run lint
cd frontend && npm run build
```

If `gh secret list --env production` returns `no secrets found`, re-add the listed frontend production secrets before rerunning the workflow.

### Backend (AWS SAM)

Preferred path: push backend changes to `main`. GitHub Actions runs `Deploy Backend` automatically for `backend/**` changes.

Required production secrets:
- `AWS_ROLE_ARN`
- `JWT_SECRET`
- `OPENAI_API_KEY`
- `GOOGLE_MAPS_API_KEY`
- `DOKU_CLIENT_ID`
- `DOKU_SECRET_KEY`

Env-file workflow:
- committed templates: `backend/.env.production.example`, `backend/.env.development.example`
- local ignored files: `backend/.env.production`, `backend/.env.development`

Notes:
- The production stack reuses previously stored secret parameter values when they are omitted from repeated `sam deploy` runs.
- `CAPABILITY_NAMED_IAM` is required because the SAM template creates a named IAM role.
- The backend custom domain remains `https://api.propti.id`. If DNS needs to be re-pointed, use the stack output `ApiCustomDomainRegionalName`.
- GitHub Actions is the deployment mechanism for production; `workflow_dispatch` is available for manual reruns without needing to push another commit.

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for the full deployment guide.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Next.js 14, TypeScript, Tailwind CSS, Shadcn/UI |
| Backend | Go 1.24, AWS Lambda, API Gateway |
| Database | Amazon DynamoDB |
| Storage | Amazon S3 |
| AI | OpenAI GPT-4 mini |
| Auth | Google OAuth 2.0 + JWT |
| Payments | Midtrans Snap |
| Maps | Google Maps Platform |
| Hosting | Vercel (frontend) + AWS (backend) |

## License

MIT

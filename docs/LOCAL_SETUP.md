# Local Development Setup

## Prerequisites

- Go 1.24+
- Node.js 20+
- Docker (or Podman with `podman-docker`)

## 1. Clone & Install Dependencies

```bash
git clone https://github.com/fiando/propti-chat-listing.git
cd propti-chat-listing
cd frontend && npm install
cd ../backend && go mod download
cd ..
```

## 2. Environment Variables

Create local env files from the committed example templates:

```bash
cp frontend/.env.local.example frontend/.env.local
cp backend/.env.local.example backend/.env.local
```

Enable demo mode to bypass Google OAuth locally:

```bash
echo "NEXT_PUBLIC_DEMO_MODE=true" >> frontend/.env.local
```

Env file workflow:
- **Committed** examples (source of truth for required variables):
  - `frontend/.env.local.example`, `frontend/.env.development.example`, `frontend/.env.production.example`
  - `backend/.env.local.example`, `backend/.env.development.example`, `backend/.env.production.example`
- **Git-ignored** local files: `.env.local`, `.env.development`, `.env.production`

## 3. Start Local Infrastructure

```bash
docker compose up -d
```

This starts:
| Service | Port | Purpose |
|---|---|---|
| DynamoDB Local | 8000 | Local NoSQL database with persistent volume |
| MinIO | 9000 (API), 9001 (Console) | S3-compatible object storage |
| MinIO Setup | — | Creates `propti-media-development` bucket |

MinIO credentials: `minioadmin` / `minioadmin`

## 4. Seed Dummy Data

```bash
node scripts/seed-local.mjs
```

Populates DynamoDB Local with:
- Demo user (`demo@propti.app`)
- 8 sample property listings (sell/rent, various cities)
- 3 saved listings
- 3 CRM leads

## 5. Run the App

```bash
./scripts/dev-local.mjs
```

- **Frontend**: `http://localhost:3000`
- **Backend API**: `http://localhost:3001`

The script:
- Validates `frontend/.env.local` and `backend/.env.local` exist
- Bootstraps DynamoDB Local tables if they don't exist
- Starts the Go backend natively on port 3001 (no Docker/SAM needed)
- Builds and starts Next.js on port 3000
- Stops both processes together on `Ctrl+C`

For the backend URL, use `localhost` consistently in the browser and env files. Do not mix `localhost` and `127.0.0.1`, or NextAuth's state cookie validation will fail on the callback.

To reuse `backend/.env.development` instead of `backend/.env.local`:

```bash
./scripts/dev-local.mjs --backend-env-file backend/.env.development
```

## 6. Demo Mode

When `NEXT_PUBLIC_DEMO_MODE=true` in `frontend/.env.local`:

- The login page auto-authenticates as `demo@propti.app` without Google OAuth
- The backend accepts mock ID tokens prefixed with `mock-`
- All features are accessible without real Google credentials

## 7. Running Tests

```bash
# Backend tests
cd backend && go test ./...

# Frontend lint & build check
cd frontend && npm run lint

# Orchestrator tests
node --test scripts/dev-local.test.mjs
```

## 8. Google OAuth Setup (for real auth, skip if using demo mode)

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a project and enable the Google+ API
3. Create OAuth 2.0 credentials
4. Add `http://localhost:3000` to authorized origins
5. Add `http://localhost:3000/api/auth/callback/google` to redirect URIs
6. Copy Client ID and Secret to `frontend/.env.local`

## 9. API Endpoints

When running locally, use paths matching `backend/template.yaml` directly (no stage prefix):

```bash
curl http://localhost:3001/
curl http://localhost:3001/listings
curl -X POST http://localhost:3001/auth/google
```

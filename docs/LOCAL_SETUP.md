# Local Development Setup

## Prerequisites

- Go 1.24+
- Node.js 20+
- AWS CLI configured
- AWS SAM CLI
- Docker (for SAM local)

## 1. Clone & Configure

```bash
git clone https://github.com/fiando/propti.git
cd propti
```

## 2. Backend Setup

### Environment Variables

Create `backend/.env.local` from the committed example:

```bash
cp backend/.env.local.example backend/.env.local
```

Required local defaults:
```
JWT_SECRET=your-jwt-secret-min-32-chars
OPENAI_API_KEY=sk-...
AWS_REGION=ap-southeast-1
DYNAMODB_ENDPOINT_URL=http://localhost:8000
GOOGLE_MAPS_API_KEY=AIza...
PUBLIC_API_BASE_URL=http://localhost:3001
```

The API will be available at `http://localhost:3001`.

### Running Tests

```bash
cd backend
go test ./...
```

## 3. Frontend Setup

### Environment Variables

```bash
cd frontend
cp .env.local.example .env.local
```

Edit `frontend/.env.local`:
```
NEXTAUTH_URL=http://localhost:3000
NEXTAUTH_SECRET=your-nextauth-secret
GOOGLE_CLIENT_ID=xxx.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret
NEXT_PUBLIC_API_URL=http://localhost:3001
NEXT_PUBLIC_GOOGLE_MAPS_API_KEY=AIza...
```

## 4. Install Dependencies

```bash
cd frontend && npm install
cd ../backend && go mod download
```

## 5. Run the Local Stack

If you use Podman instead of Docker, start the compatibility service first:

```bash
podman system service --time=0 tcp:127.0.0.1:2375 &
export DOCKER_HOST=tcp://127.0.0.1:2375
```

```bash
./scripts/dev-local.mjs
```

The frontend will be available at `http://localhost:3000` and the API will be available at `http://localhost:3001`.

For Google sign-in, use `localhost` consistently in the browser and env files. Do not mix `localhost` and `127.0.0.1`, or NextAuth's state cookie validation will fail on the callback.

The script:
- validates `frontend/.env.local` and `backend/.env.local`
- bootstraps local DynamoDB tables when `DYNAMODB_ENDPOINT_URL` points to `localhost` or `127.0.0.1`
- builds the backend Lambda binaries with `make build`
- starts AWS SAM on port `3001`
- starts Next.js on port `3000`
- stops both processes together on `Ctrl+C`

If you want to reuse `backend/.env.development` instead of `backend/.env.local`, run:

```bash
./scripts/dev-local.mjs --backend-env-file backend/.env.development
```

The launcher will convert the dotenv file into the JSON override format that your local SAM CLI expects.

### Calling the local API

Use the paths defined in `backend/template.yaml` directly. With `sam local start-api`, local routes do **not** include a deployed stage prefix such as `/dev`.

Examples:

```bash
curl http://127.0.0.1:3001/
curl http://127.0.0.1:3001/listings
curl -X POST http://127.0.0.1:3001/auth/google
```

If you call the wrong path or HTTP method, SAM returns:

```json
{"message":"Missing Authentication Token"}
```

## 6. AWS Services Setup (Local Dev)

### DynamoDB Local

```bash
docker run -p 8000:8000 amazon/dynamodb-local
```

Once DynamoDB Local is running, `./scripts/dev-local.mjs` creates the required local tables automatically.

### S3 Local (MinIO)

```bash
docker run -p 9000:9000 minio/minio server /data
```

## 7. Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable Google+ API
4. Create OAuth 2.0 credentials
5. Add `http://localhost:3000` to authorized origins
6. Add `http://localhost:3000/api/auth/callback/google` to redirect URIs
7. Copy Client ID and Client Secret to `frontend/.env.local`

## 8. Midtrans Sandbox Setup

1. Create account at [sandbox.midtrans.com](https://sandbox.midtrans.com)
2. Get your Server Key from Settings > Access Keys
3. Set `MIDTRANS_IS_PRODUCTION=false` for sandbox

## 9. OpenAI Setup

1. Get API key from [platform.openai.com](https://platform.openai.com)
2. Add to `OPENAI_API_KEY` in backend env

## Development Workflow

```bash
# One terminal for both services
./scripts/dev-local.mjs
```

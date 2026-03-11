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

Create `backend/.env`:
```
JWT_SECRET=your-jwt-secret-min-32-chars
OPENAI_API_KEY=sk-...
GOOGLE_CLIENT_ID=xxx.apps.googleusercontent.com
MIDTRANS_SERVER_KEY=SB-Mid-server-xxx
MIDTRANS_IS_PRODUCTION=false
GOOGLE_MAPS_API_KEY=AIza...
DYNAMODB_LISTINGS_TABLE=propti-listings
DYNAMODB_USERS_TABLE=propti-users
DYNAMODB_TRANSACTIONS_TABLE=propti-transactions
DYNAMODB_MODERATIONS_TABLE=propti-moderations
S3_BUCKET=propti-media-dev
AWS_REGION=ap-southeast-1
```

### Build & Run Locally

```bash
cd backend

# Install dependencies
go mod download

# Build all Lambda functions
make build

# Start local API (requires Docker)
sam local start-api --env-vars .env
```

The API will be available at `http://localhost:3000`.

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

Edit `.env.local`:
```
NEXTAUTH_URL=http://localhost:3000
NEXTAUTH_SECRET=your-nextauth-secret
GOOGLE_CLIENT_ID=xxx.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret
NEXT_PUBLIC_API_URL=http://localhost:3000
NEXT_PUBLIC_GOOGLE_MAPS_API_KEY=AIza...
```

### Install & Run

```bash
cd frontend
npm install
npm run dev
```

The frontend will be available at `http://localhost:3000`.

## 4. AWS Services Setup (Local Dev)

### DynamoDB Local

```bash
docker run -p 8000:8000 amazon/dynamodb-local

# Create tables
aws dynamodb create-table \
  --table-name propti-listings \
  --attribute-definitions \
    AttributeName=PK,AttributeType=S \
    AttributeName=SK,AttributeType=S \
  --key-schema \
    AttributeName=PK,KeyType=HASH \
    AttributeName=SK,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST \
  --endpoint-url http://localhost:8000
```

### S3 Local (MinIO)

```bash
docker run -p 9000:9000 minio/minio server /data
```

## 5. Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable Google+ API
4. Create OAuth 2.0 credentials
5. Add `http://localhost:3000` to authorized origins
6. Add `http://localhost:3000/api/auth/callback/google` to redirect URIs
7. Copy Client ID and Client Secret to your env files

## 6. Midtrans Sandbox Setup

1. Create account at [sandbox.midtrans.com](https://sandbox.midtrans.com)
2. Get your Server Key from Settings > Access Keys
3. Set `MIDTRANS_IS_PRODUCTION=false` for sandbox

## 7. OpenAI Setup

1. Get API key from [platform.openai.com](https://platform.openai.com)
2. Add to `OPENAI_API_KEY` in backend env

## Development Workflow

```bash
# Terminal 1: Backend
cd backend && sam local start-api

# Terminal 2: Frontend  
cd frontend && npm run dev
```

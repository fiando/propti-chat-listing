# Deployment Guide

## Backend Deployment (AWS Lambda)

### Prerequisites
- AWS account with sufficient permissions
- AWS CLI configured
- SAM CLI installed
- GitHub repository secrets configured

### GitHub Secrets Required

| Secret | Description |
|--------|-------------|
| `AWS_ROLE_ARN` | IAM role ARN for GitHub Actions OIDC |
| `JWT_SECRET` | JWT signing secret (min 32 chars) |
| `OPENAI_API_KEY` | OpenAI API key |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID |
| `MIDTRANS_SERVER_KEY` | Midtrans production server key |
| `GOOGLE_MAPS_API_KEY` | Google Maps API key |

### Manual Deployment

```bash
cd backend

# Build
make build

# Deploy (first time - guided)
sam deploy --guided

# Deploy (subsequent)
sam deploy \
  --stack-name propti-backend \
  --region ap-southeast-1 \
  --parameter-overrides \
    Environment=production \
    JWTSecret=your-secret \
    OpenAIAPIKey=sk-... \
    GoogleClientID=xxx.apps.googleusercontent.com \
    MidtransServerKey=Mid-server-xxx \
    GoogleMapsAPIKey=AIza...
```

### AWS Resources Created

- **API Gateway**: REST API with CORS
- **Lambda Functions**: auth, listings, ai
- **DynamoDB Tables**: propti-listings, propti-users, propti-transactions, propti-moderations
- **S3 Bucket**: propti-media (with lifecycle rules)
- **IAM Roles**: least-privilege per function

### Post-Deployment

After deployment, note the API Gateway URL from the SAM outputs:
```
API Endpoint: https://xxxxxxxx.execute-api.ap-southeast-1.amazonaws.com/prod
```

Use this URL as `NEXT_PUBLIC_API_URL` in your frontend.

## Frontend Deployment (Vercel)

### Prerequisites
- Vercel account
- GitHub repository connected to Vercel

### GitHub Secrets Required

| Secret | Description |
|--------|-------------|
| `VERCEL_TOKEN` | Vercel API token |
| `VERCEL_ORG_ID` | Vercel organization ID |
| `VERCEL_PROJECT_ID` | Vercel project ID |
| `NEXTAUTH_URL` | Production frontend URL for NextAuth |
| `NEXTAUTH_SECRET` | NextAuth signing secret used by the session endpoint |
| `GOOGLE_CLIENT_ID` | Google OAuth web client ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth web client secret |
| `NEXT_PUBLIC_API_URL` | Backend API URL |
| `NEXT_PUBLIC_GOOGLE_MAPS_API_KEY` | Google Maps API key |

### Environment Variables in Vercel

Set these in Vercel project settings:
```
NEXTAUTH_URL=https://propti.id
NEXTAUTH_SECRET=your-nextauth-secret
GOOGLE_CLIENT_ID=xxx.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret
NEXT_PUBLIC_API_URL=https://api.propti.id/v1
NEXT_PUBLIC_GOOGLE_MAPS_API_KEY=AIza...
```

Also set the same auth values as GitHub Actions production secrets. The deploy workflow now validates and injects:

```
NEXTAUTH_URL
NEXTAUTH_SECRET
GOOGLE_CLIENT_ID
GOOGLE_CLIENT_SECRET
```

### Manual Deployment

```bash
cd frontend
npm install
npm run build
vercel --prod
```

### Custom Domain

1. Add domain in Vercel project settings
2. Update DNS records as instructed by Vercel
3. Update `NEXTAUTH_URL` to your custom domain
4. Update Google OAuth authorized origins

## AWS OIDC Setup for GitHub Actions

```bash
# Create OIDC provider for GitHub
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1

# Create IAM role with trust policy
# Trust policy in trust-policy.json:
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Federated": "arn:aws:iam::ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
    },
    "Action": "sts:AssumeRoleWithWebIdentity",
    "Condition": {
      "StringEquals": {
        "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
      },
      "StringLike": {
        "token.actions.githubusercontent.com:sub": "repo:fiando/propti:*"
      }
    }
  }]
}

aws iam create-role \
  --role-name propti-github-actions \
  --assume-role-policy-document file://trust-policy.json

# Attach policies
aws iam attach-role-policy \
  --role-name propti-github-actions \
  --policy-arn arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess

aws iam attach-role-policy \
  --role-name propti-github-actions \
  --policy-arn arn:aws:iam::aws:policy/AmazonS3FullAccess

aws iam attach-role-policy \
  --role-name propti-github-actions \
  --policy-arn arn:aws:iam::aws:policy/AWSLambda_FullAccess

aws iam attach-role-policy \
  --role-name propti-github-actions \
  --policy-arn arn:aws:iam::aws:policy/AmazonAPIGatewayAdministrator

aws iam attach-role-policy \
  --role-name propti-github-actions \
  --policy-arn arn:aws:iam::aws:policy/AWSCloudFormationFullAccess

aws iam attach-role-policy \
  --role-name propti-github-actions \
  --policy-arn arn:aws:iam::aws:policy/IAMFullAccess
```

## Cost Estimation (3 months, ~3000 users)

| Service | Cost |
|---------|------|
| AWS Lambda | Free tier (1M calls/month) |
| DynamoDB | Free tier (25 GB) |
| S3 | ~$0.40/month (18 GB) |
| API Gateway | Free tier (1M requests) |
| Vercel | Free tier |
| OpenAI API | ~Rp 600k/month (3000 listings × Rp 200) |
| Google Maps | Free tier (~5k requests) |
| Midtrans | 2.9% + Rp 5k per transaction (revenue share) |
| **Total** | **~Rp 600k/month** |

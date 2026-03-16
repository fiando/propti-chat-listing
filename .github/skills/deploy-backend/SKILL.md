---
name: Backend deploying
description: Deploys the backend in this repo to AWS production with SAM. Use when redeploying `backend/`, checking AWS auth and production parameters, validating locally, or verifying the `propti-backend` stack and `https://api.propti.id`.
---

## App

- Directory: `backend/`
- Platform: AWS SAM / CloudFormation
- Stack: `propti-backend`
- Region: `ap-southeast-1`
- Production API domain: `https://api.propti.id`

## Preferred workflow path``
``
```bash
gh run list --workflow deploy-backend.yml --limit 5
gh secret list --env production
```

Required GitHub production secrets:

- `AWS_ROLE_ARN`
- `JWT_SECRET`
- `OPENAI_API_KEY`
- `GOOGLE_MAPS_API_KEY`
- `MIDTRANS_SERVER_KEY`

If `AWS_ROLE_ARN` is missing or invalid, the GitHub Action fails during `configure-aws-credentials`.

## Local validation

```bash
aws sts get-caller-identity
cd backend
go test ./...
sam build
```

## Local SAM deploy

```bash
cd backend
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

- The production stack already stores secret parameter values, so repeated deploys can omit them and reuse the previous values
- `CAPABILITY_NAMED_IAM` is required because the SAM template creates a named IAM role
- Current production mapping still keeps `MIDTRANS_ENV=sandbox`

## Verification

```bash
aws cloudformation describe-stacks --stack-name propti-backend --region ap-southeast-1
```

Success criteria:

- Stack reaches `UPDATE_COMPLETE`
- Output `ApiCustomDomainUrl` remains `https://api.propti.id`
- Lambda/API resources update without replacement

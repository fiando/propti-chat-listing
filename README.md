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

- **Backend**: AWS Lambda via SAM (`sam deploy`)
- **Frontend**: Vercel (auto-deploy on push to main)
- **Backend custom domain**: `api.propti.id` can be managed from `backend/template.yaml` using API Gateway custom domain resources and an ACM certificate ARN in `ap-southeast-1`. After deploy, point DNS to the stack output `ApiCustomDomainRegionalName`.

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

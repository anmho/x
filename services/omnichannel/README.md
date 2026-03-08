# Omni-Channel Notification System (MVP)

A production-ready omni-channel notification system built with Golang, Next.js, Temporal, and Supabase. Supports email via SendGrid plus test channels (`sms`, `push`, `webhook`) and optional local relay channels (`app`, `imessage`) for emulator workflows.

## Features

- ✅ **Email Notifications** via SendGrid
- ✅ **App/Emulator Relay Channel** via `OMNICHANNEL_APP_RELAY_URL` (optional)
- ✅ **Template Management** with variable substitution
- ✅ **Temporal Workflows** for reliable delivery and retries
- ✅ **REST API** with API key authentication
- ✅ **Next.js Admin UI** for managing notifications and templates
- ✅ **Supabase PostgreSQL** for data storage
- ✅ **Delivery Tracking** with status monitoring

## Architecture

### Backend (Golang)
- **Chi Router** for HTTP routing
- **Temporal** for workflow orchestration and retries
- **Supabase PostgreSQL** for database
- **SendGrid** for email delivery
- **API Key Authentication** via X-API-Key header

### Frontend (Next.js)
- **TypeScript** for type safety
- **Tailwind CSS** for styling
- **Axios** for API calls
- **Lucide React** for icons

## Prerequisites

- Docker + Docker Compose
- Go 1.22+ and Node.js 20+ (optional, only for non-Docker local dev)

## Quick Start (Docker Compose)

From `omnichannel/`, boot the full stack with one command:

```bash
./scripts/dev-up.sh
```

This starts:

- PostgreSQL on `localhost:5432` (with schema + seed data loaded automatically)
- Temporal on `localhost:7233` and Temporal UI on `localhost:8233`
- API on `localhost:8080`
- Worker connected to Temporal
- Frontend on `localhost:3000`

Useful commands:

```bash
./scripts/dev-ps.sh
./scripts/dev-logs.sh
./scripts/dev-down.sh
```

If a host port is already in use, override it inline:

```bash
API_PORT=18080 FRONTEND_PORT=3000 TEMPORAL_UI_PORT=8234 ./scripts/dev-up.sh
```

The main helper accepts subcommands too:

```bash
./scripts/stack.sh up
./scripts/stack.sh logs api
./scripts/stack.sh down
```

## ConnectRPC SDK Publishing

Proto and SDK lifecycle now follows a publish-first flow:

- Proto module: `backend/proto` (`buf.build/anmhela/omnichannel`)
- SDKs are published from BSR and consumed as versioned dependencies
- Client SDKs should not be generated and manually versioned inside app repos

Commands:

```bash
# Validate proto changes
./scripts/sdk.sh lint

# Push module to BSR
./scripts/sdk.sh push

# Publish ES + Go ConnectRPC SDK versions
SDK_VERSION=v1.0.0 ./scripts/sdk.sh publish-all

# Local server-side stub generation (backend runtime only)
./scripts/sdk.sh generate-server
```

See [backend/proto/README.md](backend/proto/README.md) for details.
CI workflow: `.github/workflows/publish-connectrpc-sdks.yml` (requires `BUF_TOKEN` secret).

## Manual Start (Without Docker Compose)

You can still run each service directly if needed (Supabase + Temporal + backend + frontend), but Docker Compose is now the default local workflow.

## Usage

### Access the Cloud Console

1. Open http://localhost:3000/deployments
2. Use the dedicated Notifications service page to:
   - Manage registered apps
   - Create/revoke service API keys
   - Test all notification channels (email/sms/push/webhook/app/imessage)
   - View Temporal workflow IDs and run states

### App/Emulator Relay Setup (Optional)

To deliver `app` and `push` channel payloads to a local emulator inbox:

```bash
export OMNICHANNEL_APP_RELAY_URL=http://127.0.0.1:3000/api/emulator/messages
```

Then restart the worker. The Omnichannel UI includes an App Emulator page at `/omnichannel/emulator` that reads this inbox.

### Cron Job Auto-Run (UI Runtime)

Cron jobs created in Omnichannel are persisted in local browser storage and auto-evaluated every 30 seconds while any `/omnichannel/*` page is open. Matching schedules dispatch linked campaigns automatically.

### API Authentication

The API supports either:
- `Authorization: Bearer <oauth-access-token>` (recommended)
- `X-API-Key: <service-api-key>` (for service consumers)

Service API keys are scoped. The seeded development key includes:
- `notifications:*`
- `templates:*`

Example with API key:

```bash
curl -H "X-API-Key: test-api-key-123" http://localhost:8080/api/v1/notifications
```

Example with OAuth bearer token:

```bash
curl -H "Authorization: Bearer <access_token>" http://localhost:8080/api/v1/notifications
```

The default development API key is `test-api-key-123` (hashed in the database seed).

### Send a Notification via API

```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-123" \
  -d '{
    "recipient_email": "user@example.com",
    "subject": "Test Notification",
    "body": "<p>Hello from the Omnichannel system!</p>"
  }'
```

### Create a Template

```bash
curl -X POST http://localhost:8080/api/v1/templates \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-123" \
  -d '{
    "name": "Welcome Email",
    "description": "Sent to new users",
    "subject": "Welcome {{user_name}}!",
    "body": "<p>Hi {{user_name}}, welcome to {{app_name}}!</p>",
    "variables": ["user_name", "app_name"]
  }'
```

### Use a Template

```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-123" \
  -d '{
    "recipient_email": "user@example.com",
    "template_id": "template-uuid-here",
    "variables": {
      "user_name": "John",
      "app_name": "MyApp"
    }
  }'
```

## Project Structure

```
omnichannel/
├── backend/
│   ├── cmd/
│   │   ├── api/            # REST API server
│   │   └── worker/         # Temporal worker
│   ├── internal/
│   │   ├── api/            # HTTP handlers, middleware, routes
│   │   ├── config/         # Configuration
│   │   ├── database/       # Database connection
│   │   ├── domain/         # Domain models
│   │   ├── provider/       # External services (SendGrid)
│   │   ├── repository/     # Data access layer
│   │   ├── temporal/       # Workflows and activities
│   │   └── ...
│   └── migrations/         # Database migrations
├── frontend/
│   ├── app/
│   │   ├── page.tsx                    # Home entry page
│   │   ├── deployments/                # Dedicated cloud console deployment page
│   │   ├── notifications/create/       # Send notification
│   │   └── templates/                  # Template management
│   └── lib/
│       └── api.ts                      # API client
├── infra/docker/initdb/                # DB bootstrap SQL (schema + auth + seed)
├── docker-compose.yml                  # One-command local stack
└── scripts/
    ├── stack.sh                        # Compose helper
    ├── sdk.sh                          # Buf lint/push/publish/generate helper
    ├── dev-up.sh                       # Start all services
    ├── dev-down.sh                     # Stop all services
    ├── dev-logs.sh                     # Stream logs
    └── dev-ps.sh                       # Service status
└── supabase/
    ├── migrations/         # Database schema
    └── seed.sql           # Sample data
```

## API Endpoints

### Notifications
- `GET /api/v1/notifications` - List all notifications
- `GET /api/v1/notifications/:id` - Get notification details
- `POST /api/v1/notifications` - Create notification
- `GET /api/v1/notifications/:id/status` - Get notification status

### Templates
- `GET /api/v1/templates` - List all templates
- `GET /api/v1/templates/:id` - Get template details
- `POST /api/v1/templates` - Create template
- `PUT /api/v1/templates/:id` - Update template
- `DELETE /api/v1/templates/:id` - Delete template

### Health
- `GET /api/v1/health` - Health check

### ConnectRPC (TemporalService)
- `POST /temporal.v1.TemporalService/StartNotification`
- `POST /temporal.v1.TemporalService/GetNotificationStatus`

ConnectRPC uses the same auth model and scopes as REST (`Authorization: Bearer` or `X-API-Key`).

## Database Schema

### Tables
- **users** - User records
- **api_keys** - API key authentication
- **templates** - Email templates
- **notifications** - Notification records
- **notification_attempts** - Delivery attempt history

## Temporal Workflows

### SendNotificationWorkflow
1. Renders template with variables
2. Sends email via SendGrid
3. Records delivery status
4. Retries on failure (exponential backoff)

## Testing Email Sending

Since we're using SendGrid:

1. **Configure SendGrid:**
   - Get API key from https://sendgrid.com
   - Add to `backend/.env`
   - Verify sender email in SendGrid

2. **Test Notification:**
   - Use the UI to send a test email
   - Check Temporal UI for workflow execution
   - Check SendGrid dashboard for delivery status

## Development Notes

### Hot Reload

Backend automatically rebuilds on file changes if using tools like `air`:

```bash
go install github.com/air-verse/air@latest
cd backend && air
```

Frontend has built-in hot reload with Next.js dev server.

### Viewing Temporal Workflows

1. Open http://localhost:8233
2. Navigate to "Workflows"
3. Click on any workflow to see execution history and status

### Database Management

```bash
# View database in browser
supabase db studio

# Run migrations
supabase db reset

# Create new migration
supabase migration new <name>
```

## Next Steps (Future Enhancements)

### Phase 2: SMS & Push
- ✨ Loop Message SMS integration
- ✨ Firebase Cloud Messaging (FCM) for push notifications
- ✨ Device registration for push

### Phase 3: Advanced Features
- ✨ Batch/bulk notifications
- ✨ Scheduled notifications
- ✨ User preference management
- ✨ Analytics dashboard
- ✨ A/B testing
- ✨ Rate limiting

### Phase 4: Production
- ✨ Comprehensive testing
- ✨ Monitoring and alerting
- ✨ Performance optimization
- ✨ Production deployment guides

## Troubleshooting

### Email not sending

1. Check SendGrid API key is valid
2. Verify sender email in SendGrid
3. Check Temporal worker logs for errors
4. View workflow execution in Temporal UI

### Database connection failed

1. Ensure Supabase is running: `supabase status`
2. Check DATABASE_URL in `.env`
3. Try: `supabase stop && supabase start`

### Temporal worker not connecting

1. Ensure Temporal is running: `temporal server start-dev`
2. Check TEMPORAL_HOST in `.env`
3. Verify worker is running

## License

MIT

## Contributing

This is an MVP. Contributions welcome for expanding to SMS, push, and additional features!

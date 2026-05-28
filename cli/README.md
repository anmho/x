# X CLI

Rust CLI for omnichannel notifications.

## Commands

Send a notification:

```bash
cargo run -- notifications send \
  --channel sms \
  --to +15555550123 \
  --subject "Test SMS" \
  --body "hello from cli"
```

List recent notifications:

```bash
cargo run -- notifications list --limit 20
```

Fetch workflow status:

```bash
cargo run -- notifications status --id <notification-id>
```

## Config

Environment variables:

```bash
export OMNICHANNEL_API_URL="http://localhost:8080/api/v1"
export OMNICHANNEL_API_KEY="test-api-key-123"
```

Supported channels:

- `email`
- `sms`
- `push`
- `webhook`

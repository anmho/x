# @x/sdk-omnichannel

ConnectRPC TypeScript client for the omnichannel Temporal service. Generated locally from proto for fast iteration during development.

## Local generation

```bash
# From repo root
nx run sdk:generate-es
# or
npm run sdk:generate-es
```

Generated output lives in `src/gen/`. Do not edit generated files manually.

## Usage

```ts
import { createClient } from "@connectrpc/connect-web";
import { TemporalService } from "@x/sdk-omnichannel";

const client = createClient(TemporalService, transport);
const res = await client.startNotification({ recipientEmail: "...", subject: "...", body: "..." });
```

## Dependencies

- `@bufbuild/protobuf` — Protobuf runtime
- `@connectrpc/connect` — Connect client primitives
- `@connectrpc/connect-web` — Browser transport

This is a [Next.js](https://nextjs.org) frontend for the Omnichannel + Cloud Console surfaces.

## Getting Started

First, run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

Key routes:

- `/` home entry
- `/omnichannel` omnichannel control surface
- `/omnichannel/test` channel testing + live previews
- `/omnichannel/emulator` app/emulator inbox viewer
- `/deployments` dedicated Notifications deployment console (apps, API keys, channel tester, Temporal workflows)
- `/notifications/create` send-notification form
- `/templates` template management

App emulator relay:

- When backend worker env var `OMNICHANNEL_APP_RELAY_URL` points to `http://127.0.0.1:3000/api/emulator/messages`, `app`, `imessage`, and `push` channel sends are POSTed to the local emulator inbox.

Auth behavior:

- Console auth is disabled by default in local dev (`NODE_ENV != production`).
- To force-enable auth in dev: `CONSOLE_AUTH_DISABLED=0`.
- To force-disable auth in any env: `CONSOLE_AUTH_DISABLED=1`.

## Trading Bots and XAPI

The bots APIs support two execution sources:

- `mock` mode (default): local in-memory runs for development
- `xapi-live` mode: enabled when live XAPI credentials are set

Environment variables:

```bash
XAPI_BASE_URL=https://<your-live-xapi-host>
XAPI_API_KEY=<live-xapi-api-key>
```

Server routes used by the bots UI:

- `GET /api/trading-bots/runs`
- `POST /api/trading-bots/runs`
- `POST /api/trading-bots/runs/:runId/pause`

You can start editing the page by modifying `app/page.tsx`. The page auto-updates as you edit the file.

This project uses [`next/font`](https://nextjs.org/docs/app/building-your-application/optimizing/fonts) to automatically optimize and load [Geist](https://vercel.com/font), a new font family for Vercel.

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.

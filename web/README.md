# SupportPilot Web

Vite + TypeScript + React + Apollo Client. Three views: inbox list with filters, ticket detail
with the AI suggestion panel, account sidebar.

## Layout

```
src/lib/apolloClient.ts   Apollo Client setup (HttpLink + InMemoryCache)
src/graphql/               Operations (GET_TICKETS, TRIAGE_TICKET, ...) + hand-written types
                            mirroring api/graph/schema.graphqls
src/components/             InboxList, TicketDetail, AccountSidebar, NewTicketForm, Badges
```

## Run locally

Needs the API running (see [`../api/README.md`](../api/README.md) — `go run ./cmd/server` from `api/`,
no AWS credentials required):

```bash
npm install
npm run dev
```

Opens on http://localhost:5173, pointed at `http://localhost:8080/query` by default. To point at a
different API (e.g. once deployed), copy `.env.example` to `.env.local` and set `VITE_API_URL`.

## Build

```bash
npm run build   # tsc -b && vite build, output in dist/
```

## Deploying (Amplify Hosting)

Not deployed yet — needs an AWS account and the API's real `ApiUrl` (from the SAM stack output). Once
the API is deployed:

1. Connect this repo (or the `web/` subdirectory) to AWS Amplify Hosting.
2. Set the build's environment variable `VITE_API_URL` to the deployed API's `/query` URL.
3. Amplify auto-detects the Vite build (`npm run build`, publish `dist/`); no custom `amplify.yml`
   should be needed, but add one if the monorepo layout (`api/` + `web/`) trips up auto-detection —
   point `appRoot`/`baseDirectory` at `web/`.

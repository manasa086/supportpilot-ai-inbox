# SupportPilot

An AI-triaged support inbox for B2B teams — a mini-Pylon. Support tickets come in, an LLM classifies them, suggests a reply, and surfaces account context; a support rep works them in a React inbox.

## Stack

| Layer | Choice |
|---|---|
| API | Golang + [gqlgen](https://gqlgen.com/) (schema-first GraphQL, generated resolvers) |
| Schema | Single `schema.graphql` — `Ticket`, `Account`, `Suggestion` types; queries `tickets(status, priority)`, `ticket(id)`; mutations `createTicket`, `triageTicket`, `resolveTicket` |
| Infra | AWS Lambda (Go binary) behind API Gateway (HTTP API), DynamoDB single-table design (`PK=ACCOUNT#id`, `SK=TICKET#id`), Claude via Amazon Bedrock for AI calls, React app on Amplify Hosting. Deployed via SAM or CDK, infra-as-code lives in this repo. |
| Frontend | Vite + TypeScript + React + Apollo Client. Inbox list with filters, ticket detail with AI suggestion panel, account sidebar. |

## Repo layout

```
api/     Go GraphQL API (gqlgen), Lambda handler, DynamoDB resolvers, infra (SAM/CDK)
web/     React + Vite + TypeScript frontend (Apollo Client)
```

## Weekend MVP plan

**Saturday**
1. Write `schema.graphql`
2. `gqlgen generate`
3. Implement resolvers against DynamoDB
4. Deploy to Lambda
5. Exercise it from the GraphQL playground

**Sunday**
1. Build the React inbox list + ticket detail view
2. Wire `triageTicket` to Bedrock (returns category, priority, one-paragraph suggested reply, confidence)
3. Seed 30 fake tickets across 5 fake accounts
4. Deploy to Amplify Hosting

Stop there for the MVP.

## Stretch goals (only if the MVP ships)

- **Bulk triage mutation** — run classification over all `NEW` tickets in one call, demonstrating juggling multiple workstreams in the product itself.
- **Account-level context enrichment** — an `Account.summary` field resolved lazily via a second LLM call over that account's ticket history.
- **Semantic "similar resolved tickets"** via Bedrock embeddings — a third day of work; don't start it during the MVP weekend.

## Status

- ✅ **Saturday (API):** `schema.graphql` written, gqlgen resolvers implemented against a swappable `Store` interface
  (in-memory for local dev, DynamoDB for prod), AI triage behind a `Classifier` interface (rules-based locally, Bedrock
  in prod), 30 fake tickets / 5 fake accounts seeded, exercised end-to-end via `go run ./cmd/server` + the GraphQL
  playground. SAM template for the Lambda/API Gateway/DynamoDB deploy is in [`api/infra/template.yaml`](api/infra/template.yaml)
  but hasn't been deployed yet (needs AWS credentials + a Bedrock model ID). See [`api/README.md`](api/README.md).
- ✅ **Sunday (Web):** React (Vite + TS) inbox list / ticket detail / account sidebar built with Apollo Client,
  wired to the local API and verified end-to-end in-browser (filters, `triageTicket`, `resolveTicket`,
  `createTicket` all working live). See [`web/README.md`](web/README.md).
- ⬜ **Not yet done:** deploying the API (SAM) and the frontend (Amplify Hosting) to AWS — both need real
  AWS credentials and a Bedrock model ID, which weren't available in this session.

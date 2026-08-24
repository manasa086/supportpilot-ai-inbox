# SupportPilot API

Go + [gqlgen](https://gqlgen.com/) GraphQL service. Ticket CRUD, AI triage, and account context, backed by DynamoDB in production and callable locally with zero AWS dependencies.

## Layout

```
graph/               gqlgen: schema.graphqls, generated code, resolvers, model<->store conversion
internal/store/       Store interface + in-memory impl (local dev/tests) + DynamoDB impl (prod)
internal/triage/       Classifier interface + rules-based impl (local dev) + Bedrock impl (prod)
internal/seed/         Fake demo data: 5 accounts, 30 tickets
cmd/server/            Local dev entrypoint (plain HTTP server + GraphQL playground)
cmd/lambda/             Production entrypoint (API Gateway HTTP API -> Lambda)
infra/template.yaml     AWS SAM template: HTTP API + Lambda + DynamoDB table
```

## Run locally

No AWS credentials needed — this uses an in-memory store (seeded with fake data) and a keyword-based
rules classifier standing in for Bedrock.

```bash
go run ./cmd/server
```

Then open http://localhost:8080 for the GraphQL playground. Try:

```graphql
query { accounts { id name plan } }
query { tickets(status: NEW) { id subject account { name } } }
mutation { triageTicket(id: "<ticket id from above>") { status category priority suggestion { reply confidence } } }
```

### Using real Claude via Bedrock locally

```bash
export TRIAGE_BACKEND=bedrock
export BEDROCK_MODEL_ID=<a model/inference-profile ID enabled in your AWS account>
# picks up credentials from your normal AWS config (env vars, profile, SSO, etc.)
go run ./cmd/server
```

## Regenerating gqlgen code

After editing `graph/schema.graphqls`:

```bash
go run github.com/99designs/gqlgen generate
```

`graph/resolver.go` and `graph/convert.go` are hand-written and untouched by generation; `graph/schema.resolvers.go`
is regenerated but preserves existing resolver bodies.

## Tests

```bash
go test ./...
```

## Deploying (SAM)

Not run yet — needs the [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html),
an AWS account/credentials, and a Bedrock model ID enabled in that account/region:

```bash
cd infra
sam build --template-file template.yaml
sam deploy --guided --template-file template.yaml \
  --parameter-overrides BedrockModelId=<your-model-or-inference-profile-id>
```

This provisions:
- A DynamoDB table (`PK`/`SK` single-table design — see `internal/store/dynamo.go` for the item layout)
- The Go binary from `cmd/lambda`, built via the `Makefile`'s `build-SupportPilotFunction` target, deployed as a `provided.al2023` Lambda
- An API Gateway HTTP API in front of it, with CORS configured for the frontend origin(s)

After deploying, seed the table with demo data (script TBD — Sunday) and point the React app's Apollo
client at the stack's `ApiUrl` output.

# Invoked by `sam build` for the provided.al2023 (custom runtime) Lambda
# function defined in infra/template.yaml. SAM CLI looks for a
# `build-<LogicalId>` target here automatically — see
# https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/using-makefiles.html
build-SupportPilotFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/lambda

.PHONY: run
run:
	go run ./cmd/server

.PHONY: generate
generate:
	go run github.com/99designs/gqlgen generate

.PHONY: test
test:
	go test ./...

// Command lambda is the production entrypoint: the same GraphQL server as
// cmd/server, wired to DynamoDB and Bedrock instead of the in-memory/rules
// stand-ins, and adapted to run behind an API Gateway HTTP API (payload
// format 2.0) — see infra/template.yaml.
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/manasa086/supportpilot-ai-inbox/api/graph"
	"github.com/manasa086/supportpilot-ai-inbox/api/internal/store"
	"github.com/manasa086/supportpilot-ai-inbox/api/internal/triage"
)

var adapter *httpadapter.HandlerAdapterV2

func init() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("loading AWS config: %v", err)
	}

	tableName := mustEnv("TABLE_NAME")
	modelID := mustEnv("BEDROCK_MODEL_ID")

	resolver := &graph.Resolver{
		Store:      store.NewDynamo(dynamodb.NewFromConfig(cfg), tableName),
		Classifier: triage.NewBedrock(bedrockruntime.NewFromConfig(cfg), modelID),
	}
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{Cache: lru.New[string](100)})

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("SupportPilot GraphQL playground", "/query"))
	mux.Handle("/query", srv)

	adapter = httpadapter.NewV2(mux)
}

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("required environment variable %s is not set", name)
	}
	return v
}

func handleRequest(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return adapter.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(handleRequest)
}

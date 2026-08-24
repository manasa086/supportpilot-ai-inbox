// Command server runs SupportPilot's GraphQL API as a plain local HTTP
// server, for local development: `go run ./cmd/server`, then open
// http://localhost:8080 for the GraphQL playground.
//
// By default it uses an in-memory store (seeded with 5 fake accounts / 30
// fake tickets) and a zero-dependency rules-based classifier, so it runs
// with no AWS credentials at all. Set TRIAGE_BACKEND=bedrock (and
// BEDROCK_MODEL_ID) to exercise the real Claude-via-Bedrock path locally.
//
// cmd/lambda is the production entrypoint — same graph.Resolver, same
// schema, wired to DynamoDB + Bedrock instead.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/manasa086/supportpilot-ai-inbox/api/graph"
	"github.com/manasa086/supportpilot-ai-inbox/api/internal/seed"
	"github.com/manasa086/supportpilot-ai-inbox/api/internal/store"
	"github.com/manasa086/supportpilot-ai-inbox/api/internal/triage"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	mem := store.NewMemory()
	seed.Into(mem)

	resolver := &graph.Resolver{
		Store:      mem,
		Classifier: buildClassifier(),
	}
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("SupportPilot GraphQL playground", "/query"))
	mux.Handle("/query", srv)

	log.Printf("SupportPilot API listening — GraphQL playground at http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, withCORS(mux)))
}

func buildClassifier() triage.Classifier {
	if os.Getenv("TRIAGE_BACKEND") != "bedrock" {
		log.Println("triage backend: rules (local heuristic, no AWS calls) — set TRIAGE_BACKEND=bedrock + BEDROCK_MODEL_ID to use Claude via Bedrock")
		return triage.NewRules()
	}
	modelID := os.Getenv("BEDROCK_MODEL_ID")
	if modelID == "" {
		log.Fatal("TRIAGE_BACKEND=bedrock requires BEDROCK_MODEL_ID (see internal/triage/bedrock.go)")
	}
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("loading AWS config: %v", err)
	}
	log.Printf("triage backend: bedrock (model=%s)", modelID)
	return triage.NewBedrock(bedrockruntime.NewFromConfig(cfg), modelID)
}

// withCORS is a minimal hand-rolled CORS layer so the Vite dev server (and
// later the Amplify-hosted build) can call this API cross-origin without
// pulling in a router/middleware dependency. Configure extra allowed
// origins via ALLOWED_ORIGINS (comma-separated) once the frontend is
// deployed.
func withCORS(next http.Handler) http.Handler {
	allowed := map[string]bool{
		"http://localhost:5173": true, // Vite dev server default
		"http://localhost:3000": true,
	}
	for _, o := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			allowed[o] = true
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

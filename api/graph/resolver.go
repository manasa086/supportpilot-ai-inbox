package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

import (
	"github.com/manasa086/supportpilot-ai-inbox/api/internal/store"
	"github.com/manasa086/supportpilot-ai-inbox/api/internal/triage"
)

// Resolver holds the dependencies every resolver needs: the persistence
// layer and the AI classifier. Both are interfaces so cmd/server can wire
// up in-memory/rules implementations for local dev while cmd/lambda wires
// up DynamoDB/Bedrock for production — see internal/store and
// internal/triage.
type Resolver struct {
	Store      store.Store
	Classifier triage.Classifier
}

package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

import (
	"context"
	"fmt"

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

// triageAndSave classifies t (category, priority, suggested reply,
// confidence), moves its status to TRIAGED, and persists the result.
// Shared by the explicit triageTicket mutation and by CreateTicket's
// automatic triage-on-creation — see schema.resolvers.go.
func (r *Resolver) triageAndSave(ctx context.Context, t *store.Ticket) error {
	account, err := r.Store.GetAccount(ctx, t.AccountID)
	if err != nil {
		return fmt.Errorf("loading account %q: %w", t.AccountID, err)
	}

	result, err := r.Classifier.Triage(ctx, triage.Input{
		Subject:     t.Subject,
		Body:        t.Body,
		AccountName: account.Name,
		AccountPlan: account.Plan,
	})
	if err != nil {
		return fmt.Errorf("classifying ticket: %w", err)
	}

	t.Category = &result.Category
	t.Priority = &result.Priority
	t.Status = "TRIAGED"
	t.Suggestion = &store.Suggestion{
		Category:   result.Category,
		Priority:   result.Priority,
		Reply:      result.Reply,
		Confidence: result.Confidence,
	}

	if err := r.Store.UpdateTicket(ctx, t); err != nil {
		return fmt.Errorf("saving triage result: %w", err)
	}
	return nil
}

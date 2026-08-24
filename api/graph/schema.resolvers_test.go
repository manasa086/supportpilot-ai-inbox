package graph

import (
	"context"
	"testing"
	"time"

	"github.com/manasa086/supportpilot-ai-inbox/api/graph/model"
	"github.com/manasa086/supportpilot-ai-inbox/api/internal/store"
	"github.com/manasa086/supportpilot-ai-inbox/api/internal/triage"
)

func newTestResolver() (*Resolver, *store.Memory) {
	mem := store.NewMemory()
	mem.SeedAccount(&store.Account{ID: "acct-1", Name: "Acme", Plan: "Growth", Seats: 20, CreatedAt: time.Now()})
	return &Resolver{Store: mem, Classifier: triage.NewRules()}, mem
}

func TestCreateTicket_AutoTriages(t *testing.T) {
	r, _ := newTestResolver()
	ctx := context.Background()

	ticket, err := r.Mutation().CreateTicket(ctx, model.CreateTicketInput{
		AccountID:      "acct-1",
		Subject:        "Production outage: everything is down",
		Body:           "Total outage, urgent, please help immediately.",
		RequesterEmail: "user@acme.example",
	})
	if err != nil {
		t.Fatalf("CreateTicket: %v", err)
	}

	if ticket.Status != model.TicketStatusTriaged {
		t.Errorf("expected ticket to be auto-triaged (status TRIAGED), got %q", ticket.Status)
	}
	if ticket.Priority == nil {
		t.Fatal("expected priority to be set by auto-triage")
	}
	if *ticket.Priority != model.TicketPriorityUrgent {
		t.Errorf("expected URGENT priority for an outage ticket, got %q", *ticket.Priority)
	}
	if ticket.Suggestion == nil || ticket.Suggestion.Reply == "" {
		t.Error("expected a suggested reply to be attached")
	}
}

func TestTriageTicket_StillWorksExplicitly(t *testing.T) {
	r, _ := newTestResolver()
	ctx := context.Background()

	created, err := r.Mutation().CreateTicket(ctx, model.CreateTicketInput{
		AccountID:      "acct-1",
		Subject:        "Feature request: CSV export",
		Body:           "Would be great to export tickets as CSV.",
		RequesterEmail: "user@acme.example",
	})
	if err != nil {
		t.Fatalf("CreateTicket: %v", err)
	}
	// Already auto-triaged; re-triaging explicitly should still succeed.
	retriaged, err := r.Mutation().TriageTicket(ctx, created.ID)
	if err != nil {
		t.Fatalf("TriageTicket: %v", err)
	}
	if retriaged.Status != model.TicketStatusTriaged {
		t.Errorf("expected status TRIAGED after re-triage, got %q", retriaged.Status)
	}
}

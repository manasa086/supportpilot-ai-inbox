package store

import (
	"context"
	"testing"
	"time"
)

func TestMemory_CreateAndListTickets(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	m.SeedAccount(&Account{ID: "acct-1", Name: "Acme", Plan: "Growth", Seats: 10, CreatedAt: time.Now()})

	created, err := m.CreateTicket(ctx, NewTicketInput{
		AccountID:      "acct-1",
		Subject:        "Help",
		Body:           "Something broke",
		RequesterEmail: "user@acme.example",
	})
	if err != nil {
		t.Fatalf("CreateTicket: %v", err)
	}
	if created.Status != "NEW" {
		t.Errorf("expected status NEW, got %q", created.Status)
	}

	got, err := m.GetTicket(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetTicket: %v", err)
	}
	if got.Subject != "Help" {
		t.Errorf("expected subject %q, got %q", "Help", got.Subject)
	}

	byAccount, err := m.ListTicketsByAccount(ctx, "acct-1")
	if err != nil {
		t.Fatalf("ListTicketsByAccount: %v", err)
	}
	if len(byAccount) != 1 {
		t.Errorf("expected 1 ticket for account, got %d", len(byAccount))
	}

	if _, err := m.CreateTicket(ctx, NewTicketInput{AccountID: "does-not-exist"}); err == nil {
		t.Error("expected error creating a ticket for a nonexistent account")
	}
}

func TestMemory_ListTicketsFilters(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	m.SeedAccount(&Account{ID: "acct-1", Name: "Acme"})

	t1, _ := m.CreateTicket(ctx, NewTicketInput{AccountID: "acct-1", Subject: "A"})
	t2, _ := m.CreateTicket(ctx, NewTicketInput{AccountID: "acct-1", Subject: "B"})

	t1.Status = "TRIAGED"
	high := "HIGH"
	t1.Priority = &high
	if err := m.UpdateTicket(ctx, t1); err != nil {
		t.Fatalf("UpdateTicket: %v", err)
	}

	status := "TRIAGED"
	filtered, err := m.ListTickets(ctx, TicketFilter{Status: &status})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != t1.ID {
		t.Errorf("expected only t1 in TRIAGED filter, got %+v", filtered)
	}

	all, err := m.ListTickets(ctx, TicketFilter{})
	if err != nil {
		t.Fatalf("ListTickets: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 tickets total, got %d", len(all))
	}
	_ = t2
}

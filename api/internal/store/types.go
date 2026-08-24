// Package store defines the persistence layer for SupportPilot.
//
// Records here are plain domain structs, deliberately decoupled from the
// generated GraphQL models in graph/model — the resolvers do the mapping.
// This keeps the storage layer swappable (in-memory for local dev and
// tests, DynamoDB in production) without gqlgen-generated code leaking in.
//
// DynamoDB single-table layout (see infra/template.yaml):
//
//	Accounts: PK = ACCOUNT#<accountID>   SK = METADATA
//	Tickets:  PK = ACCOUNT#<accountID>   SK = TICKET#<ticketID>
//
// Listing an account's tickets is a single Query on PK with a SK begins_with
// "TICKET#" — no secondary index needed for the MVP's access patterns.
package store

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a lookup by ID finds nothing.
var ErrNotFound = errors.New("supportpilot: not found")

type Account struct {
	ID        string
	Name      string
	Plan      string
	Seats     int32
	CreatedAt time.Time
	// Summary is the cached result of the account-context LLM call (stretch
	// goal). Nil until AI.SummarizeAccount has been run for this account.
	Summary *string
}

type Ticket struct {
	ID             string
	AccountID      string
	Subject        string
	Body           string
	Status         string // NEW | TRIAGED | IN_PROGRESS | RESOLVED
	Priority       *string
	Category       *string
	RequesterEmail string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Suggestion     *Suggestion
}

type Suggestion struct {
	Category   string
	Priority   string
	Reply      string
	Confidence float64
}

// NewTicketInput is what CreateTicket needs; ID/timestamps/status are
// assigned by the store.
type NewTicketInput struct {
	AccountID      string
	Subject        string
	Body           string
	RequesterEmail string
}

// TicketFilter narrows ListTickets. Nil fields mean "don't filter on this".
type TicketFilter struct {
	Status   *string
	Priority *string
}

// Store is the persistence interface the GraphQL resolvers depend on.
type Store interface {
	ListAccounts(ctx context.Context) ([]*Account, error)
	GetAccount(ctx context.Context, id string) (*Account, error)
	// SetAccountSummary caches an LLM-generated summary for an account.
	SetAccountSummary(ctx context.Context, accountID, summary string) error

	ListTickets(ctx context.Context, filter TicketFilter) ([]*Ticket, error)
	ListTicketsByAccount(ctx context.Context, accountID string) ([]*Ticket, error)
	GetTicket(ctx context.Context, id string) (*Ticket, error)
	CreateTicket(ctx context.Context, in NewTicketInput) (*Ticket, error)
	// UpdateTicket persists a full ticket record (used after triage/resolve).
	UpdateTicket(ctx context.Context, t *Ticket) error
}

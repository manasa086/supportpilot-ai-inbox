package store

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Memory is an in-process Store implementation. It's used for local
// development (cmd/server) and unit tests so contributors don't need AWS
// credentials or a running DynamoDB Local to iterate on resolvers.
//
// Ticket IDs are "<accountID>:<uuid>" — the same composite scheme the
// DynamoDB store uses, so behavior (and seed data) is identical across
// backends. See internal/store/dynamo.go for why.
type Memory struct {
	mu       sync.RWMutex
	accounts map[string]*Account
	tickets  map[string]*Ticket // keyed by full ticket ID
}

func NewMemory() *Memory {
	return &Memory{
		accounts: make(map[string]*Account),
		tickets:  make(map[string]*Ticket),
	}
}

// SeedAccount inserts an account directly, bypassing normal validation.
// Intended for local dev / demo data seeding only.
func (m *Memory) SeedAccount(a *Account) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.accounts[a.ID] = a
}

// SeedTicket inserts a ticket directly, bypassing normal validation.
func (m *Memory) SeedTicket(t *Ticket) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tickets[t.ID] = t
}

func ticketID(accountID string) string {
	return fmt.Sprintf("%s:%s", accountID, uuid.NewString())
}

func (m *Memory) ListAccounts(_ context.Context) ([]*Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Account, 0, len(m.accounts))
	for _, a := range m.accounts {
		cp := *a
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (m *Memory) GetAccount(_ context.Context, id string) (*Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.accounts[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *a
	return &cp, nil
}

func (m *Memory) SetAccountSummary(_ context.Context, accountID, summary string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.accounts[accountID]
	if !ok {
		return ErrNotFound
	}
	a.Summary = &summary
	return nil
}

func (m *Memory) ListTickets(_ context.Context, filter TicketFilter) ([]*Ticket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Ticket, 0, len(m.tickets))
	for _, t := range m.tickets {
		if filter.Status != nil && t.Status != *filter.Status {
			continue
		}
		if filter.Priority != nil && (t.Priority == nil || *t.Priority != *filter.Priority) {
			continue
		}
		cp := *t
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (m *Memory) ListTicketsByAccount(_ context.Context, accountID string) ([]*Ticket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Ticket, 0)
	for _, t := range m.tickets {
		if t.AccountID == accountID {
			cp := *t
			out = append(out, &cp)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (m *Memory) GetTicket(_ context.Context, id string) (*Ticket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tickets[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (m *Memory) CreateTicket(_ context.Context, in NewTicketInput) (*Ticket, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.accounts[in.AccountID]; !ok {
		return nil, fmt.Errorf("%w: account %q", ErrNotFound, in.AccountID)
	}
	now := time.Now().UTC()
	t := &Ticket{
		ID:             ticketID(in.AccountID),
		AccountID:      in.AccountID,
		Subject:        in.Subject,
		Body:           in.Body,
		Status:         "NEW",
		RequesterEmail: in.RequesterEmail,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	m.tickets[t.ID] = t
	cp := *t
	return &cp, nil
}

func (m *Memory) UpdateTicket(_ context.Context, t *Ticket) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tickets[t.ID]; !ok {
		return ErrNotFound
	}
	cp := *t
	cp.UpdatedAt = time.Now().UTC()
	m.tickets[t.ID] = &cp
	return nil
}

var _ Store = (*Memory)(nil)

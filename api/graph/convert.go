package graph

// Mapping between internal/store domain records and the generated GraphQL
// models. Kept separate from schema.resolvers.go (which gqlgen rewrites)
// so it survives regeneration untouched.

import (
	"strings"

	"github.com/manasa086/supportpilot-ai-inbox/api/graph/model"
	"github.com/manasa086/supportpilot-ai-inbox/api/internal/store"
)

func toAccountModel(a *store.Account) *model.Account {
	if a == nil {
		return nil
	}
	return &model.Account{
		ID:        a.ID,
		Name:      a.Name,
		Plan:      a.Plan,
		Seats:     a.Seats,
		CreatedAt: a.CreatedAt,
		Summary:   a.Summary,
		// Tickets and Summary are @goField(forceResolver: true) — left
		// unset here, resolved lazily by accountResolver.
	}
}

func toTicketModel(t *store.Ticket) *model.Ticket {
	if t == nil {
		return nil
	}
	m := &model.Ticket{
		ID:             t.ID,
		Subject:        t.Subject,
		Body:           t.Body,
		Status:         model.TicketStatus(t.Status),
		RequesterEmail: t.RequesterEmail,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
		// Account is @goField(forceResolver: true) — resolved lazily by
		// ticketResolver.Account, which parses the account ID back out of
		// t.ID (see internal/store's "<accountID>:<uuid>" ID scheme).
	}
	if t.Priority != nil {
		p := model.TicketPriority(*t.Priority)
		m.Priority = &p
	}
	if t.Category != nil {
		m.Category = t.Category
	}
	if t.Suggestion != nil {
		m.Suggestion = &model.Suggestion{
			Category:   t.Suggestion.Category,
			Priority:   model.TicketPriority(t.Suggestion.Priority),
			Reply:      t.Suggestion.Reply,
			Confidence: t.Suggestion.Confidence,
		}
	}
	return m
}

func toTicketModels(ts []*store.Ticket) []*model.Ticket {
	out := make([]*model.Ticket, 0, len(ts))
	for _, t := range ts {
		out = append(out, toTicketModel(t))
	}
	return out
}

func toAccountModels(as []*store.Account) []*model.Account {
	out := make([]*model.Account, 0, len(as))
	for _, a := range as {
		out = append(out, toAccountModel(a))
	}
	return out
}

// ticketAccountID recovers the owning account ID from a ticket's composite
// "<accountID>:<uuid>" ID, without a round-trip to the store.
func ticketAccountID(ticketID string) (string, bool) {
	accountID, _, ok := strings.Cut(ticketID, ":")
	return accountID, ok && accountID != ""
}

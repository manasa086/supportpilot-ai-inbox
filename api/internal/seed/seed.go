// Package seed generates deterministic fake demo data — 5 accounts and 30
// tickets — shared by local dev (cmd/server) and the one-off DynamoDB
// seeding script used before the Amplify/Lambda deploy.
package seed

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/manasa086/supportpilot-ai-inbox/api/internal/store"
)

type accountSeed struct {
	name  string
	plan  string
	seats int32
}

var accountSeeds = []accountSeed{
	{"Northwind Traders", "Enterprise", 240},
	{"Globex Logistics", "Growth", 85},
	{"Initech", "Growth", 40},
	{"Umbrella Analytics", "Starter", 12},
	{"Soylent Retail", "Enterprise", 310},
}

type ticketSeed struct {
	subject string
	body    string
	email   string
	status  string // "" -> NEW
}

// ticketSeeds is deliberately longer than 30/5=6 per account on average is
// achieved by cycling through this list per account below.
var ticketSeeds = []ticketSeed{
	{"Production outage: API returning 500s", "Our integration has been down for 20 minutes, every request to /v2/orders returns a 500. This is blocking checkout for our customers right now.", "ops@", ""},
	{"Can't log in after password reset", "I reset my password an hour ago and the reset link now says expired when I try again. I'm locked out of the admin console.", "admin@", ""},
	{"Question about SSO setup", "We're trying to configure SAML SSO with Okta. Is there a guide for mapping custom attributes to roles?", "it@", ""},
	{"Feature request: bulk CSV export", "It would be great if we could export the full ticket list as CSV instead of paging through the UI one screen at a time.", "ops-lead@", ""},
	{"Invoice shows wrong seat count", "This month's invoice charges us for 45 seats but we only have 38 active users. Can you take a look?", "finance@", ""},
	{"Webhook payloads missing a field", "The 'resolved_at' field is missing from the ticket.updated webhook payload as of this week. Was this a recent change?", "eng@", ""},
	{"How do I set up custom fields?", "We'd like to add a 'region' field to tickets. Where in the settings can I define custom fields?", "support-lead@", ""},
	{"Refund request for duplicate charge", "We were charged twice for our March subscription. Please refund the duplicate transaction.", "billing@", ""},
	{"Data export request (GDPR)", "One of our end users has requested full deletion of their data under GDPR. What's the process on your side?", "legal@", ""},
	{"UI bug: filter dropdown closes immediately", "When I click the priority filter dropdown on the inbox view it opens and closes instantly, making it unusable.", "cs@", ""},
	{"Feature request: dark mode", "Several of our reps work late shifts and would love a dark mode option in the inbox.", "cs-manager@", ""},
	{"Slow load times on ticket detail page", "The ticket detail view has been taking 5-8 seconds to load for the last two days. Other pages seem fine.", "eng@", ""},
	{"Can we get a second admin seat added?", "We just hired a new team lead and need to add them as a second account admin. How do we do that?", "hr@", ""},
	{"Docs link is broken", "The 'Getting Started' link in your onboarding email 404s. Could you fix or resend?", "newuser@", ""},
	{"Suggestion: keyboard shortcuts for triage", "It would speed up our workflow a lot to have keyboard shortcuts for assign/resolve/snooze in the ticket queue.", "power-user@", ""},
}

func syntheticEmail(local, name string) string {
	return fmt.Sprintf("%s%s.example.com", local, slug(name))
}

func slug(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			out = append(out, r)
		case r >= 'A' && r <= 'Z':
			out = append(out, r+32)
		case r == ' ':
			out = append(out, '-')
		}
	}
	return string(out)
}

// Into populates the given store with 5 accounts and 30 tickets. Intended
// for a fresh store (in-memory for local dev, or an empty DynamoDB table
// before first deploy) — it doesn't check for existing data.
func Into(target interface {
	SeedAccount(*store.Account)
	SeedTicket(*store.Ticket)
}) {
	now := time.Now().UTC()

	accounts := make([]*store.Account, 0, len(accountSeeds))
	for i, as := range accountSeeds {
		a := &store.Account{
			ID:        uuid.NewString(),
			Name:      as.name,
			Plan:      as.plan,
			Seats:     as.seats,
			CreatedAt: now.AddDate(0, 0, -180+i*7),
		}
		accounts = append(accounts, a)
		target.SeedAccount(a)
	}

	ticketsPerAccount := 30 / len(accounts) // 6 each for 5 accounts
	seedIdx := 0
	for _, a := range accounts {
		for j := 0; j < ticketsPerAccount; j++ {
			ts := ticketSeeds[seedIdx%len(ticketSeeds)]
			seedIdx++
			created := now.Add(-time.Duration(seedIdx) * 3 * time.Hour)
			t := &store.Ticket{
				ID:             a.ID + ":" + uuid.NewString(),
				AccountID:      a.ID,
				Subject:        ts.subject,
				Body:           ts.body,
				Status:         "NEW",
				RequesterEmail: syntheticEmail(ts.email, a.Name),
				CreatedAt:      created,
				UpdatedAt:      created,
			}
			target.SeedTicket(t)
		}
	}
}

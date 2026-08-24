// Package triage classifies incoming support tickets: category, priority,
// a suggested reply, and a confidence score. The real implementation calls
// Claude via Amazon Bedrock (bedrock.go); Rules is a zero-dependency
// fallback used for local dev and tests when no AWS credentials are
// configured.
package triage

import "context"

// Input is the ticket context handed to a Classifier.
type Input struct {
	Subject     string
	Body        string
	AccountName string
	AccountPlan string
}

// Result is what triageTicket persists onto the ticket as its Suggestion,
// plus the Category/Priority that get promoted onto the ticket itself.
type Result struct {
	Category   string
	Priority   string // LOW | MEDIUM | HIGH | URGENT
	Reply      string
	Confidence float64
}

// Classifier triages one ticket.
type Classifier interface {
	Triage(ctx context.Context, in Input) (*Result, error)
}

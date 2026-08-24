package triage

import (
	"context"
	"strings"
)

// Rules is a keyword-based Classifier with no external dependencies. It
// backs local dev (`go run ./cmd/server`) so the full triageTicket flow
// works out of the box without AWS credentials or Bedrock model access —
// swap in Bedrock (see bedrock.go) by setting TRIAGE_BACKEND=bedrock.
type Rules struct{}

func NewRules() *Rules { return &Rules{} }

var urgentWords = []string{"down", "outage", "urgent", "asap", "breach", "data loss", "can't login", "cannot login", "production"}
var billingWords = []string{"invoice", "billing", "charge", "refund", "payment", "subscription", "price"}
var bugWords = []string{"bug", "error", "broken", "crash", "exception", "500", "stack trace"}
var featureWords = []string{"feature request", "would be great", "could you add", "enhancement", "suggestion"}
var howToWords = []string{"how do i", "how to", "documentation", "docs", "where is", "question"}

func containsAny(haystack string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(haystack, n) {
			return true
		}
	}
	return false
}

func (r *Rules) Triage(_ context.Context, in Input) (*Result, error) {
	text := strings.ToLower(in.Subject + " " + in.Body)

	category := "General"
	switch {
	case containsAny(text, billingWords):
		category = "Billing"
	case containsAny(text, bugWords):
		category = "Bug"
	case containsAny(text, featureWords):
		category = "Feature Request"
	case containsAny(text, howToWords):
		category = "How-To"
	}

	priority := "MEDIUM"
	confidence := 0.62
	switch {
	case containsAny(text, urgentWords):
		priority = "URGENT"
		confidence = 0.81
	case category == "Bug":
		priority = "HIGH"
		confidence = 0.7
	case category == "Feature Request":
		priority = "LOW"
		confidence = 0.58
	}

	reply := buildReply(in, category)

	return &Result{
		Category:   category,
		Priority:   priority,
		Reply:      reply,
		Confidence: confidence,
	}, nil
}

func buildReply(in Input, category string) string {
	var sb strings.Builder
	sb.WriteString("Hi there, thanks for reaching out")
	if in.AccountName != "" {
		sb.WriteString(" from " + in.AccountName)
	}
	sb.WriteString(". ")
	switch category {
	case "Billing":
		sb.WriteString("I've flagged this with our billing team and we'll get your account details sorted out shortly. In the meantime, could you confirm the invoice or transaction ID in question?")
	case "Bug":
		sb.WriteString("Sorry for the trouble — this looks like a bug on our end. I'm escalating it to engineering now. Could you share the steps to reproduce and any error messages you're seeing?")
	case "Feature Request":
		sb.WriteString("Thanks for the suggestion! I've logged this as a feature request for our product team to review. We'll follow up here if it's prioritized.")
	case "How-To":
		sb.WriteString("Happy to help walk you through this. I'll follow up shortly with the relevant documentation and steps.")
	default:
		sb.WriteString("I'm looking into this now and will follow up shortly with next steps.")
	}
	return sb.String()
}

var _ Classifier = (*Rules)(nil)

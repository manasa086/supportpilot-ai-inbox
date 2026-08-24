package triage

import (
	"context"
	"testing"
)

func TestRules_Triage(t *testing.T) {
	r := NewRules()
	ctx := context.Background()

	cases := []struct {
		name         string
		in           Input
		wantCategory string
		wantPriority string
	}{
		{
			name:         "outage is urgent bug",
			in:           Input{Subject: "Production down", Body: "Everything is broken, total outage."},
			wantCategory: "Bug",
			wantPriority: "URGENT",
		},
		{
			name:         "billing question",
			in:           Input{Subject: "Invoice question", Body: "Why was I charged twice on this invoice?"},
			wantCategory: "Billing",
			wantPriority: "MEDIUM",
		},
		{
			name:         "feature request is low priority",
			in:           Input{Subject: "Feature request: dark mode", Body: "Would be great to add dark mode."},
			wantCategory: "Feature Request",
			wantPriority: "LOW",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := r.Triage(ctx, tc.in)
			if err != nil {
				t.Fatalf("Triage: %v", err)
			}
			if result.Category != tc.wantCategory {
				t.Errorf("category = %q, want %q", result.Category, tc.wantCategory)
			}
			if result.Priority != tc.wantPriority {
				t.Errorf("priority = %q, want %q", result.Priority, tc.wantPriority)
			}
			if result.Reply == "" {
				t.Error("expected a non-empty suggested reply")
			}
			if result.Confidence <= 0 || result.Confidence > 1 {
				t.Errorf("confidence = %v, want in (0, 1]", result.Confidence)
			}
		})
	}
}

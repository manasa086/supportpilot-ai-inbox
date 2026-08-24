package triage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// Bedrock classifies tickets by calling Claude through Amazon Bedrock's
// Converse API.
//
// modelID must be a model or cross-region inference profile ID that's
// actually enabled in your AWS account/region — e.g.
//
//	aws bedrock list-inference-profiles \
//	  --query 'inferenceProfileSummaries[].inferenceProfileId'
//
// or check the Bedrock console's "Model access" page. There's no safe
// hardcoded default here: Bedrock's Claude catalog and ID format (a plain
// model ID vs. a "us.anthropic...." cross-region inference profile ID) vary
// by account, region, and rollout — pass it explicitly via the
// BEDROCK_MODEL_ID env var (see cmd/server/main.go / cmd/lambda/main.go).
type Bedrock struct {
	client  *bedrockruntime.Client
	modelID string
}

func NewBedrock(client *bedrockruntime.Client, modelID string) *Bedrock {
	return &Bedrock{client: client, modelID: modelID}
}

const systemPrompt = `You are a support-ticket triage assistant for a B2B SaaS company.
Given a support ticket, classify it and draft a first reply.

Rules:
- category: a short label such as "Billing", "Bug", "Feature Request", "How-To", "Account Access", or "General".
- priority: one of LOW, MEDIUM, HIGH, URGENT. Use URGENT only for outages, data loss, security issues, or a customer explicitly blocked from using the product.
- reply: one paragraph, professional and empathetic, addressed to the customer. Don't promise specific timelines you can't know.
- confidence: your confidence in this classification, from 0.0 to 1.0.

Respond with ONLY a JSON object, no prose, no markdown code fences:
{"category": string, "priority": "LOW"|"MEDIUM"|"HIGH"|"URGENT", "reply": string, "confidence": number}`

func (b *Bedrock) Triage(ctx context.Context, in Input) (*Result, error) {
	userPrompt := fmt.Sprintf(
		"Account: %s (%s plan)\nSubject: %s\n\nBody:\n%s",
		in.AccountName, in.AccountPlan, in.Subject, in.Body,
	)

	out, err := b.client.Converse(ctx, &bedrockruntime.ConverseInput{
		ModelId: aws.String(b.modelID),
		System: []types.SystemContentBlock{
			&types.SystemContentBlockMemberText{Value: systemPrompt},
		},
		Messages: []types.Message{
			{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{Value: userPrompt},
				},
			},
		},
		InferenceConfig: &types.InferenceConfiguration{
			MaxTokens:   aws.Int32(1024),
			Temperature: aws.Float32(0.2),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("bedrock converse: %w", err)
	}

	msg, ok := out.Output.(*types.ConverseOutputMemberMessage)
	if !ok || len(msg.Value.Content) == 0 {
		return nil, fmt.Errorf("bedrock converse: empty response")
	}
	textBlock, ok := msg.Value.Content[0].(*types.ContentBlockMemberText)
	if !ok {
		return nil, fmt.Errorf("bedrock converse: unexpected content block type %T", msg.Value.Content[0])
	}

	var parsed struct {
		Category   string  `json:"category"`
		Priority   string  `json:"priority"`
		Reply      string  `json:"reply"`
		Confidence float64 `json:"confidence"`
	}
	raw := strings.TrimSpace(textBlock.Value)
	raw = strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(raw, "```json"), "```"), "```")
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &parsed); err != nil {
		return nil, fmt.Errorf("parsing model output as JSON: %w (raw: %s)", err, raw)
	}

	return &Result{
		Category:   parsed.Category,
		Priority:   strings.ToUpper(parsed.Priority),
		Reply:      parsed.Reply,
		Confidence: parsed.Confidence,
	}, nil
}

var _ Classifier = (*Bedrock)(nil)

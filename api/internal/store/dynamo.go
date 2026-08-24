package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// Dynamo is the production Store, backed by a single DynamoDB table
// (see infra/template.yaml for the CloudFormation/SAM definition):
//
//	Item kind | PK                  | SK
//	----------|---------------------|-------------------
//	Account   | ACCOUNT#<accountID> | METADATA
//	Ticket    | ACCOUNT#<accountID> | TICKET#<ticketUUID>
//
// A ticket's public GraphQL ID is "<accountID>:<ticketUUID>" so GetTicket(id)
// can do a direct GetItem without a secondary index — the account is always
// known from the ID itself. Listing all tickets across accounts (the
// unfiltered `tickets` query) does a table Scan, which is fine at this
// project's scale (a few dozen tickets); a GSI keyed by status/createdAt
// would be the production-scale fix and is called out inline below.
type Dynamo struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamo(client *dynamodb.Client, tableName string) *Dynamo {
	return &Dynamo{client: client, tableName: tableName}
}

const (
	skMetadata     = "METADATA"
	skTicketPrefix = "TICKET#"
)

func accountPK(id string) string { return "ACCOUNT#" + id }

func splitTicketID(id string) (accountID, ticketUUID string, err error) {
	accountID, ticketUUID, ok := strings.Cut(id, ":")
	if !ok || accountID == "" || ticketUUID == "" {
		return "", "", fmt.Errorf("supportpilot: malformed ticket id %q", id)
	}
	return accountID, ticketUUID, nil
}

// --- item shapes -----------------------------------------------------------

type accountItem struct {
	PK        string  `dynamodbav:"PK"`
	SK        string  `dynamodbav:"SK"`
	Type      string  `dynamodbav:"Type"`
	ID        string  `dynamodbav:"ID"`
	Name      string  `dynamodbav:"Name"`
	Plan      string  `dynamodbav:"Plan"`
	Seats     int32   `dynamodbav:"Seats"`
	CreatedAt string  `dynamodbav:"CreatedAt"`
	Summary   *string `dynamodbav:"Summary,omitempty"`
}

func (i accountItem) toDomain() (*Account, error) {
	createdAt, err := time.Parse(time.RFC3339, i.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing CreatedAt: %w", err)
	}
	return &Account{
		ID:        i.ID,
		Name:      i.Name,
		Plan:      i.Plan,
		Seats:     i.Seats,
		CreatedAt: createdAt,
		Summary:   i.Summary,
	}, nil
}

func accountToItem(a *Account) accountItem {
	return accountItem{
		PK:        accountPK(a.ID),
		SK:        skMetadata,
		Type:      "ACCOUNT",
		ID:        a.ID,
		Name:      a.Name,
		Plan:      a.Plan,
		Seats:     a.Seats,
		CreatedAt: a.CreatedAt.UTC().Format(time.RFC3339),
		Summary:   a.Summary,
	}
}

type ticketItem struct {
	PK                   string   `dynamodbav:"PK"`
	SK                   string   `dynamodbav:"SK"`
	Type                 string   `dynamodbav:"Type"`
	ID                   string   `dynamodbav:"ID"`
	AccountID            string   `dynamodbav:"AccountID"`
	Subject              string   `dynamodbav:"Subject"`
	Body                 string   `dynamodbav:"Body"`
	Status               string   `dynamodbav:"Status"`
	Priority             *string  `dynamodbav:"Priority,omitempty"`
	Category             *string  `dynamodbav:"Category,omitempty"`
	RequesterEmail       string   `dynamodbav:"RequesterEmail"`
	CreatedAt            string   `dynamodbav:"CreatedAt"`
	UpdatedAt            string   `dynamodbav:"UpdatedAt"`
	SuggestionCategory   *string  `dynamodbav:"SuggestionCategory,omitempty"`
	SuggestionPriority   *string  `dynamodbav:"SuggestionPriority,omitempty"`
	SuggestionReply      *string  `dynamodbav:"SuggestionReply,omitempty"`
	SuggestionConfidence *float64 `dynamodbav:"SuggestionConfidence,omitempty"`
}

func (i ticketItem) toDomain() (*Ticket, error) {
	createdAt, err := time.Parse(time.RFC3339, i.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing CreatedAt: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, i.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing UpdatedAt: %w", err)
	}
	t := &Ticket{
		ID:             i.ID,
		AccountID:      i.AccountID,
		Subject:        i.Subject,
		Body:           i.Body,
		Status:         i.Status,
		Priority:       i.Priority,
		Category:       i.Category,
		RequesterEmail: i.RequesterEmail,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}
	if i.SuggestionCategory != nil {
		t.Suggestion = &Suggestion{
			Category:   deref(i.SuggestionCategory),
			Priority:   deref(i.SuggestionPriority),
			Reply:      deref(i.SuggestionReply),
			Confidence: derefF(i.SuggestionConfidence),
		}
	}
	return t, nil
}

func ticketToItem(t *Ticket) ticketItem {
	_, ticketUUID, _ := splitTicketID(t.ID)
	item := ticketItem{
		PK:             accountPK(t.AccountID),
		SK:             skTicketPrefix + ticketUUID,
		Type:           "TICKET",
		ID:             t.ID,
		AccountID:      t.AccountID,
		Subject:        t.Subject,
		Body:           t.Body,
		Status:         t.Status,
		Priority:       t.Priority,
		Category:       t.Category,
		RequesterEmail: t.RequesterEmail,
		CreatedAt:      t.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      t.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if t.Suggestion != nil {
		item.SuggestionCategory = aws.String(t.Suggestion.Category)
		item.SuggestionPriority = aws.String(t.Suggestion.Priority)
		item.SuggestionReply = aws.String(t.Suggestion.Reply)
		item.SuggestionConfidence = aws.Float64(t.Suggestion.Confidence)
	}
	return item
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefF(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

// --- Store implementation ---------------------------------------------------

func (d *Dynamo) ListAccounts(ctx context.Context) ([]*Account, error) {
	out, err := d.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(d.tableName),
		FilterExpression: aws.String("SK = :meta"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":meta": &types.AttributeValueMemberS{Value: skMetadata},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("scanning accounts: %w", err)
	}
	accounts := make([]*Account, 0, len(out.Items))
	for _, raw := range out.Items {
		var item accountItem
		if err := attributevalue.UnmarshalMap(raw, &item); err != nil {
			return nil, fmt.Errorf("unmarshalling account: %w", err)
		}
		a, err := item.toDomain()
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (d *Dynamo) GetAccount(ctx context.Context, id string) (*Account, error) {
	out, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: accountPK(id)},
			"SK": &types.AttributeValueMemberS{Value: skMetadata},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("getting account %q: %w", id, err)
	}
	if out.Item == nil {
		return nil, ErrNotFound
	}
	var item accountItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, fmt.Errorf("unmarshalling account: %w", err)
	}
	return item.toDomain()
}

func (d *Dynamo) SetAccountSummary(ctx context.Context, accountID, summary string) error {
	_, err := d.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: accountPK(accountID)},
			"SK": &types.AttributeValueMemberS{Value: skMetadata},
		},
		UpdateExpression: aws.String("SET Summary = :s"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":s": &types.AttributeValueMemberS{Value: summary},
		},
		ConditionExpression: aws.String("attribute_exists(PK)"),
	})
	if err != nil {
		return fmt.Errorf("setting account summary: %w", err)
	}
	return nil
}

// ListTickets scans the whole table for TICKET# items and filters in
// memory. Fine for a demo-scale table; swap for a GSI (e.g. GSI1PK=Status,
// GSI1SK=CreatedAt) if this ever needs to scale past a few thousand tickets.
func (d *Dynamo) ListTickets(ctx context.Context, filter TicketFilter) ([]*Ticket, error) {
	out, err := d.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(d.tableName),
		FilterExpression: aws.String("#type = :ticket"),
		ExpressionAttributeNames: map[string]string{
			"#type": "Type",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":ticket": &types.AttributeValueMemberS{Value: "TICKET"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("scanning tickets: %w", err)
	}
	tickets := make([]*Ticket, 0, len(out.Items))
	for _, raw := range out.Items {
		var item ticketItem
		if err := attributevalue.UnmarshalMap(raw, &item); err != nil {
			return nil, fmt.Errorf("unmarshalling ticket: %w", err)
		}
		if filter.Status != nil && item.Status != *filter.Status {
			continue
		}
		if filter.Priority != nil && (item.Priority == nil || *item.Priority != *filter.Priority) {
			continue
		}
		t, err := item.toDomain()
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	return tickets, nil
}

func (d *Dynamo) ListTicketsByAccount(ctx context.Context, accountID string) ([]*Ticket, error) {
	out, err := d.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(d.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :skPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":       &types.AttributeValueMemberS{Value: accountPK(accountID)},
			":skPrefix": &types.AttributeValueMemberS{Value: skTicketPrefix},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("querying tickets for account %q: %w", accountID, err)
	}
	tickets := make([]*Ticket, 0, len(out.Items))
	for _, raw := range out.Items {
		var item ticketItem
		if err := attributevalue.UnmarshalMap(raw, &item); err != nil {
			return nil, fmt.Errorf("unmarshalling ticket: %w", err)
		}
		t, err := item.toDomain()
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	return tickets, nil
}

func (d *Dynamo) GetTicket(ctx context.Context, id string) (*Ticket, error) {
	accountID, ticketUUID, err := splitTicketID(id)
	if err != nil {
		return nil, err
	}
	out, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: accountPK(accountID)},
			"SK": &types.AttributeValueMemberS{Value: skTicketPrefix + ticketUUID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("getting ticket %q: %w", id, err)
	}
	if out.Item == nil {
		return nil, ErrNotFound
	}
	var item ticketItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, fmt.Errorf("unmarshalling ticket: %w", err)
	}
	return item.toDomain()
}

func (d *Dynamo) CreateTicket(ctx context.Context, in NewTicketInput) (*Ticket, error) {
	if _, err := d.GetAccount(ctx, in.AccountID); err != nil {
		return nil, fmt.Errorf("looking up account %q: %w", in.AccountID, err)
	}
	now := time.Now().UTC()
	t := &Ticket{
		ID:             in.AccountID + ":" + uuid.NewString(),
		AccountID:      in.AccountID,
		Subject:        in.Subject,
		Body:           in.Body,
		Status:         "NEW",
		RequesterEmail: in.RequesterEmail,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	item, err := attributevalue.MarshalMap(ticketToItem(t))
	if err != nil {
		return nil, fmt.Errorf("marshalling ticket: %w", err)
	}
	if _, err := d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      item,
	}); err != nil {
		return nil, fmt.Errorf("putting ticket: %w", err)
	}
	return t, nil
}

func (d *Dynamo) UpdateTicket(ctx context.Context, t *Ticket) error {
	t.UpdatedAt = time.Now().UTC()
	item, err := attributevalue.MarshalMap(ticketToItem(t))
	if err != nil {
		return fmt.Errorf("marshalling ticket: %w", err)
	}
	if _, err := d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(d.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(PK)"),
	}); err != nil {
		return fmt.Errorf("updating ticket %q: %w", t.ID, err)
	}
	return nil
}

var _ Store = (*Dynamo)(nil)

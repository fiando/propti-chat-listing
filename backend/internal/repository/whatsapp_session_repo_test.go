package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type fakeWhatsAppSessionDynamoClient struct {
	lastUpdateInput *dynamodb.UpdateItemInput
	lastGetInput    *dynamodb.GetItemInput
	updateOutput    *dynamodb.UpdateItemOutput
	getOutput       *dynamodb.GetItemOutput
	updateErr       error
	getErr          error
}

func (f *fakeWhatsAppSessionDynamoClient) UpdateItem(_ context.Context, params *dynamodb.UpdateItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	f.lastUpdateInput = params
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	if f.updateOutput != nil {
		return f.updateOutput, nil
	}
	return &dynamodb.UpdateItemOutput{}, nil
}

func (f *fakeWhatsAppSessionDynamoClient) GetItem(_ context.Context, params *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	f.lastGetInput = params
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.getOutput != nil {
		return f.getOutput, nil
	}
	return &dynamodb.GetItemOutput{}, nil
}

func TestWhatsAppSessionRepoTrackCustomerInboundBuildsConditionalUpdate(t *testing.T) {
	client := &fakeWhatsAppSessionDynamoClient{}
	repo := newWhatsAppSessionRepoWithClient(client, "wa-session-table")

	inboundAt := time.Date(2026, 4, 5, 15, 0, 0, 0, time.FixedZone("WIB", 7*60*60))
	now := time.Date(2026, 4, 5, 8, 0, 0, 0, time.UTC)
	repo.now = func() time.Time { return now }
	err := repo.TrackCustomerInbound(context.Background(), "customer-1", "business-1", inboundAt)
	if err != nil {
		t.Fatalf("TrackCustomerInbound returned error: %v", err)
	}

	if client.lastUpdateInput == nil {
		t.Fatal("expected UpdateItem to be called")
	}
	if got := aws.ToString(client.lastUpdateInput.TableName); got != "wa-session-table" {
		t.Fatalf("expected table wa-session-table, got %q", got)
	}
	if got := aws.ToString(client.lastUpdateInput.ConditionExpression); got != "attribute_not_exists(lastCustomerInboundAt) OR lastCustomerInboundAt < :inboundAt" {
		t.Fatalf("unexpected condition expression: %q", got)
	}
	if got := aws.ToString(client.lastUpdateInput.UpdateExpression); got != "SET lastCustomerInboundAt = :inboundAt, updatedAt = :updatedAt" {
		t.Fatalf("unexpected update expression: %q", got)
	}
	if key, ok := client.lastUpdateInput.Key["customerId"].(*types.AttributeValueMemberS); !ok || key.Value != "customer-1" {
		t.Fatalf("unexpected customerId key: %#v", client.lastUpdateInput.Key["customerId"])
	}
	if key, ok := client.lastUpdateInput.Key["businessId"].(*types.AttributeValueMemberS); !ok || key.Value != "business-1" {
		t.Fatalf("unexpected businessId key: %#v", client.lastUpdateInput.Key["businessId"])
	}
	if got, ok := client.lastUpdateInput.ExpressionAttributeValues[":inboundAt"].(*types.AttributeValueMemberS); !ok || got.Value != inboundAt.UTC().Format(time.RFC3339Nano) {
		t.Fatalf("unexpected inboundAt value: %#v", client.lastUpdateInput.ExpressionAttributeValues[":inboundAt"])
	}
	if got, ok := client.lastUpdateInput.ExpressionAttributeValues[":updatedAt"].(*types.AttributeValueMemberS); !ok || got.Value != now.UTC().Format(time.RFC3339Nano) {
		t.Fatalf("unexpected updatedAt value: %#v", client.lastUpdateInput.ExpressionAttributeValues[":updatedAt"])
	}
}

func TestWhatsAppSessionRepoTrackCustomerInboundIgnoresStaleConditionalFailure(t *testing.T) {
	client := &fakeWhatsAppSessionDynamoClient{updateErr: &types.ConditionalCheckFailedException{Message: aws.String("stale")}}
	repo := newWhatsAppSessionRepoWithClient(client, "wa-session-table")

	err := repo.TrackCustomerInbound(context.Background(), "customer-1", "business-1", time.Now())
	if err != nil {
		t.Fatalf("expected stale inbound conditional failure to be ignored, got %v", err)
	}
}

func TestWhatsAppSessionRepoTrackCustomerInboundWrapsUpdateErrors(t *testing.T) {
	client := &fakeWhatsAppSessionDynamoClient{updateErr: errors.New("dynamo unavailable")}
	repo := newWhatsAppSessionRepoWithClient(client, "wa-session-table")

	err := repo.TrackCustomerInbound(context.Background(), "customer-1", "business-1", time.Now())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "track whatsapp inbound session") {
		t.Fatalf("expected contextual error message, got %v", err)
	}
}

func TestWhatsAppSessionRepoGetLastCustomerInboundReturnsTimestamp(t *testing.T) {
	inboundAt := time.Date(2026, 4, 5, 8, 0, 0, 123456000, time.UTC)
	client := &fakeWhatsAppSessionDynamoClient{getOutput: &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{
		"customerId":            &types.AttributeValueMemberS{Value: "customer-1"},
		"businessId":            &types.AttributeValueMemberS{Value: "business-1"},
		"lastCustomerInboundAt": &types.AttributeValueMemberS{Value: inboundAt.Format(time.RFC3339Nano)},
	}}}
	repo := newWhatsAppSessionRepoWithClient(client, "wa-session-table")

	result, err := repo.GetLastCustomerInbound(context.Background(), "customer-1", "business-1")
	if err != nil {
		t.Fatalf("GetLastCustomerInbound returned error: %v", err)
	}
	if result == nil || !result.Equal(inboundAt) {
		t.Fatalf("expected inbound timestamp %s, got %#v", inboundAt, result)
	}
	if got := aws.ToString(client.lastGetInput.TableName); got != "wa-session-table" {
		t.Fatalf("expected table wa-session-table, got %q", got)
	}
}

func TestWhatsAppSessionRepoGetLastCustomerInboundReturnsNilWhenMissing(t *testing.T) {
	client := &fakeWhatsAppSessionDynamoClient{getOutput: &dynamodb.GetItemOutput{}}
	repo := newWhatsAppSessionRepoWithClient(client, "wa-session-table")

	result, err := repo.GetLastCustomerInbound(context.Background(), "customer-1", "business-1")
	if err != nil {
		t.Fatalf("GetLastCustomerInbound returned error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil inbound timestamp, got %v", *result)
	}
}

func TestWhatsAppSessionRepoGetLastCustomerInboundRejectsInvalidTimestamp(t *testing.T) {
	client := &fakeWhatsAppSessionDynamoClient{getOutput: &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{
		"lastCustomerInboundAt": &types.AttributeValueMemberS{Value: "not-a-timestamp"},
	}}}
	repo := newWhatsAppSessionRepoWithClient(client, "wa-session-table")

	_, err := repo.GetLastCustomerInbound(context.Background(), "customer-1", "business-1")
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "parse last inbound timestamp") {
		t.Fatalf("expected parse context in error, got %v", err)
	}
}

func TestWhatsAppSessionRepoValidatesArguments(t *testing.T) {
	client := &fakeWhatsAppSessionDynamoClient{}
	repo := newWhatsAppSessionRepoWithClient(client, "wa-session-table")

	if err := repo.TrackCustomerInbound(context.Background(), "", "business-1", time.Now()); err == nil {
		t.Fatal("expected TrackCustomerInbound to reject empty customerID")
	}
	if _, err := repo.GetLastCustomerInbound(context.Background(), "customer-1", ""); err == nil {
		t.Fatal("expected GetLastCustomerInbound to reject empty businessID")
	}
}

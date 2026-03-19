package models

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func TestListingUnmarshalMapAcceptsNullImages(t *testing.T) {
	item := map[string]types.AttributeValue{
		"PK":               &types.AttributeValueMemberS{Value: "user-1#listing-1"},
		"SK":               &types.AttributeValueMemberS{Value: "listing-1"},
		"listingId":        &types.AttributeValueMemberS{Value: "listing-1"},
		"userId":           &types.AttributeValueMemberS{Value: "user-1"},
		"title":            &types.AttributeValueMemberS{Value: "Rumah contoh"},
		"description":      &types.AttributeValueMemberS{Value: "Deskripsi contoh"},
		"price":            &types.AttributeValueMemberN{Value: "850000000"},
		"priceUnit":        &types.AttributeValueMemberS{Value: "total"},
		"listingType":      &types.AttributeValueMemberS{Value: "sell"},
		"status":           &types.AttributeValueMemberS{Value: "active"},
		"moderationStatus": &types.AttributeValueMemberS{Value: "pending"},
		"images":           &types.AttributeValueMemberNULL{Value: true},
		"videos":           &types.AttributeValueMemberNULL{Value: true},
	}

	var listing Listing
	if err := attributevalue.UnmarshalMap(item, &listing); err != nil {
		t.Fatalf("expected listing with null images to unmarshal, got error: %v", err)
	}
	if len(listing.Images) != 0 {
		t.Fatalf("expected empty image entries after null unmarshal, got %d", len(listing.Images))
	}
	if listing.Images == nil {
		t.Fatalf("expected image entries to normalize to an empty slice, got nil")
	}
}

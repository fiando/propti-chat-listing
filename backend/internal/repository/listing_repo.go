package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fiando/propti/backend/internal/models"
)

// ListingRepo provides CRUD operations on the propti-listings DynamoDB table.
type ListingRepo struct {
	db *DynamoDB
}

// NewListingRepo creates a new ListingRepo.
func NewListingRepo(db *DynamoDB) *ListingRepo {
	return &ListingRepo{db: db}
}

// Put writes (creates or replaces) a listing.
func (r *ListingRepo) Put(ctx context.Context, listing *models.Listing) error {
	item, err := attributevalue.MarshalMap(listing)
	if err != nil {
		return fmt.Errorf("marshal listing: %w", err)
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.ListingsTable),
		Item:      item,
	})
	return err
}

// GetByID retrieves a listing by its composite key (PK = userId#listingId, SK = listingId).
func (r *ListingRepo) GetByID(ctx context.Context, userID, listingID string) (*models.Listing, error) {
	pk := fmt.Sprintf("%s#%s", userID, listingID)
	sk := listingID

	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.ListingsTable),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get listing: %w", err)
	}
	if result.Item == nil {
		return nil, nil
	}

	var listing models.Listing
	if err := attributevalue.UnmarshalMap(result.Item, &listing); err != nil {
		return nil, fmt.Errorf("unmarshal listing: %w", err)
	}
	return &listing, nil
}

// GetByListingID retrieves a listing using only the listingId via the GSI.
func (r *ListingRepo) GetByListingID(ctx context.Context, listingID string) (*models.Listing, error) {
	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.db.ListingsTable),
		IndexName:              aws.String("listingId-index"),
		KeyConditionExpression: aws.String("listingId = :lid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":lid": &types.AttributeValueMemberS{Value: listingID},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("query listing by id: %w", err)
	}
	if len(result.Items) == 0 {
		return nil, nil
	}

	var listing models.Listing
	if err := attributevalue.UnmarshalMap(result.Items[0], &listing); err != nil {
		return nil, fmt.Errorf("unmarshal listing: %w", err)
	}
	return &listing, nil
}

// ListByUserID returns all listings for a given user.
func (r *ListingRepo) ListByUserID(ctx context.Context, userID string, limit int32) ([]models.Listing, error) {
	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.db.ListingsTable),
		IndexName:              aws.String("userId-createdAt-index"),
		KeyConditionExpression: aws.String("userId = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: userID},
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list listings by user: %w", err)
	}

	listings := make([]models.Listing, 0, len(result.Items))
	for _, item := range result.Items {
		var l models.Listing
		if err := attributevalue.UnmarshalMap(item, &l); err != nil {
			continue
		}
		listings = append(listings, l)
	}
	return listings, nil
}

// CountMonthlyByUserID counts how many listings a user has created in the current month.
func (r *ListingRepo) CountMonthlyByUserID(ctx context.Context, userID string) (int, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.db.ListingsTable),
		IndexName:              aws.String("userId-createdAt-index"),
		KeyConditionExpression: aws.String("userId = :uid AND createdAt >= :start"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid":   &types.AttributeValueMemberS{Value: userID},
			":start": &types.AttributeValueMemberS{Value: monthStart},
		},
		Select: types.SelectCount,
	})
	if err != nil {
		return 0, fmt.Errorf("count monthly listings: %w", err)
	}
	return int(result.Count), nil
}

// Delete removes a listing from the table.
func (r *ListingRepo) Delete(ctx context.Context, userID, listingID string) error {
	pk := fmt.Sprintf("%s#%s", userID, listingID)
	_, err := r.db.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.db.ListingsTable),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: listingID},
		},
	})
	return err
}

// Scan returns listings with optional location/price filters (full-table scan — use for dev/small datasets).
func (r *ListingRepo) Scan(ctx context.Context, params *models.ListingSearchParams) ([]models.Listing, error) {
	filterExpr := "moderationStatus = :approved AND #st = :active"
	exprAttrValues := map[string]types.AttributeValue{
		":approved": &types.AttributeValueMemberS{Value: string(models.ModerationStatusApproved)},
		":active":   &types.AttributeValueMemberS{Value: string(models.ListingStatusActive)},
	}
	exprAttrNames := map[string]string{
		"#st": "status",
	}

	if params.Province != "" {
		exprAttrNames["#loc"] = "location"
		exprAttrNames["#province"] = "province"
		filterExpr += " AND #loc.#province = :province"
		exprAttrValues[":province"] = &types.AttributeValueMemberS{Value: params.Province}
	}
	if params.City != "" {
		exprAttrNames["#loc"] = "location"
		exprAttrNames["#city"] = "city"
		filterExpr += " AND #loc.#city = :city"
		exprAttrValues[":city"] = &types.AttributeValueMemberS{Value: params.City}
	}
	if params.PriceMin > 0 {
		filterExpr += " AND price >= :pmin"
		exprAttrValues[":pmin"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", params.PriceMin)}
	}
	if params.PriceMax > 0 {
		filterExpr += " AND price <= :pmax"
		exprAttrValues[":pmax"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", params.PriceMax)}
	}
	if params.Bedrooms > 0 {
		filterExpr += " AND bedrooms = :bed"
		exprAttrValues[":bed"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", params.Bedrooms)}
	}

	limit := int32(params.PageSize)
	if limit <= 0 {
		limit = 20
	}

	result, err := r.db.Client.Scan(ctx, &dynamodb.ScanInput{
		TableName:                 aws.String(r.db.ListingsTable),
		FilterExpression:          aws.String(filterExpr),
		ExpressionAttributeValues: exprAttrValues,
		ExpressionAttributeNames:  exprAttrNames,
		Limit:                     aws.Int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("scan listings: %w", err)
	}

	listings := make([]models.Listing, 0, len(result.Items))
	for _, item := range result.Items {
		var l models.Listing
		if err := attributevalue.UnmarshalMap(item, &l); err != nil {
			continue
		}
		listings = append(listings, l)
	}
	return listings, nil
}

// ScanNearby queries listings near a given latitude/longitude within radiusKm.
func (r *ListingRepo) ScanNearby(ctx context.Context, lat, lng, radiusKm float64, limit int32) ([]models.Listing, error) {
	// Approximate bounding-box filter: 1 degree ≈ 111 km
	delta := radiusKm / 111.0
	latMin := lat - delta
	latMax := lat + delta
	lngMin := lng - delta
	lngMax := lng + delta

	result, err := r.db.Client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(r.db.ListingsTable),
		FilterExpression: aws.String("latitude BETWEEN :latMin AND :latMax AND longitude BETWEEN :lngMin AND :lngMax AND moderationStatus = :approved AND #st = :active"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":latMin":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", latMin)},
			":latMax":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", latMax)},
			":lngMin":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", lngMin)},
			":lngMax":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", lngMax)},
			":approved": &types.AttributeValueMemberS{Value: string(models.ModerationStatusApproved)},
			":active":   &types.AttributeValueMemberS{Value: string(models.ListingStatusActive)},
		},
		ExpressionAttributeNames: map[string]string{
			"#st": "status",
		},
		Limit: aws.Int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("scan nearby: %w", err)
	}

	listings := make([]models.Listing, 0, len(result.Items))
	for _, item := range result.Items {
		var l models.Listing
		if err := attributevalue.UnmarshalMap(item, &l); err != nil {
			continue
		}
		listings = append(listings, l)
	}
	return listings, nil
}

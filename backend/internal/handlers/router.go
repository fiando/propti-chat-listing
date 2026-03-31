package handlers

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fiando/propti/backend/internal/utils"
)

// CombinedListingHandler dispatches an APIGatewayProxyRequest to either the listing
// handler or the search handler based on the path prefix.
func CombinedListingHandler(ctx context.Context, rawReq interface{}, listingHandler *ListingHandler, searchHandler *SearchHandler, leadHandler *LeadHandler) (events.APIGatewayProxyResponse, error) {
	// Re-encode and decode to get a typed request (Lambda passes map[string]interface{} when invoked via Start(func)).
	b, err := json.Marshal(rawReq)
	if err != nil {
		return jsonResponse(500, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	var req events.APIGatewayProxyRequest
	if err := json.Unmarshal(b, &req); err != nil {
		return jsonResponse(500, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	path := req.Path
	if strings.HasPrefix(path, "/leads") {
		return leadHandler.Handle(ctx, req)
	}
	if strings.HasPrefix(path, "/search/") || strings.HasPrefix(path, "/locations/") {
		return searchHandler.Handle(ctx, req)
	}
	return listingHandler.Handle(ctx, req)
}

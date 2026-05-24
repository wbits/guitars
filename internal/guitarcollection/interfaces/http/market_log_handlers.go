package httpapi

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"

	"github.com/wbits/guitars/internal/guitarcollection/application"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/auth"
)

func (h *Handler) listMarketLogs(ctx context.Context, ownerID, guitarID string) (events.APIGatewayProxyResponse, error) {
	logs, err := h.marketLogs.ListMarketLogs(ctx, ownerID, guitarID)
	if err != nil {
		return errorToResponse(err)
	}
	out := make([]marketLogResponse, 0, len(logs))
	for _, log := range logs {
		out = append(out, toMarketLogResponse(log))
	}
	return jsonResponse(200, out)
}

func (h *Handler) createMarketLogs(ctx context.Context, callerID, callerEmail, guitarID, body string) (events.APIGatewayProxyResponse, error) {
	reqs, err := decodeMarketLogRequests(body)
	if err != nil {
		return jsonResponse(400, errorResponse{Error: err.Error()})
	}
	if len(reqs) == 0 {
		return jsonResponse(400, errorResponse{Error: "at least one market log entry is required"})
	}
	inputs := make([]application.MarketLogInput, 0, len(reqs))
	for _, req := range reqs {
		input, err := marketLogRequestToInput(req)
		if err != nil {
			return jsonResponse(400, errorResponse{Error: err.Error()})
		}
		inputs = append(inputs, input)
	}
	logs, err := h.marketLogs.AddMarketLogs(ctx, callerID, callerEmail, guitarID, inputs)
	if err != nil {
		return errorToResponse(err)
	}
	out := make([]marketLogResponse, 0, len(logs))
	for _, log := range logs {
		out = append(out, toMarketLogResponse(log))
	}
	if len(out) == 1 {
		return jsonResponse(201, out[0])
	}
	return jsonResponse(201, out)
}

func (h *Handler) deleteCollectionMarketLogs(ctx context.Context, principal auth.Principal, userID string) (events.APIGatewayProxyResponse, error) {
	if !auth.IsAdmin(principal, h.adminGroup) {
		return jsonResponse(403, errorResponse{Error: "forbidden"})
	}
	deleted, err := h.marketLogs.ClearCollectionMarketLogs(ctx, userID)
	if err != nil {
		return errorToResponse(err)
	}
	return jsonResponse(200, clearCollectionMarketLogsResponse{DeletedCount: deleted})
}

func decodeMarketLogRequests(body string) ([]marketLogRequest, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, errInvalidJSON
	}
	if body[0] == '[' {
		var batch []marketLogRequest
		if err := json.Unmarshal([]byte(body), &batch); err != nil {
			return nil, errInvalidJSON
		}
		return batch, nil
	}
	var single marketLogRequest
	if err := json.Unmarshal([]byte(body), &single); err != nil {
		return nil, errInvalidJSON
	}
	return []marketLogRequest{single}, nil
}

var errInvalidJSON = errorString("invalid JSON body")

type errorString string

func (e errorString) Error() string { return string(e) }

func marketLogRequestToInput(req marketLogRequest) (application.MarketLogInput, error) {
	var observedAt time.Time
	if strings.TrimSpace(req.ObservedAt) != "" {
		parsed, err := time.Parse(time.RFC3339, req.ObservedAt)
		if err != nil {
			return application.MarketLogInput{}, errInvalidObservedAt
		}
		observedAt = parsed.UTC()
	}
	return application.MarketLogInput{
		ObservedAt:        observedAt,
		Source:            req.Source,
		Action:            req.Action,
		PriceAmount:       req.PriceAmount,
		PriceCurrency:     req.PriceCurrency,
		ListingURL:        req.ListingURL,
		ListingTitle:      req.ListingTitle,
		ExternalListingID: req.ExternalListingID,
		ListingImageURL:   req.ListingImageURL,
	}, nil
}

var errInvalidObservedAt = errorString("observedAt must be an RFC3339 timestamp")

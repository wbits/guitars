package httpapi

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"

	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/auth"
)

func (h *Handler) putAssistantBYOK(ctx context.Context, principal auth.Principal, body string) (events.APIGatewayProxyResponse, error) {
	if h.profiles == nil {
		return jsonResponse(503, errorResponse{Error: "profiles are not configured"})
	}
	var req assistantBYOKPutRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return jsonResponse(400, errorResponse{Error: "invalid JSON body"})
	}
	profile, err := h.profiles.SetAssistantBYOK(ctx, principal.UserID, principal.Email, req.APIKey, req.BaseURL, req.Model)
	if err != nil {
		return profileErrorToResponse(err)
	}
	return jsonResponse(200, toMeResponse(profile, auth.IsAdmin(principal, h.adminGroup)))
}

func (h *Handler) deleteAssistantBYOK(ctx context.Context, principal auth.Principal) (events.APIGatewayProxyResponse, error) {
	if h.profiles == nil {
		return jsonResponse(503, errorResponse{Error: "profiles are not configured"})
	}
	profile, err := h.profiles.ClearAssistantBYOK(ctx, principal.UserID, principal.Email)
	if err != nil {
		return profileErrorToResponse(err)
	}
	return jsonResponse(200, toMeResponse(profile, auth.IsAdmin(principal, h.adminGroup)))
}

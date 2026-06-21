package httpapi

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/aws/aws-lambda-go/events"

	"github.com/wbits/guitars/internal/assistant"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/auth"
)

type assistantChatRequest struct {
	CollectionUserID string `json:"collectionUserId"`
	Message          string `json:"message"`
}

type assistantChatResponse struct {
	Message     string            `json:"message"`
	MatchingIDs []string          `json:"matchingIds"`
	Filter      *assistant.Filter `json:"filter,omitempty"`
}

func (h *Handler) assistantChat(ctx context.Context, principal auth.Principal, body string) (events.APIGatewayProxyResponse, error) {
	if h.assistant == nil {
		return jsonResponse(503, errorResponse{Error: "assistant is not configured"})
	}
	var req assistantChatRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return jsonResponse(400, errorResponse{Error: "invalid JSON body"})
	}
	out, err := h.assistant.Chat(ctx, assistant.ChatRequest{
		CollectionUserID: req.CollectionUserID,
		Message:          req.Message,
		CallerUserID:     principal.UserID,
	})
	if err != nil {
		if assistant.IsRateLimited(err) {
			return jsonResponse(429, errorResponse{Error: err.Error()})
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return jsonResponse(504, errorResponse{Error: "assistant timed out"})
		}
		return jsonResponse(400, errorResponse{Error: err.Error()})
	}
	return jsonResponse(200, assistantChatResponse{
		Message:     out.Message,
		MatchingIDs: out.MatchingIDs,
		Filter:      out.Filter,
	})
}

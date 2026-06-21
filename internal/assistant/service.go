package assistant

import (
	"context"
	"fmt"
	"strings"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// ChatRequest is the viewer assistant input.
type ChatRequest struct {
	CollectionUserID string
	Message          string
	CallerUserID     string
}

// ChatResponse drives the webapp gallery and chat transcript.
type ChatResponse struct {
	Message     string   `json:"message"`
	MatchingIDs []string `json:"matchingIds"`
	Filter      *Filter  `json:"filter,omitempty"`
}

// GuitarLister loads a public collection.
type GuitarLister interface {
	ListUserGuitars(ctx context.Context, userID string) ([]*domain.Guitar, error)
}

// Service runs hosted viewer chat (tier 1 — operator LLM key, rate limited).
type Service struct {
	guitars GuitarLister
	llm     LLM
	limiter RateLimiter
}

// NewService constructs a viewer assistant service.
func NewService(guitars GuitarLister, llm LLM, limiter RateLimiter) *Service {
	if llm == nil {
		llm = RuleLLM{}
	}
	if limiter == nil {
		limiter = NewMemoryRateLimiter(10)
	}
	return &Service{guitars: guitars, llm: llm, limiter: limiter}
}

// Chat applies rate limits, parses the message, and returns matching guitar ids.
func (s *Service) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	collectionUserID := strings.TrimSpace(req.CollectionUserID)
	message := strings.TrimSpace(req.Message)
	callerUserID := strings.TrimSpace(req.CallerUserID)
	if collectionUserID == "" {
		return ChatResponse{}, fmt.Errorf("collectionUserId is required")
	}
	if message == "" {
		return ChatResponse{}, fmt.Errorf("message is required")
	}
	if callerUserID == "" {
		return ChatResponse{}, fmt.Errorf("caller identity is required")
	}
	if err := s.limiter.Allow(ctx, callerUserID); err != nil {
		return ChatResponse{}, err
	}

	guitars, err := s.guitars.ListUserGuitars(ctx, collectionUserID)
	if err != nil {
		return ChatResponse{}, err
	}

	filter, reply, err := s.llm.ParseFilter(ctx, message, guitars)
	if err != nil {
		return ChatResponse{}, err
	}

	matched := ApplyFilter(guitars, filter)
	if len(matched) == 0 && !filter.isEmpty() {
		reply = strings.TrimSpace(reply) + fmt.Sprintf(" No guitars match (%d in collection).", len(guitars))
	} else if !filter.isEmpty() {
		reply = strings.TrimSpace(reply) + fmt.Sprintf(" Showing %d of %d.", len(matched), len(guitars))
	}

	var filterOut *Filter
	if !filter.isEmpty() {
		f := filter
		filterOut = &f
	}

	return ChatResponse{
		Message:     reply,
		MatchingIDs: guitarIDs(matched),
		Filter:      filterOut,
	}, nil
}

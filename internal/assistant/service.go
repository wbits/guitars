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

// Service runs viewer chat with tier-1 hosted LLM or tier-2 owner BYOK on own collection.
type Service struct {
	guitars      GuitarLister
	hostedLLM    LLM
	tier1Limiter RateLimiter
	byok         BYOKProvider
	byokLimiter  RateLimiter
	analysis     *GuitarAnalysisIndex
}

// NewService constructs an assistant service.
func NewService(guitars GuitarLister, hostedLLM LLM, tier1Limiter RateLimiter, byok BYOKProvider, byokLimiter RateLimiter, analysis *GuitarAnalysisIndex) *Service {
	if hostedLLM == nil {
		hostedLLM = RuleLLM{}
	}
	if tier1Limiter == nil {
		tier1Limiter = NewMemoryRateLimiter(10)
	}
	if byokLimiter == nil {
		byokLimiter = NewMemoryRateLimiter(200)
	}
	return &Service{
		guitars:      guitars,
		hostedLLM:    hostedLLM,
		tier1Limiter: tier1Limiter,
		byok:         byok,
		byokLimiter:  byokLimiter,
		analysis:     analysis,
	}
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
	llm, limiter := s.resolveTier(ctx, collectionUserID, callerUserID)
	if err := limiter.Allow(ctx, callerUserID); err != nil {
		return ChatResponse{}, err
	}

	guitars, err := s.guitars.ListUserGuitars(ctx, collectionUserID)
	if err != nil {
		return ChatResponse{}, err
	}

	filter, reply, err := llm.ParseFilter(ctx, message, guitars)
	if err != nil {
		return ChatResponse{}, err
	}

	analysisMap := map[string]AnalysisSearch{}
	if s.analysis != nil {
		analysisMap, err = s.analysis.MapForGuitars(ctx, guitarIDs(guitars))
		if err != nil {
			return ChatResponse{}, err
		}
	}

	matched := ApplyFilter(guitars, filter, analysisMap)
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

func (s *Service) resolveTier(ctx context.Context, collectionUserID, callerUserID string) (LLM, RateLimiter) {
	if s.byok != nil &&
		strings.TrimSpace(collectionUserID) != "" &&
		collectionUserID == strings.TrimSpace(callerUserID) {
		creds, ok, err := s.byok.Credentials(ctx, collectionUserID)
		if err == nil && ok && strings.TrimSpace(creds.APIKey) != "" {
			return &OpenAICompatibleLLM{
				APIKey:  creds.APIKey,
				BaseURL: creds.BaseURL,
				Model:   creds.Model,
			}, s.byokLimiter
		}
	}
	return s.hostedLLM, s.tier1Limiter
}

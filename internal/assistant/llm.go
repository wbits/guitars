package assistant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// LLM interprets a viewer question against collection context.
type LLM interface {
	ParseFilter(ctx context.Context, message string, guitars []*domain.Guitar) (Filter, string, error)
}

// RuleLLM uses deterministic parsing only.
type RuleLLM struct{}

func (RuleLLM) ParseFilter(_ context.Context, message string, guitars []*domain.Guitar) (Filter, string, error) {
	f, reply := ParseRules(message, guitars)
	return f, reply, nil
}

// OpenAICompatibleLLM calls a chat-completions API (OpenAI or compatible).
type OpenAICompatibleLLM struct {
	APIKey  string
	BaseURL string
	Model   string
	Client  *http.Client
}

func (o *OpenAICompatibleLLM) ParseFilter(ctx context.Context, message string, guitars []*domain.Guitar) (Filter, string, error) {
	if strings.TrimSpace(o.APIKey) == "" {
		return RuleLLM{}.ParseFilter(ctx, message, guitars)
	}
	brands := distinctBrands(guitars)
	system := buildLLMSystemPrompt(brands)
	user := strings.TrimSpace(message)

	body, err := json.Marshal(map[string]any{
		"model": o.model(),
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"temperature": 0,
		"response_format": map[string]string{"type": "json_object"},
	})
	if err != nil {
		return Filter{}, "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL()+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Filter{}, "", err
	}
	req.Header.Set("Authorization", "Bearer "+o.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := o.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return RuleLLM{}.ParseFilter(ctx, message, guitars)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return RuleLLM{}.ParseFilter(ctx, message, guitars)
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil || len(parsed.Choices) == 0 {
		return RuleLLM{}.ParseFilter(ctx, message, guitars)
	}

	var out struct {
		Message       string  `json:"message"`
		Brand         string  `json:"brand"`
		TypeName      string  `json:"typeName"`
		Color         string  `json:"color"`
		MinPriceMajor *float64 `json:"minPriceMajor"`
		MaxPriceMajor *float64 `json:"maxPriceMajor"`
		MinYear       *int     `json:"minYear"`
		MaxYear       *int     `json:"maxYear"`
	}
	if err := json.Unmarshal([]byte(parsed.Choices[0].Message.Content), &out); err != nil {
		return RuleLLM{}.ParseFilter(ctx, message, guitars)
	}
	f := Filter{
		Brand:         strings.TrimSpace(out.Brand),
		TypeName:      strings.TrimSpace(out.TypeName),
		Color:         strings.TrimSpace(out.Color),
		MinPriceMajor: out.MinPriceMajor,
		MaxPriceMajor: out.MaxPriceMajor,
		MinYear:       out.MinYear,
		MaxYear:       out.MaxYear,
	}
	reply := strings.TrimSpace(out.Message)
	if reply == "" {
		_, reply = ParseRules(message, guitars)
	}
	return f, reply, nil
}

func (o *OpenAICompatibleLLM) baseURL() string {
	if u := strings.TrimRight(strings.TrimSpace(o.BaseURL), "/"); u != "" {
		return u
	}
	return "https://api.openai.com/v1"
}

func (o *OpenAICompatibleLLM) model() string {
	if m := strings.TrimSpace(o.Model); m != "" {
		return m
	}
	return "gpt-4o-mini"
}

func distinctBrands(guitars []*domain.Guitar) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0)
	for _, g := range guitars {
		b := strings.TrimSpace(g.Brand())
		if b == "" {
			continue
		}
		key := strings.ToLower(b)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, b)
	}
	return out
}

func buildLLMSystemPrompt(brands []string) string {
	brandList := "none"
	if len(brands) > 0 {
		brandList = strings.Join(brands, ", ")
	}
	return fmt.Sprintf(`You help visitors filter a guitar collection. Brands in this collection: %s.
Return JSON only with keys: message (short friendly reply), brand, typeName, color, minPriceMajor, maxPriceMajor, minYear, maxYear.
Use major currency units for prices (e.g. 1000 for €1000). Omit unused fields or set them null.
Never invent guitars. Only suggest filters; do not claim to mutate data.`, brandList)
}

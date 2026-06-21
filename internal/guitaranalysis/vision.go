package guitaranalysis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// VisionCredentials are owner BYOK settings for vision calls.
type VisionCredentials struct {
	APIKey  string
	BaseURL string
	Model   string
}

// VisionAnalyzer calls an OpenAI-compatible vision model.
type VisionAnalyzer struct {
	Client *http.Client
}

type visionResult struct {
	VisualSummary string   `json:"visualSummary"`
	Tags          []string `json:"tags"`
	Confidence    float64  `json:"confidence"`
}

func (v *VisionAnalyzer) AnalyzePictures(ctx context.Context, creds VisionCredentials, pictureURLs []string, guitarBrand, guitarType string) (visionResult, error) {
	if v == nil {
		return visionResult{}, fmt.Errorf("vision analyzer not configured")
	}
	urls := limitPictureURLs(pictureURLs, 3)
	if len(urls) == 0 {
		return visionResult{}, InvalidField("pictures", "requires at least one image URL")
	}
	model := strings.TrimSpace(creds.Model)
	if model == "" {
		model = "gpt-4o-mini"
	}
	content := []map[string]any{
		{
			"type": "text",
			"text": fmt.Sprintf(`Analyze these guitar photos for a collection catalog. The owner entered brand=%q model=%q — treat that as context only; describe what you see.
Return JSON only with keys: visualSummary (2-3 sentences), tags (array of lowercase kebab-case visual tags like "sunburst", "humbucker", "maple-neck"), confidence (0-1).
Do not guess serial numbers or exact year. Focus on visible features useful for search.`, guitarBrand, guitarType),
		},
	}
	for _, url := range urls {
		content = append(content, map[string]any{
			"type": "image_url",
			"image_url": map[string]string{"url": url},
		})
	}
	body, err := json.Marshal(map[string]any{
		"model": model,
		"messages": []map[string]any{
			{"role": "user", "content": content},
		},
		"temperature":     0,
		"response_format": map[string]string{"type": "json_object"},
	})
	if err != nil {
		return visionResult{}, err
	}
	baseURL := strings.TrimRight(strings.TrimSpace(creds.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return visionResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(creds.APIKey))
	req.Header.Set("Content-Type", "application/json")
	client := v.Client
	if client == nil {
		client = &http.Client{Timeout: 45 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return visionResult{}, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return visionResult{}, fmt.Errorf("vision API status %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil || len(parsed.Choices) == 0 {
		return visionResult{}, fmt.Errorf("invalid vision API response")
	}
	var out visionResult
	if err := json.Unmarshal([]byte(parsed.Choices[0].Message.Content), &out); err != nil {
		return visionResult{}, fmt.Errorf("parse vision JSON: %w", err)
	}
	out.VisualSummary = strings.TrimSpace(out.VisualSummary)
	out.Tags = normalizeTags(out.Tags)
	if out.VisualSummary == "" {
		return visionResult{}, fmt.Errorf("vision model returned empty summary")
	}
	if out.Confidence <= 0 {
		out.Confidence = 0.7
	}
	return out, nil
}

func limitPictureURLs(urls []string, max int) []string {
	out := make([]string, 0, max)
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		out = append(out, u)
		if len(out) >= max {
			break
		}
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

package httpapi

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/wbits/guitars/internal/guitarcollection/application"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/auth"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/persistence"
)

const validBearer = "Bearer test-secret"

type sequentialIDs struct {
	ids []string
	i   int
}

func (s *sequentialIDs) NewID() string {
	id := s.ids[s.i%len(s.ids)]
	s.i++
	return id
}

func newTestHandler() *Handler {
	repo := persistence.NewMemoryRepository()
	marketRepo := persistence.NewMemoryMarketLogRepository()
	ids := &sequentialIDs{ids: []string{"g-1", "g-2", "g-3", "ml-1", "ml-2", "ml-3"}}
	svc := application.NewService(repo, ids)
	marketLogs := application.NewMarketLogService(repo, marketRepo, ids)
	authn := auth.NewBearerAuthenticator(auth.TokenLoaderFunc(func(context.Context) (string, error) {
		return "test-secret", nil
	}), 0)
	return NewHandler(svc, marketLogs, authn, nil)
}

func reqWithAuth(method, path, body string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Path:       path,
		Headers:    map[string]string{"Authorization": validBearer},
		Body:       body,
	}
}

func validBody() string {
	return `{
		"serialNumber":"SN-1",
		"pictures":["https://example.com/a.jpg"],
		"description":"1996 sunburst",
		"brand":"Fender",
		"typeName":"Stratocaster",
		"buildYear":1996,
		"priceAmount":199900,
		"priceCurrency":"EUR"
	}`
}

func TestHandler_Unauthorized_WhenAuthHeaderMissing(t *testing.T) {
	h := newTestHandler()
	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/guitar",
	})
	if resp.StatusCode != 401 {
		t.Errorf("want 401, got %d (%s)", resp.StatusCode, resp.Body)
	}
}

func TestHandler_Unauthorized_WhenTokenWrong(t *testing.T) {
	h := newTestHandler()
	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/guitar",
		Headers:    map[string]string{"Authorization": "Bearer wrong"},
	})
	if resp.StatusCode != 401 {
		t.Errorf("want 401, got %d (%s)", resp.StatusCode, resp.Body)
	}
}

func TestHandler_PostGuitar_Creates(t *testing.T) {
	h := newTestHandler()
	resp, _ := h.Handle(context.Background(), reqWithAuth("POST", "/guitar", validBody()))
	if resp.StatusCode != 201 {
		t.Fatalf("want 201, got %d (%s)", resp.StatusCode, resp.Body)
	}
	var got guitarResponse
	if err := json.Unmarshal([]byte(resp.Body), &got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.ID != "g-1" || got.Brand != "Fender" || got.PriceAmount != 199900 {
		t.Errorf("unexpected response: %+v", got)
	}
}

func TestHandler_PostGuitar_400OnInvalidJSON(t *testing.T) {
	h := newTestHandler()
	resp, _ := h.Handle(context.Background(), reqWithAuth("POST", "/guitar", "{not-json"))
	if resp.StatusCode != 400 {
		t.Errorf("want 400, got %d", resp.StatusCode)
	}
}

func TestHandler_PostGuitar_400OnValidationError(t *testing.T) {
	h := newTestHandler()
	body := strings.Replace(validBody(), `"Fender"`, `""`, 1)
	resp, _ := h.Handle(context.Background(), reqWithAuth("POST", "/guitar", body))
	if resp.StatusCode != 400 {
		t.Errorf("want 400, got %d (%s)", resp.StatusCode, resp.Body)
	}
}

func TestHandler_GetGuitar_NotFound(t *testing.T) {
	h := newTestHandler()
	resp, _ := h.Handle(context.Background(), reqWithAuth("GET", "/guitar/missing", ""))
	if resp.StatusCode != 404 {
		t.Errorf("want 404, got %d", resp.StatusCode)
	}
}

func TestHandler_FullCRUDLifecycle(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	// Create
	resp, _ := h.Handle(ctx, reqWithAuth("POST", "/guitar", validBody()))
	if resp.StatusCode != 201 {
		t.Fatalf("create: want 201, got %d (%s)", resp.StatusCode, resp.Body)
	}
	var created guitarResponse
	_ = json.Unmarshal([]byte(resp.Body), &created)

	// List
	resp, _ = h.Handle(ctx, reqWithAuth("GET", "/guitar", ""))
	if resp.StatusCode != 200 {
		t.Fatalf("list: want 200, got %d (%s)", resp.StatusCode, resp.Body)
	}
	var list []guitarResponse
	_ = json.Unmarshal([]byte(resp.Body), &list)
	if len(list) != 1 {
		t.Errorf("list: expected 1 guitar, got %d", len(list))
	}

	// Get
	resp, _ = h.Handle(ctx, reqWithAuth("GET", "/guitar/"+created.ID, ""))
	if resp.StatusCode != 200 {
		t.Fatalf("get: want 200, got %d", resp.StatusCode)
	}

	// Update
	updatedBody := strings.Replace(validBody(), `"Fender"`, `"Gibson"`, 1)
	updatedBody = strings.Replace(updatedBody, `"Stratocaster"`, `"Les Paul"`, 1)
	resp, _ = h.Handle(ctx, reqWithAuth("PUT", "/guitar/"+created.ID, updatedBody))
	if resp.StatusCode != 200 {
		t.Fatalf("update: want 200, got %d (%s)", resp.StatusCode, resp.Body)
	}
	var updated guitarResponse
	_ = json.Unmarshal([]byte(resp.Body), &updated)
	if updated.Brand != "Gibson" || updated.TypeName != "Les Paul" {
		t.Errorf("update did not apply: %+v", updated)
	}

	// Delete
	resp, _ = h.Handle(ctx, reqWithAuth("DELETE", "/guitar/"+created.ID, ""))
	if resp.StatusCode != 204 {
		t.Fatalf("delete: want 204, got %d", resp.StatusCode)
	}

	// Deleting again -> 404
	resp, _ = h.Handle(ctx, reqWithAuth("DELETE", "/guitar/"+created.ID, ""))
	if resp.StatusCode != 404 {
		t.Errorf("second delete: want 404, got %d", resp.StatusCode)
	}
}

func TestHandler_PresignUpload_Returns503WhenNotConfigured(t *testing.T) {
	h := newTestHandler()
	resp, _ := h.Handle(context.Background(), reqWithAuth("POST", "/upload/presign", `{"contentType":"image/jpeg"}`))
	if resp.StatusCode != 503 {
		t.Errorf("want 503, got %d (%s)", resp.StatusCode, resp.Body)
	}
}

func TestHandler_UnknownRouteReturns404(t *testing.T) {
	h := newTestHandler()
	resp, _ := h.Handle(context.Background(), reqWithAuth("GET", "/banjo", ""))
	if resp.StatusCode != 404 {
		t.Errorf("want 404, got %d", resp.StatusCode)
	}
}

func TestHandler_AcceptsLowercaseAuthorizationHeader(t *testing.T) {
	h := newTestHandler()
	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/guitar",
		Headers:    map[string]string{"authorization": validBearer},
	})
	if resp.StatusCode != 200 {
		t.Errorf("want 200, got %d (%s)", resp.StatusCode, resp.Body)
	}
}

func TestHandler_OptionsPreflight_Returns204WithoutAuth(t *testing.T) {
	h := newTestHandler()
	resp, err := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "OPTIONS",
		Path:       "/guitar",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 204 {
		t.Errorf("want 204, got %d", resp.StatusCode)
	}
	if resp.Headers["Access-Control-Allow-Origin"] != "*" {
		t.Errorf("Allow-Origin: got %q, want *", resp.Headers["Access-Control-Allow-Origin"])
	}
	if !strings.Contains(resp.Headers["Access-Control-Allow-Headers"], "Authorization") {
		t.Errorf("Allow-Headers missing Authorization: %q", resp.Headers["Access-Control-Allow-Headers"])
	}
}

func TestHandler_Get_IncludesCORSHeaders(t *testing.T) {
	h := newTestHandler()
	resp, _ := h.Handle(context.Background(), reqWithAuth("GET", "/guitar", ""))
	if resp.Headers["Access-Control-Allow-Origin"] != "*" {
		t.Errorf("Allow-Origin: got %q, want *", resp.Headers["Access-Control-Allow-Origin"])
	}
}

func TestHandler_MarketLog_CreateAndList(t *testing.T) {
	h := newTestHandler()
	createResp, _ := h.Handle(context.Background(), reqWithAuth("POST", "/guitar", validBody()))
	if createResp.StatusCode != 201 {
		t.Fatalf("create guitar: want 201, got %d (%s)", createResp.StatusCode, createResp.Body)
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(createResp.Body), &created); err != nil {
		t.Fatalf("decode guitar: %v", err)
	}

	logBody := `[{
		"source":"reverb",
		"action":"for_sale",
		"priceAmount":150000,
		"priceCurrency":"EUR",
		"listingUrl":"https://reverb.com/item/1",
		"listingTitle":"Fender Strat"
	}]`
	postResp, _ := h.Handle(context.Background(), reqWithAuth("POST", "/guitar/"+created.ID+"/market-log", logBody))
	if postResp.StatusCode != 201 {
		t.Fatalf("create market log: want 201, got %d (%s)", postResp.StatusCode, postResp.Body)
	}

	listResp, _ := h.Handle(context.Background(), reqWithAuth("GET", "/guitar/"+created.ID+"/market-log", ""))
	if listResp.StatusCode != 200 {
		t.Fatalf("list market logs: want 200, got %d (%s)", listResp.StatusCode, listResp.Body)
	}
	var logs []struct {
		Source string `json:"source"`
		Action string `json:"action"`
	}
	if err := json.Unmarshal([]byte(listResp.Body), &logs); err != nil {
		t.Fatalf("decode logs: %v", err)
	}
	if len(logs) != 1 || logs[0].Source != "reverb" || logs[0].Action != "for_sale" {
		t.Fatalf("unexpected logs: %+v", logs)
	}
}

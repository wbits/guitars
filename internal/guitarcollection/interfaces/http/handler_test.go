package httpapi

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/wbits/guitars/internal/assistant"
	"github.com/wbits/guitars/internal/guitarcollection/application"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/auth"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/persistence"
	profileapp "github.com/wbits/guitars/internal/userprofile/application"
	profilepersistence "github.com/wbits/guitars/internal/userprofile/infrastructure/persistence"
	profilecrypto "github.com/wbits/guitars/internal/userprofile/infrastructure/crypto"
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

func testBYOKEncryptor() *profilecrypto.KeyStore {
	store, err := profilecrypto.NewKeyStoreFromBase64("MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDE=")
	if err != nil {
		panic(err)
	}
	return store
}

func newTestHandler() *Handler {
	repo := persistence.NewMemoryRepository()
	marketRepo := persistence.NewMemoryMarketLogRepository()
	profileRepo := profilepersistence.NewMemoryRepository()
	ids := &sequentialIDs{ids: []string{"g-1", "g-2", "g-3", "ml-1", "ml-2", "ml-3"}}
	svc := application.NewService(repo, ids)
	profiles := profileapp.NewService(profileRepo, testBYOKEncryptor())
	marketLogs := application.NewMarketLogService(repo, marketRepo, ids, nil, nil, profiles)
	authn := auth.NewBearerAuthenticator(auth.TokenLoaderFunc(func(context.Context) (string, error) {
		return "test-secret", nil
	}), 0)
	assistantSvc := assistant.NewService(
		svc,
		assistant.RuleLLM{},
		assistant.NewMemoryRateLimiter(100),
		&assistant.ProfileBYOKProvider{Profiles: profiles},
		assistant.NewMemoryRateLimiter(200),
	)
	return NewHandler(svc, marketLogs, profiles, authn, nil, "guitars-admins", assistantSvc)
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

func TestHandler_GetMe_ReturnsPrincipalUserID(t *testing.T) {
	h := newTestHandler()
	resp, _ := h.Handle(context.Background(), reqWithAuth("GET", "/me", ""))
	if resp.StatusCode != 200 {
		t.Fatalf("want 200, got %d (%s)", resp.StatusCode, resp.Body)
	}
	var got meResponse
	if err := json.Unmarshal([]byte(resp.Body), &got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.UserID != "local-dev-user" {
		t.Errorf("want local-dev-user, got %q", got.UserID)
	}
	if got.Email != "local-dev@example.com" {
		t.Errorf("want local-dev@example.com, got %q", got.Email)
	}
	if got.DisplayName != "local-dev@example.com" {
		t.Errorf("want displayName local-dev@example.com, got %q", got.DisplayName)
	}
}

func TestHandler_PatchMe_UpdatesUsername(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()
	patchResp, _ := h.Handle(ctx, reqWithAuth("PATCH", "/me", `{"username":"collector"}`))
	if patchResp.StatusCode != 200 {
		t.Fatalf("patch: want 200, got %d (%s)", patchResp.StatusCode, patchResp.Body)
	}
	var updated meResponse
	if err := json.Unmarshal([]byte(patchResp.Body), &updated); err != nil {
		t.Fatalf("decode patch body: %v", err)
	}
	if updated.Username != "collector" || updated.DisplayName != "collector" {
		t.Fatalf("unexpected profile: %+v", updated)
	}

	getResp, _ := h.Handle(ctx, reqWithAuth("GET", "/me", ""))
	if getResp.StatusCode != 200 {
		t.Fatalf("get: want 200, got %d (%s)", getResp.StatusCode, getResp.Body)
	}
	var got meResponse
	if err := json.Unmarshal([]byte(getResp.Body), &got); err != nil {
		t.Fatalf("decode get body: %v", err)
	}
	if got.Username != "collector" {
		t.Fatalf("want persisted username, got %+v", got)
	}
}

func TestHandler_ListCollections_ReturnsOwnersWithCounts(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()
	resp, _ := h.Handle(ctx, reqWithAuth("POST", "/guitar", validBody()))
	if resp.StatusCode != 201 {
		t.Fatalf("create: want 201, got %d (%s)", resp.StatusCode, resp.Body)
	}

	resp, _ = h.Handle(ctx, reqWithAuth("GET", "/collections", ""))
	if resp.StatusCode != 200 {
		t.Fatalf("list collections: want 200, got %d (%s)", resp.StatusCode, resp.Body)
	}
	var owners []collectionOwnerResponse
	if err := json.Unmarshal([]byte(resp.Body), &owners); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(owners) != 1 || owners[0].UserID != "local-dev-user" || owners[0].GuitarCount != 1 {
		t.Fatalf("unexpected owners: %+v", owners)
	}
	if owners[0].DisplayName != "local-dev@example.com" {
		t.Fatalf("unexpected displayName: %+v", owners[0])
	}
	if owners[0].MarketCrawlEnabled {
		t.Fatalf("market crawl should default to false: %+v", owners[0])
	}
}

func TestHandler_PatchCollectionMarketCrawl_RequiresAdmin(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()
	resp, _ := h.Handle(ctx, reqWithAuth("PATCH", "/collections/local-dev-user/market-crawl", `{"marketCrawlEnabled":true}`))
	if resp.StatusCode != 403 {
		t.Fatalf("non-admin patch: want 403, got %d (%s)", resp.StatusCode, resp.Body)
	}

	t.Setenv("LOCAL_DEV_ADMIN_GROUPS", "guitars-admins")
	resp, _ = h.Handle(ctx, reqWithAuth("PATCH", "/collections/local-dev-user/market-crawl", `{"marketCrawlEnabled":true}`))
	if resp.StatusCode != 200 {
		t.Fatalf("admin patch: want 200, got %d (%s)", resp.StatusCode, resp.Body)
	}
	var owner collectionOwnerResponse
	if err := json.Unmarshal([]byte(resp.Body), &owner); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !owner.MarketCrawlEnabled {
		t.Fatalf("expected market crawl enabled: %+v", owner)
	}
}

func TestHandler_DeleteCollectionMarketLogs_RequiresAdmin(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	createResp, _ := h.Handle(ctx, reqWithAuth("POST", "/guitar", validBody()))
	if createResp.StatusCode != 201 {
		t.Fatalf("create guitar: want 201, got %d (%s)", createResp.StatusCode, createResp.Body)
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(createResp.Body), &created); err != nil {
		t.Fatalf("decode guitar: %v", err)
	}

	logBody := `[{"source":"reverb","action":"for_sale","priceAmount":150000,"priceCurrency":"EUR"}]`
	postResp, _ := h.Handle(ctx, reqWithAuth("POST", "/guitar/"+created.ID+"/market-log", logBody))
	if postResp.StatusCode != 201 {
		t.Fatalf("create market log: want 201, got %d (%s)", postResp.StatusCode, postResp.Body)
	}

	resp, _ := h.Handle(ctx, reqWithAuth("DELETE", "/collections/local-dev-user/market-log", ""))
	if resp.StatusCode != 403 {
		t.Fatalf("non-admin delete: want 403, got %d (%s)", resp.StatusCode, resp.Body)
	}

	t.Setenv("LOCAL_DEV_ADMIN_GROUPS", "guitars-admins")
	resp, _ = h.Handle(ctx, reqWithAuth("DELETE", "/collections/local-dev-user/market-log", ""))
	if resp.StatusCode != 200 {
		t.Fatalf("admin delete: want 200, got %d (%s)", resp.StatusCode, resp.Body)
	}
	var out clearCollectionMarketLogsResponse
	if err := json.Unmarshal([]byte(resp.Body), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.DeletedCount != 1 {
		t.Fatalf("want 1 deleted log, got %d", out.DeletedCount)
	}

	listResp, _ := h.Handle(ctx, reqWithAuth("GET", "/guitar/"+created.ID+"/market-log", ""))
	if listResp.StatusCode != 200 {
		t.Fatalf("list market logs: want 200, got %d (%s)", listResp.StatusCode, listResp.Body)
	}
	if listResp.Body != "[]" && listResp.Body != "null" {
		t.Fatalf("expected empty logs, got %s", listResp.Body)
	}
}

func TestHandler_ListUserCollection_ReturnsOwnedGuitars(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()
	resp, _ := h.Handle(ctx, reqWithAuth("POST", "/guitar", validBody()))
	if resp.StatusCode != 201 {
		t.Fatalf("create: want 201, got %d (%s)", resp.StatusCode, resp.Body)
	}

	resp, _ = h.Handle(ctx, reqWithAuth("GET", "/collections/local-dev-user/guitar", ""))
	if resp.StatusCode != 200 {
		t.Fatalf("list user collection: want 200, got %d (%s)", resp.StatusCode, resp.Body)
	}
	var list []guitarResponse
	if err := json.Unmarshal([]byte(resp.Body), &list); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(list) != 1 || list[0].Brand != "Fender" {
		t.Fatalf("unexpected list: %+v", list)
	}
}

func TestHandler_GetGuitar_AllowsReadingOtherUsersGuitar(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()
	resp, _ := h.Handle(ctx, reqWithAuth("POST", "/guitar", validBody()))
	if resp.StatusCode != 201 {
		t.Fatalf("create: want 201, got %d (%s)", resp.StatusCode, resp.Body)
	}
	var created guitarResponse
	if err := json.Unmarshal([]byte(resp.Body), &created); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	resp, _ = h.Handle(ctx, reqWithAuth("GET", "/guitar/"+created.ID, ""))
	if resp.StatusCode != 200 {
		t.Fatalf("get guitar: want 200, got %d (%s)", resp.StatusCode, resp.Body)
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
	if !strings.Contains(resp.Headers["Access-Control-Allow-Methods"], "PATCH") {
		t.Errorf("Allow-Methods missing PATCH: %q", resp.Headers["Access-Control-Allow-Methods"])
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
		"listingTitle":"Fender Stratocaster"
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

func TestHandler_AssistantChat_FiltersCollection(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()
	createResp, _ := h.Handle(ctx, reqWithAuth("POST", "/guitar", validBody()))
	if createResp.StatusCode != 201 {
		t.Fatalf("create: %d %s", createResp.StatusCode, createResp.Body)
	}

	body := `{"collectionUserId":"local-dev-user","message":"Fender under 2000 euro"}`
	resp, _ := h.Handle(ctx, reqWithAuth("POST", "/assistant/chat", body))
	if resp.StatusCode != 200 {
		t.Fatalf("assistant chat: want 200, got %d (%s)", resp.StatusCode, resp.Body)
	}
	var out assistantChatResponse
	if err := json.Unmarshal([]byte(resp.Body), &out); err != nil {
		t.Fatal(err)
	}
	if out.Message == "" {
		t.Fatal("expected message")
	}
	if len(out.MatchingIDs) != 1 {
		t.Fatalf("matching ids: %+v message: %q filter: %+v body: %s", out.MatchingIDs, out.Message, out.Filter, resp.Body)
	}
}

func TestHandler_AssistantBYOK_PutAndDelete(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	putResp, _ := h.Handle(ctx, reqWithAuth("PUT", "/me/assistant-byok", `{"apiKey":"sk-owner","model":"gpt-test"}`))
	if putResp.StatusCode != 200 {
		t.Fatalf("put BYOK: want 200, got %d (%s)", putResp.StatusCode, putResp.Body)
	}
	var me meResponse
	if err := json.Unmarshal([]byte(putResp.Body), &me); err != nil {
		t.Fatal(err)
	}
	if !me.AssistantByokConfigured || me.AssistantLlmModel != "gpt-test" {
		t.Fatalf("unexpected me: %+v", me)
	}

	delResp, _ := h.Handle(ctx, reqWithAuth("DELETE", "/me/assistant-byok", ""))
	if delResp.StatusCode != 200 {
		t.Fatalf("delete BYOK: want 200, got %d (%s)", delResp.StatusCode, delResp.Body)
	}
	if err := json.Unmarshal([]byte(delResp.Body), &me); err != nil {
		t.Fatal(err)
	}
	if me.AssistantByokConfigured {
		t.Fatalf("expected cleared: %+v", me)
	}
}

// Package httpapi exposes the GuitarCollection application service over the
// API Gateway proxy integration. Routing is done manually (no extra
// dependency) because the surface is very small: five endpoints.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	"github.com/wbits/guitars/internal/assistant"
	"github.com/wbits/guitars/internal/guitaranalysis"
	"github.com/wbits/guitars/internal/guitarcollection/application"
	"github.com/wbits/guitars/internal/guitarcollection/domain"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/auth"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/storage"
	profileapp "github.com/wbits/guitars/internal/userprofile/application"
	profiledomain "github.com/wbits/guitars/internal/userprofile/domain"
)

// Handler routes API Gateway proxy requests to the GuitarCollection
// application service. It owns:
//   - authentication (delegated to auth.Authenticator),
//   - request/response (de)serialisation,
//   - mapping domain/application errors to HTTP status codes.
//
// It deliberately holds no business rules of its own.
type Handler struct {
	svc        *application.Service
	marketLogs *application.MarketLogService
	profiles   *profileapp.Service
	assistant  *assistant.Service
	analysis   *guitaranalysis.Service
	auth       auth.Authenticator
	presigner  *storage.Presigner
	adminGroup string
}

// NewHandler constructs a Handler wired to the supplied services and
// authenticator. Both service arguments are required. presigner may be nil
// when image uploads are not configured. profiles may be nil to disable
// profile endpoints. adminGroup names the Cognito group that grants admin access.
func NewHandler(svc *application.Service, marketLogs *application.MarketLogService, profiles *profileapp.Service, a auth.Authenticator, presigner *storage.Presigner, adminGroup string, assistantSvc *assistant.Service, analysisSvc *guitaranalysis.Service) *Handler {
	return &Handler{svc: svc, marketLogs: marketLogs, profiles: profiles, assistant: assistantSvc, analysis: analysisSvc, auth: a, presigner: presigner, adminGroup: adminGroup}
}

var guitarItemPath = regexp.MustCompile(`^/guitar/([^/]+)/?$`)
var guitarAnalyzePath = regexp.MustCompile(`^/guitar/([^/]+)/analyze/?$`)
var guitarMarketLogPath = regexp.MustCompile(`^/guitar/([^/]+)/market-log/?$`)
var userCollectionPath = regexp.MustCompile(`^/collections/([^/]+)/guitar/?$`)
var collectionMarketCrawlPath = regexp.MustCompile(`^/collections/([^/]+)/market-crawl/?$`)
var collectionMarketLogPath = regexp.MustCompile(`^/collections/([^/]+)/market-log/?$`)

// Handle is the entrypoint suitable for lambda.Start.
func (h *Handler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if strings.EqualFold(req.HTTPMethod, "OPTIONS") {
		return corsPreflightResponse(), nil
	}

	principal, err := h.auth.Authenticate(ctx, pickHeader(req.Headers, "Authorization"))
	if err != nil {
		if errors.Is(err, auth.ErrUnauthorized) {
			return jsonResponse(401, errorResponse{Error: "unauthorized"})
		}
		return jsonResponse(500, errorResponse{Error: "authentication failure"})
	}

	method := strings.ToUpper(req.HTTPMethod)
	path := normalisePath(req.Path)

	switch {
	case path == "/me" && method == "GET":
		return h.me(ctx, principal)
	case path == "/me" && method == "PATCH":
		return h.patchMe(ctx, principal, req.Body)
	case path == "/me/assistant-byok" && method == "PUT":
		return h.putAssistantBYOK(ctx, principal, req.Body)
	case path == "/me/assistant-byok" && method == "DELETE":
		return h.deleteAssistantBYOK(ctx, principal)
	case path == "/me/reanalyze-collection" && method == "POST":
		return h.reanalyzeCollection(ctx, principal)
	case path == "/collections" && method == "GET":
		return h.listCollections(ctx)
	case path == "/guitar" && method == "GET":
		return h.list(ctx, principal.UserID)
	case path == "/guitar" && method == "POST":
		return h.create(ctx, principal.UserID, req.Body)
	case path == "/upload/presign" && method == "POST":
		return h.presignUpload(ctx, req.Body)
	case path == "/assistant/chat" && method == "POST":
		return h.assistantChat(ctx, principal, req.Body)
	default:
		if m := collectionMarketCrawlPath.FindStringSubmatch(path); m != nil {
			if method == "PATCH" {
				return h.patchCollectionMarketCrawl(ctx, principal, m[1], req.Body)
			}
		}
		if m := collectionMarketLogPath.FindStringSubmatch(path); m != nil {
			if method == "DELETE" {
				return h.deleteCollectionMarketLogs(ctx, principal, m[1])
			}
		}
		if m := userCollectionPath.FindStringSubmatch(path); m != nil {
			if method == "GET" {
				return h.listUserCollection(ctx, m[1])
			}
		}
		if m := guitarMarketLogPath.FindStringSubmatch(path); m != nil {
			id := m[1]
			switch method {
			case "GET":
				return h.listMarketLogs(ctx, principal.UserID, id)
			case "POST":
				return h.createMarketLogs(ctx, principal.UserID, principal.Email, id, req.Body)
			}
		}
		if m := guitarAnalyzePath.FindStringSubmatch(path); m != nil && method == "POST" {
			return h.analyzeGuitar(ctx, principal.UserID, m[1])
		}
		if m := guitarItemPath.FindStringSubmatch(path); m != nil {
			id := m[1]
			switch method {
			case "GET":
				return h.get(ctx, principal.UserID, id)
			case "PUT":
				return h.update(ctx, principal.UserID, id, req.Body)
			case "DELETE":
				return h.delete(ctx, principal.UserID, id)
			}
		}
		// PUT/DELETE without an id are not part of the contract but produce
		// the same useful 405 response.
		if path == "/guitar" && (method == "PUT" || method == "DELETE") {
			return jsonResponse(400, errorResponse{Error: "guitar id is required in the path"})
		}
	}
	return jsonResponse(404, errorResponse{Error: "not found"})
}

func (h *Handler) me(ctx context.Context, principal auth.Principal) (events.APIGatewayProxyResponse, error) {
	isAdmin := auth.IsAdmin(principal, h.adminGroup)
	if h.profiles == nil {
		return jsonResponse(200, meResponse{UserID: principal.UserID, Email: principal.Email, DisplayName: profileapp.DisplayNameForUser(principal.UserID, nil), IsAdmin: isAdmin})
	}
	profile, err := h.profiles.GetProfile(ctx, principal.UserID, principal.Email)
	if err != nil {
		return profileErrorToResponse(err)
	}
	resp := toMeResponse(profile, isAdmin)
	configured, usable, err := h.profiles.AssistantBYOKMeStatus(ctx, principal.UserID)
	if err != nil {
		return profileErrorToResponse(err)
	}
	resp.AssistantByokConfigured = usable
	resp.AssistantByokNeedsResave = configured && !usable
	return jsonResponse(200, resp)
}

func (h *Handler) patchMe(ctx context.Context, principal auth.Principal, body string) (events.APIGatewayProxyResponse, error) {
	if h.profiles == nil {
		return jsonResponse(503, errorResponse{Error: "profiles are not configured"})
	}
	var req profilePatchRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return jsonResponse(400, errorResponse{Error: "invalid JSON body"})
	}
	profile, err := h.profiles.UpdateUsername(ctx, principal.UserID, principal.Email, req.Username)
	if err != nil {
		return profileErrorToResponse(err)
	}
	if req.PhotoAnalysisEnabled != nil {
		profile, err = h.profiles.SetPhotoAnalysisEnabled(ctx, principal.UserID, principal.Email, *req.PhotoAnalysisEnabled)
		if err != nil {
			return profileErrorToResponse(err)
		}
	}
	return jsonResponse(200, toMeResponse(profile, auth.IsAdmin(principal, h.adminGroup)))
}

func (h *Handler) patchCollectionMarketCrawl(ctx context.Context, principal auth.Principal, userID, body string) (events.APIGatewayProxyResponse, error) {
	if !auth.IsAdmin(principal, h.adminGroup) {
		return jsonResponse(403, errorResponse{Error: "forbidden"})
	}
	if h.profiles == nil {
		return jsonResponse(503, errorResponse{Error: "profiles are not configured"})
	}
	var req collectionMarketCrawlPatchRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return jsonResponse(400, errorResponse{Error: "invalid JSON body"})
	}
	profile, err := h.profiles.SetMarketCrawlEnabled(ctx, userID, req.MarketCrawlEnabled)
	if err != nil {
		return profileErrorToResponse(err)
	}
	guitars, err := h.svc.ListUserGuitars(ctx, userID)
	if err != nil {
		return errorToResponse(err)
	}
	return jsonResponse(200, toCollectionOwnerResponse(userID, profile, len(guitars)))
}

func (h *Handler) listCollections(ctx context.Context) (events.APIGatewayProxyResponse, error) {
	owners, err := h.svc.ListCollectionOwners(ctx)
	if err != nil {
		return errorToResponse(err)
	}
	var profileRecords map[string]*profiledomain.Profile
	if h.profiles != nil {
		profileRecords, err = h.profiles.GetProfilesByUserIDs(ctx, owners)
		if err != nil {
			return profileErrorToResponse(err)
		}
	}
	out := make([]collectionOwnerResponse, 0, len(owners))
	for _, ownerID := range owners {
		guitars, err := h.svc.ListUserGuitars(ctx, ownerID)
		if err != nil {
			return errorToResponse(err)
		}
		resp := collectionOwnerResponse{
			UserID:      ownerID,
			GuitarCount: len(guitars),
		}
		if profile, ok := profileRecords[ownerID]; ok {
			resp.Username = profile.Username()
			resp.Email = profile.Email()
			resp.DisplayName = profile.DisplayName()
			resp.MarketCrawlEnabled = profile.MarketCrawlEnabled()
		} else {
			resp.DisplayName = profileapp.DisplayNameForUser(ownerID, nil)
		}
		out = append(out, resp)
	}
	return jsonResponse(200, out)
}

func (h *Handler) listUserCollection(ctx context.Context, userID string) (events.APIGatewayProxyResponse, error) {
	guitars, err := h.svc.ListUserGuitars(ctx, userID)
	if err != nil {
		return errorToResponse(err)
	}
	out := h.toResponsesWithAnalysis(ctx, guitars)
	return jsonResponse(200, out)
}

func (h *Handler) list(ctx context.Context, ownerID string) (events.APIGatewayProxyResponse, error) {
	guitars, err := h.svc.ListGuitars(ctx, ownerID)
	if err != nil {
		return errorToResponse(err)
	}
	out := h.toResponsesWithAnalysis(ctx, guitars)
	return jsonResponse(200, out)
}

func (h *Handler) get(ctx context.Context, ownerID, id string) (events.APIGatewayProxyResponse, error) {
	g, err := h.svc.GetGuitar(ctx, ownerID, id)
	if err != nil {
		return errorToResponse(err)
	}
	return jsonResponse(200, h.toResponseWithAnalysis(ctx, g))
}

func (h *Handler) create(ctx context.Context, ownerID, body string) (events.APIGatewayProxyResponse, error) {
	var req guitarRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return jsonResponse(400, errorResponse{Error: "invalid JSON body"})
	}
	g, err := h.svc.AddGuitar(ctx, ownerID, requestToInput(req))
	if err != nil {
		return errorToResponse(err)
	}
	h.triggerAnalysis(ctx, g)
	return jsonResponse(201, h.toResponseWithAnalysis(ctx, g))
}

func (h *Handler) update(ctx context.Context, ownerID, id, body string) (events.APIGatewayProxyResponse, error) {
	var req guitarRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return jsonResponse(400, errorResponse{Error: "invalid JSON body"})
	}
	g, err := h.svc.UpdateGuitar(ctx, ownerID, id, requestToInput(req))
	if err != nil {
		return errorToResponse(err)
	}
	h.triggerAnalysis(ctx, g)
	return jsonResponse(200, h.toResponseWithAnalysis(ctx, g))
}

func (h *Handler) delete(ctx context.Context, ownerID, id string) (events.APIGatewayProxyResponse, error) {
	if err := h.svc.DeleteGuitar(ctx, ownerID, id); err != nil {
		return errorToResponse(err)
	}
	if h.analysis != nil {
		_ = h.analysis.DeleteForGuitar(ctx, id)
	}
	return events.APIGatewayProxyResponse{StatusCode: 204, Headers: responseHeaders(nil)}, nil
}

func (h *Handler) analyzeGuitar(ctx context.Context, ownerID, id string) (events.APIGatewayProxyResponse, error) {
	g, err := h.svc.GetGuitar(ctx, ownerID, id)
	if err != nil {
		return errorToResponse(err)
	}
	if g.Owner() != ownerID {
		return jsonResponse(403, errorResponse{Error: "forbidden"})
	}
	if h.analysis == nil {
		return jsonResponse(503, errorResponse{Error: "photo analysis is not configured"})
	}
	if _, err := h.analysis.Reanalyze(ctx, g); err != nil {
		return analysisErrorToResponse(err)
	}
	return jsonResponse(200, h.toResponseWithAnalysis(ctx, g))
}

func (h *Handler) reanalyzeCollection(ctx context.Context, principal auth.Principal) (events.APIGatewayProxyResponse, error) {
	if h.analysis == nil {
		return jsonResponse(503, errorResponse{Error: "photo analysis is not configured"})
	}
	guitars, err := h.svc.ListGuitars(ctx, principal.UserID)
	if err != nil {
		return errorToResponse(err)
	}
	result, err := h.analysis.ReanalyzeCollection(ctx, principal.UserID, guitars)
	if err != nil {
		return analysisErrorToResponse(err)
	}
	return jsonResponse(200, result)
}

func (h *Handler) triggerAnalysis(ctx context.Context, g *domain.Guitar) {
	if h.analysis == nil || g == nil {
		return
	}
	_, _ = h.analysis.ScheduleIfEligible(ctx, g)
}

func (h *Handler) toResponseWithAnalysis(ctx context.Context, g *domain.Guitar) guitarResponse {
	resp := toResponse(g)
	if h.analysis == nil {
		return resp
	}
	rec, err := h.analysis.Get(ctx, g.ID())
	if err != nil || rec == nil {
		return resp
	}
	resp.Analysis = toAnalysisResponse(rec)
	return resp
}

func (h *Handler) toResponsesWithAnalysis(ctx context.Context, guitars []*domain.Guitar) []guitarResponse {
	out := make([]guitarResponse, 0, len(guitars))
	if len(guitars) == 0 {
		return out
	}
	analysisMap := map[string]*guitaranalysis.Record{}
	if h.analysis != nil {
		ids := make([]string, len(guitars))
		for i, g := range guitars {
			ids[i] = g.ID()
		}
		var err error
		analysisMap, err = h.analysis.MapForGuitars(ctx, ids)
		if err != nil {
			analysisMap = map[string]*guitaranalysis.Record{}
		}
	}
	for _, g := range guitars {
		resp := toResponse(g)
		if rec := analysisMap[g.ID()]; rec != nil {
			resp.Analysis = toAnalysisResponse(rec)
		}
		out = append(out, resp)
	}
	return out
}

func (h *Handler) presignUpload(ctx context.Context, body string) (events.APIGatewayProxyResponse, error) {
	if h.presigner == nil {
		return jsonResponse(503, errorResponse{Error: "image uploads are not configured"})
	}

	var req presignUploadRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return jsonResponse(400, errorResponse{Error: "invalid JSON body"})
	}
	if strings.TrimSpace(req.ContentType) == "" {
		return jsonResponse(400, errorResponse{Error: "contentType is required"})
	}

	var result *storage.PresignResult
	var err error
	if strings.EqualFold(strings.TrimSpace(req.Purpose), "market-log") {
		result, err = h.presigner.PresignMarketLogImage(ctx, req.ContentType)
	} else {
		result, err = h.presigner.PresignPut(ctx, req.ContentType)
	}
	if err != nil {
		return jsonResponse(400, errorResponse{Error: err.Error()})
	}

	return jsonResponse(200, presignUploadResponse{
		UploadURL: result.UploadURL,
		PublicURL: result.PublicURL,
		Key:       result.Key,
	})
}

// --- helpers ---------------------------------------------------------------

func requestToInput(r guitarRequest) application.GuitarInput {
	return application.GuitarInput{
		SerialNumber:      r.SerialNumber,
		Color:             r.Color,
		Country:           r.Country,
		Factory:           r.Factory,
		Pictures:          r.Pictures,
		CoverPictureIndex: r.CoverPictureIndex,
		Description:       r.Description,
		Brand:             r.Brand,
		TypeName:          r.TypeName,
		BuildYear:         r.BuildYear,
		PriceAmount:       r.PriceAmount,
		PriceCurrency:     r.PriceCurrency,
	}
}

func errorToResponse(err error) (events.APIGatewayProxyResponse, error) {
	switch {
	case errors.Is(err, domain.ErrGuitarNotFound):
		return jsonResponse(404, errorResponse{Error: "guitar not found"})
	case domain.IsValidationError(err):
		return jsonResponse(400, errorResponse{Error: err.Error()})
	default:
		return jsonResponse(500, errorResponse{Error: "internal server error"})
	}
}

func profileErrorToResponse(err error) (events.APIGatewayProxyResponse, error) {
	switch {
	case profileapp.IsUsernameTaken(err):
		return jsonResponse(409, errorResponse{Error: "username is already taken"})
	case profileapp.IsValidationError(err):
		return jsonResponse(400, errorResponse{Error: err.Error()})
	case profileapp.IsBYOKNotConfigured(err):
		return jsonResponse(503, errorResponse{Error: "assistant BYOK is not configured on the server"})
	case profileapp.IsBYOKDecryptFailed(err):
		return jsonResponse(400, errorResponse{Error: "re-save your assistant API key in profile settings"})
	default:
		return jsonResponse(500, errorResponse{Error: "internal server error"})
	}
}

func analysisErrorToResponse(err error) (events.APIGatewayProxyResponse, error) {
	switch {
	case errors.Is(err, guitaranalysis.ErrBYOKNotConfigured):
		return jsonResponse(400, errorResponse{Error: "configure an assistant API key before re-analyzing"})
	case profileapp.IsBYOKDecryptFailed(err):
		return jsonResponse(400, errorResponse{Error: "re-save your assistant API key in profile settings"})
	case guitaranalysis.IsValidationError(err):
		return jsonResponse(400, errorResponse{Error: err.Error()})
	default:
		return jsonResponse(502, errorResponse{Error: "photo analysis failed"})
	}
}

func jsonResponse(status int, body interface{}) (events.APIGatewayProxyResponse, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    responseHeaders(map[string]string{"Content-Type": "application/json"}),
			Body:       `{"error":"failed to encode response"}`,
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers:    responseHeaders(map[string]string{"Content-Type": "application/json"}),
		Body:       string(payload),
	}, nil
}

func responseHeaders(extra map[string]string) map[string]string {
	return mergeHeaders(extra)
}

func normalisePath(p string) string {
	p = strings.TrimRight(p, "/")
	if p == "" {
		return "/"
	}
	return p
}

// pickHeader fetches the first match for name from headers using a
// case-insensitive comparison (API Gateway is inconsistent about casing).
func pickHeader(headers map[string]string, name string) string {
	if v, ok := headers[name]; ok {
		return v
	}
	lower := strings.ToLower(name)
	for k, v := range headers {
		if strings.ToLower(k) == lower {
			return v
		}
	}
	return ""
}

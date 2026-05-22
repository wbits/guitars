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

	"github.com/wbits/guitars/internal/guitarcollection/application"
	"github.com/wbits/guitars/internal/guitarcollection/domain"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/auth"
)

// Handler routes API Gateway proxy requests to the GuitarCollection
// application service. It owns:
//   - authentication (delegated to auth.Authenticator),
//   - request/response (de)serialisation,
//   - mapping domain/application errors to HTTP status codes.
//
// It deliberately holds no business rules of its own.
type Handler struct {
	svc  *application.Service
	auth auth.Authenticator
}

// NewHandler constructs a Handler wired to the supplied service and
// authenticator. Both arguments are required.
func NewHandler(svc *application.Service, a auth.Authenticator) *Handler {
	return &Handler{svc: svc, auth: a}
}

var guitarItemPath = regexp.MustCompile(`^/guitar/([^/]+)/?$`)

// Handle is the entrypoint suitable for lambda.Start.
func (h *Handler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if strings.EqualFold(req.HTTPMethod, "OPTIONS") {
		return corsPreflightResponse(), nil
	}

	if err := h.auth.Authenticate(ctx, pickHeader(req.Headers, "Authorization")); err != nil {
		if errors.Is(err, auth.ErrUnauthorized) {
			return jsonResponse(401, errorResponse{Error: "unauthorized"})
		}
		return jsonResponse(500, errorResponse{Error: "authentication failure"})
	}

	method := strings.ToUpper(req.HTTPMethod)
	path := normalisePath(req.Path)

	switch {
	case path == "/guitar" && method == "GET":
		return h.list(ctx)
	case path == "/guitar" && method == "POST":
		return h.create(ctx, req.Body)
	default:
		if m := guitarItemPath.FindStringSubmatch(path); m != nil {
			id := m[1]
			switch method {
			case "GET":
				return h.get(ctx, id)
			case "PUT":
				return h.update(ctx, id, req.Body)
			case "DELETE":
				return h.delete(ctx, id)
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

func (h *Handler) list(ctx context.Context) (events.APIGatewayProxyResponse, error) {
	guitars, err := h.svc.ListGuitars(ctx)
	if err != nil {
		return errorToResponse(err)
	}
	out := make([]guitarResponse, 0, len(guitars))
	for _, g := range guitars {
		out = append(out, toResponse(g))
	}
	return jsonResponse(200, out)
}

func (h *Handler) get(ctx context.Context, id string) (events.APIGatewayProxyResponse, error) {
	g, err := h.svc.GetGuitar(ctx, id)
	if err != nil {
		return errorToResponse(err)
	}
	return jsonResponse(200, toResponse(g))
}

func (h *Handler) create(ctx context.Context, body string) (events.APIGatewayProxyResponse, error) {
	var req guitarRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return jsonResponse(400, errorResponse{Error: "invalid JSON body"})
	}
	g, err := h.svc.AddGuitar(ctx, requestToInput(req))
	if err != nil {
		return errorToResponse(err)
	}
	return jsonResponse(201, toResponse(g))
}

func (h *Handler) update(ctx context.Context, id, body string) (events.APIGatewayProxyResponse, error) {
	var req guitarRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return jsonResponse(400, errorResponse{Error: "invalid JSON body"})
	}
	g, err := h.svc.UpdateGuitar(ctx, id, requestToInput(req))
	if err != nil {
		return errorToResponse(err)
	}
	return jsonResponse(200, toResponse(g))
}

func (h *Handler) delete(ctx context.Context, id string) (events.APIGatewayProxyResponse, error) {
	if err := h.svc.DeleteGuitar(ctx, id); err != nil {
		return errorToResponse(err)
	}
	return events.APIGatewayProxyResponse{StatusCode: 204, Headers: responseHeaders(nil)}, nil
}

// --- helpers ---------------------------------------------------------------

func requestToInput(r guitarRequest) application.GuitarInput {
	return application.GuitarInput{
		SerialNumber:  r.SerialNumber,
		Pictures:      r.Pictures,
		Description:   r.Description,
		Brand:         r.Brand,
		TypeName:      r.TypeName,
		BuildYear:     r.BuildYear,
		PriceAmount:   r.PriceAmount,
		PriceCurrency: r.PriceCurrency,
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

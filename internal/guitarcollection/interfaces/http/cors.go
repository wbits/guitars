package httpapi

import (
	"os"

	"github.com/aws/aws-lambda-go/events"
)

func corsAllowOrigin() string {
	if v := os.Getenv("CORS_ALLOW_ORIGIN"); v != "" {
		return v
	}
	return "*"
}

func corsHeaders() map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin":  corsAllowOrigin(),
		"Access-Control-Allow-Headers": "Content-Type,Authorization,Accept",
		"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
	}
}

func mergeHeaders(base map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(corsHeaders()))
	for k, v := range corsHeaders() {
		out[k] = v
	}
	for k, v := range base {
		out[k] = v
	}
	return out
}

func corsPreflightResponse() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: 204,
		Headers:    corsHeaders(),
	}
}

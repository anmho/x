package auth

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/anmhela/x/mcp/internal/keys"
)

// NewInterceptor returns a connect.UnaryInterceptorFunc that validates
// Authorization: Bearer <key> or X-Api-Key: <key> headers.
// storePath is the path to the keys file.
func NewInterceptor(storePath string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			key := extractKey(req)
			if key == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing API key"))
			}
			if !keys.Validate(storePath, key) {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid API key"))
			}
			return next(ctx, req)
		}
	}
}

func extractKey(req connect.AnyRequest) string {
	h := req.Header()

	// Try Authorization: Bearer <key>
	auth := h.Get("Authorization")
	if auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
		// bare token without Bearer prefix
		return auth
	}

	// Try X-Api-Key header
	if apiKey := h.Get("X-Api-Key"); apiKey != "" {
		return apiKey
	}

	return ""
}

package handler

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	mcpv1 "github.com/anmhela/x/mcp/internal/rpc/gen/mcp/v1"
	"github.com/anmhela/x/mcp/internal/keys"
)

// AdminHandler implements McpAdminServiceHandler.
type AdminHandler struct {
	storePath string
}

// NewAdminHandler creates an AdminHandler that uses the given key store path.
func NewAdminHandler(storePath string) *AdminHandler {
	return &AdminHandler{storePath: storePath}
}

// GenerateKey creates a new API key and returns it.
func (h *AdminHandler) GenerateKey(
	_ context.Context,
	req *connect.Request[mcpv1.GenerateKeyRequest],
) (*connect.Response[mcpv1.GenerateKeyResponse], error) {
	name := req.Msg.Name
	if name == "" {
		name = "default"
	}
	rec, err := keys.Generate(h.storePath, name)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("generate key: %w", err))
	}
	return connect.NewResponse(&mcpv1.GenerateKeyResponse{
		Key: rec.Key,
		Metadata: &mcpv1.ApiKey{
			Id:        rec.ID,
			Name:      rec.Name,
			CreatedAt: rec.CreatedAt,
		},
	}), nil
}

// ListKeys returns all API key metadata (no key values).
func (h *AdminHandler) ListKeys(
	_ context.Context,
	_ *connect.Request[mcpv1.ListKeysRequest],
) (*connect.Response[mcpv1.ListKeysResponse], error) {
	records, err := keys.Load(h.storePath)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list keys: %w", err))
	}
	apiKeys := make([]*mcpv1.ApiKey, 0, len(records))
	for _, r := range records {
		apiKeys = append(apiKeys, &mcpv1.ApiKey{
			Id:        r.ID,
			Name:      r.Name,
			CreatedAt: r.CreatedAt,
		})
	}
	return connect.NewResponse(&mcpv1.ListKeysResponse{Keys: apiKeys}), nil
}

// RevokeKey removes a key by ID.
func (h *AdminHandler) RevokeKey(
	_ context.Context,
	req *connect.Request[mcpv1.RevokeKeyRequest],
) (*connect.Response[mcpv1.RevokeKeyResponse], error) {
	found, err := keys.Revoke(h.storePath, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("revoke key: %w", err))
	}
	return connect.NewResponse(&mcpv1.RevokeKeyResponse{Success: found}), nil
}

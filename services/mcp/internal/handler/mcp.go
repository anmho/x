package handler

import (
	"context"

	"connectrpc.com/connect"
	mcpv1 "github.com/anmhela/x/mcp/internal/rpc/gen/mcp/v1"
	"github.com/anmhela/x/mcp/internal/tools"
)

// McpHandler implements McpServiceHandler.
type McpHandler struct {
	registry *tools.Registry
}

// NewMcpHandler creates a new McpHandler backed by the given tool registry.
func NewMcpHandler(reg *tools.Registry) *McpHandler {
	return &McpHandler{registry: reg}
}

// ListTools returns all available tools.
func (h *McpHandler) ListTools(
	_ context.Context,
	_ *connect.Request[mcpv1.ListToolsRequest],
) (*connect.Response[mcpv1.ListToolsResponse], error) {
	toolList := h.registry.Tools()
	defs := make([]*mcpv1.ToolDefinition, 0, len(toolList))
	for _, t := range toolList {
		defs = append(defs, &mcpv1.ToolDefinition{
			Name:            t.Name,
			Description:     t.Description,
			InputSchemaJson: t.InputSchemaJSON,
		})
	}
	return connect.NewResponse(&mcpv1.ListToolsResponse{Tools: defs}), nil
}

// CallTool invokes a tool by name.
func (h *McpHandler) CallTool(
	_ context.Context,
	req *connect.Request[mcpv1.CallToolRequest],
) (*connect.Response[mcpv1.CallToolResponse], error) {
	result, isError, err := h.registry.Call(req.Msg.Name, req.Msg.ArgumentsJson)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&mcpv1.CallToolResponse{
		Result:  result,
		IsError: isError,
	}), nil
}

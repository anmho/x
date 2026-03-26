package jsonrpc

import (
	"encoding/json"
	"net/http"

	"github.com/anmhela/x/mcp/internal/tools"
)

// Handler is an http.Handler that implements MCP JSON-RPC 2.0 passthrough.
type Handler struct {
	registry *tools.Registry
}

// NewHandler creates a new JSON-RPC Handler backed by the given registry.
func NewHandler(reg *tools.Registry) *Handler {
	return &Handler{registry: reg}
}

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type jsonRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req jsonRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, nil, -32700, "parse error")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch req.Method {
	case "initialize":
		writeResult(w, req.ID, map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "x-mcp",
				"version": "1.0.0",
			},
		})

	case "ping":
		writeResult(w, req.ID, map[string]interface{}{})

	case "tools/list":
		toolList := h.registry.Tools()
		items := make([]map[string]interface{}, 0, len(toolList))
		for _, t := range toolList {
			var inputSchema interface{}
			if err := json.Unmarshal([]byte(t.InputSchemaJSON), &inputSchema); err != nil {
				inputSchema = map[string]interface{}{"type": "object"}
			}
			items = append(items, map[string]interface{}{
				"name":        t.Name,
				"description": t.Description,
				"inputSchema": inputSchema,
			})
		}
		writeResult(w, req.ID, map[string]interface{}{"tools": items})

	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			writeError(w, req.ID, -32602, "invalid params")
			return
		}
		argsJSON := ""
		if len(params.Arguments) > 0 {
			argsJSON = string(params.Arguments)
		}
		result, isError, err := h.registry.Call(params.Name, argsJSON)
		if err != nil {
			writeError(w, req.ID, -32603, err.Error())
			return
		}
		content := []map[string]interface{}{
			{"type": "text", "text": result},
		}
		writeResult(w, req.ID, map[string]interface{}{
			"content": content,
			"isError": isError,
		})

	default:
		writeError(w, req.ID, -32601, "method not found")
	}
}

func writeResult(w http.ResponseWriter, id interface{}, result interface{}) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, id interface{}, code int, message string) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: message},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

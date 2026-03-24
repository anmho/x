package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"connectrpc.com/connect"
	"github.com/anmhela/x/mcp/internal/auth"
	"github.com/anmhela/x/mcp/internal/handler"
	"github.com/anmhela/x/mcp/internal/jsonrpc"
	"github.com/anmhela/x/mcp/internal/keys"
	mcpv1connect "github.com/anmhela/x/mcp/internal/rpc/gen/mcp/v1/mcpv1connect"
	"github.com/anmhela/x/mcp/internal/tools"
	"github.com/anmhela/x/mcp/internal/watch"
)

func main() {
	port := getEnv("MCP_PORT", "8765")
	storePath := keys.DefaultStorePath()
	if p := os.Getenv("MCP_KEYS_FILE"); p != "" {
		storePath = p
	}

	// Find repo root (walk up from executable or use MCP_ROOT env)
	root := getEnv("MCP_ROOT", repoRoot())

	// Auto-generate default key if store is empty
	existing, _ := keys.Load(storePath)
	if len(existing) == 0 {
		rec, err := keys.Generate(storePath, "default")
		if err == nil {
			fmt.Printf("[mcp] Generated API key: %s\n", rec.Key)
			fmt.Printf("[mcp] Saved to: %s\n", storePath)
		}
	}

	// Build tool registry and handlers
	reg := tools.NewRegistry(root)
	authInterceptor := auth.NewInterceptor(storePath)

	mux := http.NewServeMux()

	// ConnectRPC endpoints
	path, h := mcpv1connect.NewMcpServiceHandler(
		handler.NewMcpHandler(reg),
		connect.WithInterceptors(authInterceptor),
	)
	mux.Handle(path, h)

	adminPath, adminH := mcpv1connect.NewMcpAdminServiceHandler(
		handler.NewAdminHandler(storePath),
		connect.WithInterceptors(authInterceptor),
	)
	mux.Handle(adminPath, adminH)

	// MCP JSON-RPC passthrough (for Claude Code)
	jrpcHandler := jsonrpc.NewHandler(reg)
	mux.Handle("/mcp", authMiddleware(storePath, jrpcHandler))
	mux.Handle("/mailbox/events", authMiddleware(storePath, watch.NewMailboxEventsHandler(reg.CollabStore())))
	mux.Handle("/health", http.HandlerFunc(healthHandler))

	addr := ":" + port
	fmt.Printf("[mcp] Server listening on http://localhost%s\n", addr)
	fmt.Printf("[mcp] ConnectRPC: POST /mcp.v1.McpService/{ListTools,CallTool}\n")
	fmt.Printf("[mcp] ConnectRPC: POST /mcp.v1.McpAdminService/{GenerateKey,ListKeys,RevokeKey}\n")
	fmt.Printf("[mcp] JSON-RPC:   POST /mcp\n")
	fmt.Printf("[mcp] Mailbox:    GET  /mailbox/events\n")
	fmt.Printf("[mcp] Health:     GET  /health\n")
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "[mcp] fatal: %v\n", err)
		os.Exit(1)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// repoRoot walks up from the current executable location looking for
// nx.json, package.json, or go.mod to identify the repo root.
func repoRoot() string {
	// First try working directory
	if wd, err := os.Getwd(); err == nil {
		if root := findRepoRoot(wd); root != "" {
			return root
		}
	}
	// Fall back to executable location
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	dir := filepath.Dir(exe)
	if root := findRepoRoot(dir); root != "" {
		return root
	}
	return "."
}

func findRepoRoot(start string) string {
	markers := []string{"nx.json", "package.json"}
	dir := start
	for {
		for _, m := range markers {
			if _, err := os.Stat(filepath.Join(dir, m)); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// authMiddleware validates Bearer/X-Api-Key headers for plain HTTP handlers.
func authMiddleware(storePath string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := extractHTTPKey(r)
		if key == "" || !keys.Validate(storePath, key) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func extractHTTPKey(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
		return auth
	}
	return r.Header.Get("X-Api-Key")
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

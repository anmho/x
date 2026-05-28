package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderMCPEnvFile(t *testing.T) {
	body := renderMCPEnvFile(mcpSetupConfig{
		Port:               "8765",
		APIKey:             "mcp_test_key",
		LinearAPIKey:       "lin_test",
		ProxyDefaultURL:    "http://proxy.local/mcp",
		ProxyDefaultKey:    "proxy_key",
		ProxyDefaultHeader: "Authorization",
		ProxyDefaultScheme: "Bearer",
		ProviderOverrides: []mcpProviderOverride{
			{Name: "google_docs", URL: "http://docs.local/mcp", Key: "docs_key", Tool: "docs_document_get"},
			{Name: "perplexity", URL: "http://px.local/mcp", Key: "px_key", Tool: "perplexity_search"},
		},
	})

	for _, expected := range []string{
		"MCP_PORT=8765",
		"MCP_API_KEYS=mcp_test_key",
		"LINEAR_API_KEY=lin_test",
		"MCP_PROXY_DEFAULT_URL=http://proxy.local/mcp",
		"MCP_PROXY_GOOGLE_DOCS_URL=http://docs.local/mcp",
		"MCP_PROXY_PERPLEXITY_TOOL=perplexity_search",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected env body to contain %q, got:\n%s", expected, body)
		}
	}
}

func TestRunMcpSetupWritesEnvFile(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, "mcp.env")
	input := strings.Join([]string{
		tempDir,
		envFile,
		"8877",
		"mcp_custom_key",
		"yes",
		"lin_custom",
		"no",
		"http://proxy.local/mcp",
		"proxy_key",
		"Authorization",
		"Bearer",
		"google_docs,perplexity",
		"http://docs.local/mcp",
		"docs_key",
		"Authorization",
		"Bearer",
		"docs_document_get",
		"",
		"",
		"",
		"",
		"perplexity_search",
		"",
	}, "\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runMcpSetup(nil, strings.NewReader(input), &stdout, &stderr); err != nil {
		t.Fatalf("runMcpSetup: %v\nstderr=%s", err, stderr.String())
	}

	body, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("read env file: %v", err)
	}
	text := string(body)
	for _, expected := range []string{
		"MCP_PORT=8877",
		"MCP_API_KEYS=mcp_custom_key",
		"LINEAR_API_KEY=lin_custom",
		"MCP_PROXY_DEFAULT_URL=http://proxy.local/mcp",
		"MCP_PROXY_GOOGLE_DOCS_URL=http://docs.local/mcp",
		"MCP_PROXY_PERPLEXITY_TOOL=perplexity_search",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected env file to contain %q, got:\n%s", expected, text)
		}
	}
	if !strings.Contains(stdout.String(), "docker run --rm -p 8877:8877") {
		t.Fatalf("expected docker instructions, got:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "platform mcp --server http://127.0.0.1:8877 --key 'mcp_custom_key' tools list") {
		t.Fatalf("expected client instructions, got:\n%s", stdout.String())
	}
}

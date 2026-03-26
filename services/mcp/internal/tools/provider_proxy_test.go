package tools

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

type proxyRoundTripFunc func(*http.Request) (*http.Response, error)

func (f proxyRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestRegistryProviderProxyWrappers(t *testing.T) {
	t.Setenv("MCP_MAILBOX_FILE", filepath.Join(t.TempDir(), "collab.json"))

	type requestEnvelope struct {
		Method string `json:"method"`
		Params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		} `json:"params"`
	}

	var gotAuth string
	var gotMethod string
	var gotTool string
	var gotArgs map[string]interface{}
	client := &http.Client{Transport: proxyRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotAuth = r.Header.Get("Authorization")
		var req requestEnvelope
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode upstream request: %v", err)
		}
		gotMethod = req.Method
		gotTool = req.Params.Name
		gotArgs = req.Params.Arguments
		body, err := json.Marshal(map[string]any{
			"jsonrpc": "2.0",
			"id":      "x-mcp-proxy",
			"result": map[string]any{
				"content": []map[string]any{{"type": "text", "text": "proxy ok"}},
				"isError": false,
			},
		})
		if err != nil {
			t.Fatalf("marshal upstream response: %v", err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(body)),
		}, nil
	})}

	reg := NewRegistry(t.TempDir())
	reg.SetProviderProxyClient(NewProviderProxyClient(map[string]ProviderProxyConfig{
		"google_search": {
			URL:    "http://proxy.invalid/mcp",
			APIKey: "secret-key",
			Tool:   "remote_google_search",
		},
		"google_docs": {
			URL:    "http://proxy.invalid/mcp",
			APIKey: "secret-key",
			Tool:   "docs_document_get",
		},
	}, client))

	result, isError, err := reg.Call("google_search_query", `{"query":"agent routing","limit":3}`)
	if err != nil {
		t.Fatalf("Call(google_search_query): %v", err)
	}
	if isError {
		t.Fatalf("expected success, got error result: %s", result)
	}
	if strings.TrimSpace(result) != "proxy ok" {
		t.Fatalf("unexpected result: %q", result)
	}
	if gotAuth != "Bearer secret-key" {
		t.Fatalf("unexpected auth header: %q", gotAuth)
	}
	if gotMethod != "tools/call" || gotTool != "remote_google_search" {
		t.Fatalf("unexpected upstream method/tool: %s %s", gotMethod, gotTool)
	}
	if gotArgs["query"] != "agent routing" {
		t.Fatalf("unexpected upstream args: %#v", gotArgs)
	}

	result, isError, err = reg.Call("google_docs_proxy_call", `{"arguments":{"doc_id":"abc123","include_tabs":true}}`)
	if err != nil {
		t.Fatalf("Call(google_docs_proxy_call): %v", err)
	}
	if isError {
		t.Fatalf("expected success, got error result: %s", result)
	}
	if gotTool != "docs_document_get" {
		t.Fatalf("unexpected docs tool: %s", gotTool)
	}
	if gotArgs["doc_id"] != "abc123" {
		t.Fatalf("unexpected docs args: %#v", gotArgs)
	}
}

func TestRegistryExternalProviderToolWrappers(t *testing.T) {
	t.Setenv("MCP_MAILBOX_FILE", filepath.Join(t.TempDir(), "collab.json"))

	reg := NewRegistry(t.TempDir())

	result, isError, err := reg.Call("collab_get_or_create_github_channel", `{"repository":"anmho/x","number":"212","kind":"pull_request","title":"External adapters","participants":["agent-a"]}`)
	if err != nil {
		t.Fatalf("Call(collab_get_or_create_github_channel): %v", err)
	}
	if isError {
		t.Fatalf("expected success, got error result: %s", result)
	}

	var channel struct {
		ID           string               `json:"id"`
		Key          string               `json:"key"`
		ExternalRefs []*CollabExternalRef `json:"external_refs"`
		Metadata     map[string]string    `json:"metadata"`
		Participants []string             `json:"participants"`
	}
	if err := json.Unmarshal([]byte(result), &channel); err != nil {
		t.Fatalf("unmarshal github channel: %v", err)
	}
	if channel.Key != "github:anmho/x#212" {
		t.Fatalf("unexpected key: %s", channel.Key)
	}
	if len(channel.ExternalRefs) != 1 || channel.ExternalRefs[0].ParentID != "pull_request" {
		t.Fatalf("unexpected external refs: %#v", channel.ExternalRefs)
	}

	result, isError, err = reg.Call("collab_link_google_doc", `{"channel_id":"`+channel.ID+`","doc_id":"doc-123","title":"Spec","url":"https://docs.google.com/document/d/doc-123/edit"}`)
	if err != nil {
		t.Fatalf("Call(collab_link_google_doc): %v", err)
	}
	if isError {
		t.Fatalf("expected success, got error result: %s", result)
	}
	if err := json.Unmarshal([]byte(result), &channel); err != nil {
		t.Fatalf("unmarshal linked channel: %v", err)
	}
	if len(channel.ExternalRefs) != 2 {
		t.Fatalf("expected 2 external refs, got %#v", channel.ExternalRefs)
	}

	linked, err := reg.collab.GetChannelByExternalRef("google_docs", "doc-123", "")
	if err != nil {
		t.Fatalf("GetChannelByExternalRef(google_docs): %v", err)
	}
	if linked.ID != channel.ID {
		t.Fatalf("expected same channel id, got %s and %s", linked.ID, channel.ID)
	}
}

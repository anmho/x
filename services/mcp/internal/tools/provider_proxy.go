package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var supportedProxyProviders = []string{
	"github",
	"google_docs",
	"google_search",
	"perplexity",
	"glean",
	"yahoo",
	"yahoo_finance",
	"coinbase",
}

type ProviderProxyConfig struct {
	Provider string
	URL      string
	APIKey   string
	Header   string
	Scheme   string
	Tool     string
}

type ProviderProxyClient struct {
	configs    map[string]ProviderProxyConfig
	httpClient *http.Client
}

func NewProviderProxyClient(configs map[string]ProviderProxyConfig, httpClient *http.Client) *ProviderProxyClient {
	copyConfigs := make(map[string]ProviderProxyConfig, len(configs))
	for provider, cfg := range configs {
		provider = normalizeProviderName(provider)
		if provider == "" {
			continue
		}
		cfg.Provider = provider
		if cfg.Header == "" {
			cfg.Header = "Authorization"
		}
		if cfg.Scheme == "" {
			cfg.Scheme = "Bearer"
		}
		if cfg.Tool == "" {
			cfg.Tool = defaultProxyToolName(provider)
		}
		copyConfigs[provider] = cfg
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &ProviderProxyClient{configs: copyConfigs, httpClient: httpClient}
}

func NewProviderProxyClientFromEnv() *ProviderProxyClient {
	defaults := ProviderProxyConfig{
		URL:    strings.TrimSpace(os.Getenv("MCP_PROXY_DEFAULT_URL")),
		APIKey: strings.TrimSpace(os.Getenv("MCP_PROXY_DEFAULT_KEY")),
		Header: strings.TrimSpace(os.Getenv("MCP_PROXY_DEFAULT_HEADER")),
		Scheme: strings.TrimSpace(os.Getenv("MCP_PROXY_DEFAULT_SCHEME")),
	}
	configs := make(map[string]ProviderProxyConfig, len(supportedProxyProviders))
	for _, provider := range supportedProxyProviders {
		prefix := "MCP_PROXY_" + strings.ToUpper(provider) + "_"
		cfg := defaults
		cfg.Provider = provider
		if v := strings.TrimSpace(os.Getenv(prefix + "URL")); v != "" {
			cfg.URL = v
		}
		if v := strings.TrimSpace(os.Getenv(prefix + "KEY")); v != "" {
			cfg.APIKey = v
		}
		if v := strings.TrimSpace(os.Getenv(prefix + "HEADER")); v != "" {
			cfg.Header = v
		}
		if v := strings.TrimSpace(os.Getenv(prefix + "SCHEME")); v != "" {
			cfg.Scheme = v
		}
		if v := strings.TrimSpace(os.Getenv(prefix + "TOOL")); v != "" {
			cfg.Tool = v
		}
		if cfg.URL == "" {
			continue
		}
		configs[provider] = cfg
	}
	return NewProviderProxyClient(configs, nil)
}

func (c *ProviderProxyClient) Enabled(provider string) bool {
	if c == nil {
		return false
	}
	_, ok := c.configFor(provider)
	return ok
}

func (c *ProviderProxyClient) CallTool(ctx context.Context, provider, tool string, arguments map[string]any) (string, error) {
	cfg, ok := c.configFor(provider)
	if !ok {
		return "", fmt.Errorf("provider proxy %q is not configured", normalizeProviderName(provider))
	}
	if tool = strings.TrimSpace(tool); tool == "" {
		tool = cfg.Tool
	}
	if tool == "" {
		return "", errors.New("tool name is required")
	}
	if arguments == nil {
		arguments = map[string]any{}
	}

	body := map[string]any{
		"jsonrpc": "2.0",
		"id":      "x-mcp-proxy",
		"method":  "tools/call",
		"params": map[string]any{
			"name":      tool,
			"arguments": arguments,
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.URL, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	setProxyAuthHeader(req, cfg)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var parsed struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		if parsed.Error != nil && parsed.Error.Message != "" {
			return "", fmt.Errorf("upstream proxy request failed: %s", parsed.Error.Message)
		}
		return "", fmt.Errorf("upstream proxy request failed: status %d", resp.StatusCode)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("upstream proxy error: %s", parsed.Error.Message)
	}
	text := collectProxyText(parsed.Result.Content)
	if parsed.Result.IsError {
		if text == "" {
			text = "upstream tool call failed"
		}
		return "", errors.New(text)
	}
	return text, nil
}

func (c *ProviderProxyClient) configFor(provider string) (ProviderProxyConfig, bool) {
	if c == nil {
		return ProviderProxyConfig{}, false
	}
	cfg, ok := c.configs[normalizeProviderName(provider)]
	return cfg, ok
}

func normalizeProviderName(provider string) string {
	return strings.TrimSpace(strings.ToLower(provider))
}

func defaultProxyToolName(provider string) string {
	switch normalizeProviderName(provider) {
	case "google_docs":
		return "google_docs"
	case "google_search":
		return "google_search"
	case "yahoo_finance":
		return "yahoo_finance"
	default:
		return normalizeProviderName(provider)
	}
}

func setProxyAuthHeader(req *http.Request, cfg ProviderProxyConfig) {
	if req == nil || cfg.APIKey == "" {
		return
	}
	header := cfg.Header
	if header == "" {
		header = "Authorization"
	}
	if strings.EqualFold(header, "Authorization") {
		scheme := cfg.Scheme
		if scheme == "" {
			scheme = "Bearer"
		}
		if scheme == "raw" {
			req.Header.Set(header, cfg.APIKey)
			return
		}
		req.Header.Set(header, strings.TrimSpace(scheme+" "+cfg.APIKey))
		return
	}
	req.Header.Set(header, cfg.APIKey)
}

func collectProxyText(content []struct {
	Type string `json:"type"`
	Text string `json:"text"`
}) string {
	parts := make([]string, 0, len(content))
	for _, item := range content {
		text := strings.TrimSpace(item.Text)
		if item.Type == "text" && text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n")
}

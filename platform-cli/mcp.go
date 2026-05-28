package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func runMcp(args []string) error {
	if len(args) > 0 && args[0] == "setup" {
		return runMcpSetup(args[1:], os.Stdin, os.Stdout, os.Stderr)
	}
	mcpBin, err := findMcpBin()
	if err != nil {
		return fmt.Errorf("mcp binary not found; run `make build-mcp` first: %w", err)
	}
	cmd := exec.Command(mcpBin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findMcpBin() (string, error) {
	// 1. Check PATH
	if p, err := exec.LookPath("mcp"); err == nil {
		return p, nil
	}
	// 2. Check bin/mcp relative to executable
	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "mcp")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", errors.New("mcp binary not found")
}

type mcpSetupConfig struct {
	DataDir            string
	EnvFile            string
	Port               string
	APIKey             string
	LinearAPIKey       string
	SlackSigningSecret string
	SlackBotToken      string
	ProxyDefaultURL    string
	ProxyDefaultKey    string
	ProxyDefaultHeader string
	ProxyDefaultScheme string
	ProviderOverrides  []mcpProviderOverride
}

type mcpProviderOverride struct {
	Name   string
	URL    string
	Key    string
	Header string
	Scheme string
	Tool   string
}

func runMcpSetup(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("platform mcp setup", flag.ContinueOnError)
	fs.SetOutput(stderr)

	defaultDataDir := defaultMcpDataDir()
	defaultEnvFile := filepath.Join(defaultDataDir, "mcp.env")
	dataDir := fs.String("data-dir", defaultDataDir, "Host data directory mounted to /root/.x-mcp in the container")
	envFile := fs.String("env-file", defaultEnvFile, "Path to the generated Docker env file")
	port := fs.String("port", "8765", "Container port to expose")
	apiKey := fs.String("api-key", "", "Gateway API key to write; if empty, one is generated")
	if err := fs.Parse(args); err != nil {
		return err
	}

	reader := bufio.NewReader(stdin)
	cfg := mcpSetupConfig{
		DataDir: defaultString(*dataDir, defaultDataDir),
		EnvFile: defaultString(*envFile, defaultEnvFile),
		Port:    defaultString(*port, "8765"),
		APIKey:  strings.TrimSpace(*apiKey),
	}
	if cfg.APIKey == "" {
		key, err := newMCPAPIKey()
		if err != nil {
			return err
		}
		cfg.APIKey = key
	}

	fmt.Fprintln(stdout, "platform mcp setup")
	fmt.Fprintln(stdout, "This wizard writes a Docker-ready env file for the MCP gateway.")
	fmt.Fprintln(stdout, "Secrets are stored in the output env file; keep that file private.")
	fmt.Fprintln(stdout)

	var err error
	if cfg.DataDir, err = promptText(reader, stdout, "Host data dir", cfg.DataDir); err != nil {
		return err
	}
	if cfg.EnvFile, err = promptText(reader, stdout, "Env file path", cfg.EnvFile); err != nil {
		return err
	}
	if cfg.Port, err = promptText(reader, stdout, "Container port", cfg.Port); err != nil {
		return err
	}
	if cfg.APIKey, err = promptText(reader, stdout, "Gateway API key", cfg.APIKey); err != nil {
		return err
	}

	enableLinear, err := promptYesNo(reader, stdout, "Configure Linear issue enrichment", false)
	if err != nil {
		return err
	}
	if enableLinear {
		if cfg.LinearAPIKey, err = promptText(reader, stdout, "LINEAR_API_KEY", cfg.LinearAPIKey); err != nil {
			return err
		}
	}

	enableSlack, err := promptYesNo(reader, stdout, "Configure Slack bridge", false)
	if err != nil {
		return err
	}
	if enableSlack {
		if cfg.SlackSigningSecret, err = promptText(reader, stdout, "SLACK_SIGNING_SECRET", cfg.SlackSigningSecret); err != nil {
			return err
		}
		if cfg.SlackBotToken, err = promptText(reader, stdout, "SLACK_BOT_TOKEN", cfg.SlackBotToken); err != nil {
			return err
		}
	}

	if cfg.ProxyDefaultURL, err = promptText(reader, stdout, "Default upstream MCP proxy URL (optional)", cfg.ProxyDefaultURL); err != nil {
		return err
	}
	if cfg.ProxyDefaultURL != "" {
		if cfg.ProxyDefaultKey, err = promptText(reader, stdout, "Default upstream MCP proxy key (optional)", cfg.ProxyDefaultKey); err != nil {
			return err
		}
		if cfg.ProxyDefaultHeader, err = promptText(reader, stdout, "Default upstream auth header", defaultString(cfg.ProxyDefaultHeader, "Authorization")); err != nil {
			return err
		}
		if cfg.ProxyDefaultScheme, err = promptText(reader, stdout, "Default upstream auth scheme", defaultString(cfg.ProxyDefaultScheme, "Bearer")); err != nil {
			return err
		}
	}

	providerList, err := promptText(reader, stdout, "Provider overrides to configure (comma-separated: github,google_docs,google_search,perplexity,glean,yahoo,yahoo_finance,coinbase)", "")
	if err != nil {
		return err
	}
	for _, name := range parseProviderList(providerList) {
		override, err := promptProviderOverride(reader, stdout, name, cfg)
		if err != nil {
			return err
		}
		cfg.ProviderOverrides = append(cfg.ProviderOverrides, override)
	}

	if err := writeMCPSetupFiles(cfg); err != nil {
		return err
	}
	printMCPSetupSummary(stdout, cfg)
	return nil
}

func promptText(reader *bufio.Reader, stdout io.Writer, label, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Fprintf(stdout, "%s [%s]: ", label, defaultValue)
	} else {
		fmt.Fprintf(stdout, "%s: ", label)
	}
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return strings.TrimSpace(defaultValue), nil
	}
	return line, nil
}

func promptYesNo(reader *bufio.Reader, stdout io.Writer, label string, def bool) (bool, error) {
	choice := "y/N"
	if def {
		choice = "Y/n"
	}
	for {
		fmt.Fprintf(stdout, "%s [%s]: ", label, choice)
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return false, err
		}
		value := strings.TrimSpace(strings.ToLower(line))
		switch value {
		case "":
			return def, nil
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		}
		fmt.Fprintln(stdout, "Please answer yes or no.")
	}
}

func promptProviderOverride(reader *bufio.Reader, stdout io.Writer, name string, cfg mcpSetupConfig) (mcpProviderOverride, error) {
	override := mcpProviderOverride{
		Name:   name,
		URL:    cfg.ProxyDefaultURL,
		Key:    cfg.ProxyDefaultKey,
		Header: defaultString(cfg.ProxyDefaultHeader, "Authorization"),
		Scheme: defaultString(cfg.ProxyDefaultScheme, "Bearer"),
		Tool:   defaultProviderTool(name),
	}
	fmt.Fprintf(stdout, "\nConfiguring provider override for %s\n", name)
	var err error
	if override.URL, err = promptText(reader, stdout, strings.ToUpper(name)+" URL", override.URL); err != nil {
		return mcpProviderOverride{}, err
	}
	if override.Key, err = promptText(reader, stdout, strings.ToUpper(name)+" KEY", override.Key); err != nil {
		return mcpProviderOverride{}, err
	}
	if override.Header, err = promptText(reader, stdout, strings.ToUpper(name)+" HEADER", override.Header); err != nil {
		return mcpProviderOverride{}, err
	}
	if override.Scheme, err = promptText(reader, stdout, strings.ToUpper(name)+" SCHEME", override.Scheme); err != nil {
		return mcpProviderOverride{}, err
	}
	if override.Tool, err = promptText(reader, stdout, strings.ToUpper(name)+" TOOL", override.Tool); err != nil {
		return mcpProviderOverride{}, err
	}
	return override, nil
}

func writeMCPSetupFiles(cfg mcpSetupConfig) error {
	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cfg.EnvFile), 0o700); err != nil {
		return err
	}
	body := renderMCPEnvFile(cfg)
	return os.WriteFile(cfg.EnvFile, []byte(body), 0o600)
}

func renderMCPEnvFile(cfg mcpSetupConfig) string {
	lines := []string{
		"# Generated by `platform mcp setup`",
		"# Portable Docker env for the Project X MCP gateway",
		"MCP_PORT=" + cfg.Port,
		"MCP_ROOT=/app",
		"MCP_KEYS_FILE=/root/.x-mcp/keys.json",
		"MCP_MAILBOX_FILE=/root/.x-mcp/collab.json",
		"MCP_API_KEYS=" + shellEscapeEnvValue(cfg.APIKey),
	}
	if cfg.LinearAPIKey != "" {
		lines = append(lines, "LINEAR_API_KEY="+shellEscapeEnvValue(cfg.LinearAPIKey))
	}
	if cfg.SlackSigningSecret != "" {
		lines = append(lines, "SLACK_SIGNING_SECRET="+shellEscapeEnvValue(cfg.SlackSigningSecret))
	}
	if cfg.SlackBotToken != "" {
		lines = append(lines, "SLACK_BOT_TOKEN="+shellEscapeEnvValue(cfg.SlackBotToken))
	}
	if cfg.ProxyDefaultURL != "" {
		lines = append(lines, "MCP_PROXY_DEFAULT_URL="+shellEscapeEnvValue(cfg.ProxyDefaultURL))
		if cfg.ProxyDefaultKey != "" {
			lines = append(lines, "MCP_PROXY_DEFAULT_KEY="+shellEscapeEnvValue(cfg.ProxyDefaultKey))
		}
		if cfg.ProxyDefaultHeader != "" {
			lines = append(lines, "MCP_PROXY_DEFAULT_HEADER="+shellEscapeEnvValue(cfg.ProxyDefaultHeader))
		}
		if cfg.ProxyDefaultScheme != "" {
			lines = append(lines, "MCP_PROXY_DEFAULT_SCHEME="+shellEscapeEnvValue(cfg.ProxyDefaultScheme))
		}
	}

	sort.Slice(cfg.ProviderOverrides, func(i, j int) bool {
		return cfg.ProviderOverrides[i].Name < cfg.ProviderOverrides[j].Name
	})
	for _, override := range cfg.ProviderOverrides {
		prefix := "MCP_PROXY_" + strings.ToUpper(override.Name) + "_"
		if override.URL != "" {
			lines = append(lines, prefix+"URL="+shellEscapeEnvValue(override.URL))
		}
		if override.Key != "" {
			lines = append(lines, prefix+"KEY="+shellEscapeEnvValue(override.Key))
		}
		if override.Header != "" {
			lines = append(lines, prefix+"HEADER="+shellEscapeEnvValue(override.Header))
		}
		if override.Scheme != "" {
			lines = append(lines, prefix+"SCHEME="+shellEscapeEnvValue(override.Scheme))
		}
		if override.Tool != "" {
			lines = append(lines, prefix+"TOOL="+shellEscapeEnvValue(override.Tool))
		}
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

func printMCPSetupSummary(stdout io.Writer, cfg mcpSetupConfig) {
	serverURL := "http://127.0.0.1:" + cfg.Port
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "Wrote MCP env file: %s\n", cfg.EnvFile)
	fmt.Fprintf(stdout, "Host data dir:      %s\n", cfg.DataDir)
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Run locally with Docker:")
	fmt.Fprintf(stdout, "  docker run --rm -p %s:%s --env-file %s -v %s:/root/.x-mcp x-mcp\n", cfg.Port, cfg.Port, shellQuote(cfg.EnvFile), shellQuote(cfg.DataDir))
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Then call it with platform CLI:")
	fmt.Fprintf(stdout, "  platform mcp --server %s --key %s tools list\n", serverURL, shellQuote(cfg.APIKey))
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "For centralized hosting, reuse the same env file as the container secret/env source.")
}

func defaultMcpDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return filepath.Join(os.TempDir(), "x-mcp")
	}
	return filepath.Join(home, ".x-mcp")
}

func newMCPAPIKey() (string, error) {
	var buf [24]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return "mcp_" + hex.EncodeToString(buf[:]), nil
}

func parseProviderList(raw string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, part := range strings.Split(raw, ",") {
		name := strings.TrimSpace(strings.ToLower(part))
		if name == "" {
			continue
		}
		switch name {
		case "github", "google_docs", "google_search", "perplexity", "glean", "yahoo", "yahoo_finance", "coinbase":
		default:
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func defaultProviderTool(name string) string {
	switch name {
	case "google_docs":
		return "docs_document_get"
	case "google_search":
		return "google_search"
	case "yahoo_finance":
		return "yahoo_finance"
	default:
		return name
	}
}

func shellEscapeEnvValue(value string) string {
	return strings.ReplaceAll(value, "\n", "")
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

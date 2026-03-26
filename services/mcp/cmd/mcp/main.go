package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/anmhela/x/mcp/internal/keys"
	mcpv1 "github.com/anmhela/x/mcp/internal/rpc/gen/mcp/v1"
	"github.com/anmhela/x/mcp/internal/rpc/gen/mcp/v1/mcpv1connect"
)

func getEnv(key, defaultVal string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return defaultVal
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "mcp: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	server := getEnv("MCP_SERVER", "http://localhost:8765")
	key := os.Getenv("MCP_API_KEY")

	args, server, key = parseGlobalFlags(args, server, key)

	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "tools":
		return runTools(args[1:], server, key)
	case "keys":
		return runKeys(args[1:], server, key)
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func parseGlobalFlags(args []string, server, key string) ([]string, string, string) {
	remaining := make([]string, 0, len(args))
	i := 0
	for i < len(args) {
		switch args[i] {
		case "--server":
			if i+1 < len(args) {
				server = args[i+1]
				i += 2
			} else {
				i++
			}
		case "--key":
			if i+1 < len(args) {
				key = args[i+1]
				i += 2
			} else {
				i++
			}
		default:
			if strings.HasPrefix(args[i], "--server=") {
				server = strings.TrimPrefix(args[i], "--server=")
				i++
			} else if strings.HasPrefix(args[i], "--key=") {
				key = strings.TrimPrefix(args[i], "--key=")
				i++
			} else {
				remaining = append(remaining, args[i])
				i++
			}
		}
	}
	return remaining, server, key
}

func printUsage() {
	fmt.Println(`mcp - MCP server CLI

Usage:
  mcp tools list
  mcp tools call <name> [key=value ...]
  mcp keys generate [--name <label>] [--local] [--store <path>]
  mcp keys list
  mcp keys revoke <id>

Global flags:
  --server <url>    Server base URL (default: http://localhost:8765, env: MCP_SERVER)
  --key <api-key>   API key for auth (env: MCP_API_KEY)`)
}

func authClientInterceptor(key string) connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if key != "" {
				req.Header().Set("Authorization", "Bearer "+key)
			}
			return next(ctx, req)
		}
	})
}

func newMcpClient(server, key string) mcpv1connect.McpServiceClient {
	return mcpv1connect.NewMcpServiceClient(
		http.DefaultClient,
		server,
		connect.WithInterceptors(authClientInterceptor(key)),
	)
}

func newMcpAdminClient(server, key string) mcpv1connect.McpAdminServiceClient {
	return mcpv1connect.NewMcpAdminServiceClient(
		http.DefaultClient,
		server,
		connect.WithInterceptors(authClientInterceptor(key)),
	)
}

// runTools handles the "tools" subcommand.
func runTools(args []string, server, key string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: mcp tools <list|call>")
	}
	switch args[0] {
	case "list":
		return runToolsList(server, key)
	case "call":
		return runToolsCall(args[1:], server, key)
	default:
		return fmt.Errorf("unknown tools command: %s", args[0])
	}
}

func runToolsList(server, key string) error {
	client := newMcpClient(server, key)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.ListTools(ctx, connect.NewRequest(&mcpv1.ListToolsRequest{}))
	if err != nil {
		return fmt.Errorf("list tools: %w", err)
	}

	for _, tool := range resp.Msg.GetTools() {
		fmt.Printf("%-30s %s\n", tool.GetName(), tool.GetDescription())
	}
	return nil
}

func runToolsCall(args []string, server, key string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: mcp tools call <name> [key=value ...]")
	}

	name := args[0]
	kvArgs := args[1:]

	argsMap := make(map[string]interface{})
	for _, kv := range kvArgs {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid argument %q: expected key=value", kv)
		}
		k, v := parts[0], parts[1]
		// Try to parse as a number; otherwise keep as string.
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			argsMap[k] = n
		} else if f, err := strconv.ParseFloat(v, 64); err == nil {
			argsMap[k] = f
		} else {
			argsMap[k] = v
		}
	}

	argsBytes, err := json.Marshal(argsMap)
	if err != nil {
		return fmt.Errorf("marshal arguments: %w", err)
	}

	client := newMcpClient(server, key)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.CallTool(ctx, connect.NewRequest(&mcpv1.CallToolRequest{
		Name:          name,
		ArgumentsJson: string(argsBytes),
	}))
	if err != nil {
		return fmt.Errorf("call tool: %w", err)
	}

	fmt.Println(resp.Msg.GetResult())
	return nil
}

// runKeys handles the "keys" subcommand.
func runKeys(args []string, server, key string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: mcp keys <generate|list|revoke>")
	}
	switch args[0] {
	case "generate":
		return runKeysGenerate(args[1:], server, key)
	case "list":
		return runKeysList(server, key)
	case "revoke":
		return runKeysRevoke(args[1:], server, key)
	default:
		return fmt.Errorf("unknown keys command: %s", args[0])
	}
}

func runKeysGenerate(args []string, server, key string) error {
	name := "default"
	local := false
	storePath := keys.DefaultStorePath()
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--name":
			if i+1 < len(args) {
				name = args[i+1]
				i++
			}
		case "--local":
			local = true
		case "--store":
			if i+1 < len(args) {
				storePath = args[i+1]
				i++
			}
		default:
			if strings.HasPrefix(args[i], "--name=") {
				name = strings.TrimPrefix(args[i], "--name=")
			} else if strings.HasPrefix(args[i], "--store=") {
				storePath = strings.TrimPrefix(args[i], "--store=")
			}
		}
	}

	if local {
		rec, err := keys.Generate(storePath, name)
		if err != nil {
			return fmt.Errorf("generate local key: %w", err)
		}
		fmt.Printf("Key:   %s\n", rec.Key)
		fmt.Printf("ID:    %s\n", rec.ID)
		fmt.Printf("Name:  %s\n", rec.Name)
		fmt.Printf("Store: %s\n", storePath)
		return nil
	}

	client := newMcpAdminClient(server, key)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.GenerateKey(ctx, connect.NewRequest(&mcpv1.GenerateKeyRequest{
		Name: name,
	}))
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	msg := resp.Msg
	meta := msg.GetMetadata()
	fmt.Printf("Key:  %s\n", msg.GetKey())
	if meta != nil {
		fmt.Printf("ID:   %s\n", meta.GetId())
		fmt.Printf("Name: %s\n", meta.GetName())
	}
	return nil
}

func runKeysList(server, key string) error {
	client := newMcpAdminClient(server, key)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.ListKeys(ctx, connect.NewRequest(&mcpv1.ListKeysRequest{}))
	if err != nil {
		return fmt.Errorf("list keys: %w", err)
	}

	fmt.Printf("%-36s  %-20s  %s\n", "ID", "Name", "Created")
	for _, k := range resp.Msg.GetKeys() {
		fmt.Printf("%-36s  %-20s  %s\n", k.GetId(), k.GetName(), k.GetCreatedAt())
	}
	return nil
}

func runKeysRevoke(args []string, server, key string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: mcp keys revoke <id>")
	}

	id := args[0]
	client := newMcpAdminClient(server, key)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.RevokeKey(ctx, connect.NewRequest(&mcpv1.RevokeKeyRequest{
		Id: id,
	}))
	if err != nil {
		return fmt.Errorf("revoke key: %w", err)
	}

	if resp.Msg.GetSuccess() {
		fmt.Println("Revoked.")
	} else {
		fmt.Println("Not found.")
	}
	return nil
}

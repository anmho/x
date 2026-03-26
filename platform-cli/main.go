package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func getEnv(key, defaultVal string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return defaultVal
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "platform: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "start":
		return runStack([]string{"start"})
	case "status":
		return runStack([]string{"status"})
	case "stop":
		return runStack([]string{"stop"})
	case "logs":
		if len(args) < 2 {
			return errors.New("usage: platform logs <service>")
		}
		return runStack([]string{"logs", args[1]})
	case "create":
		return runCreate(args[1:])
	case "notifications":
		return runNotifications(args[1:])
	case "stack":
		return runStack(args[1:])
	case "tokens":
		return runTokens(args[1:])
	case "docs":
		return runDocs(args[1:])
	case "project":
		return runProject(args[1:])
	case "deploy":
		return runDeploy(args[1:])
	case "control-plane":
		return runControlPlane(args[1:])
	case "config":
		return runConfig(args[1:])
	case "verify":
		return runVerify(args[1:])
	case "preflight":
		return runPreflight(args[1:])
	case "new", "scaffold":
		return runScaffold(args[1:])
	case "mcp":
		return runMcp(args[1:])
	case "doctor":
		return runDoctor()
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printUsage() {
	fmt.Println(`platform - Project X local CLI

Usage:
  platform start|status|stop
  platform logs <service>
  platform stack temporal-ui
  platform mcp <tools|keys> ...
  platform create <target> <name?>
  platform create integration <add|list|remove>
  platform notifications <send|list|status>
  platform docs [dev --port 3002]
  platform deploy --project <name> [--dry-run]
  platform control-plane <init|show|plan|apply|destroy|serve>
  platform project <list|show|keys ...>
  platform tokens <mint|list|rotate|revoke|check|audit|helpers>
  platform config <init|show>
  platform verify [all|platform|apps|docs]
  platform preflight [platform|cloud-console|omnichannel|mcp]

Examples:
  platform start
  platform stack temporal-ui
  platform deploy --project cloud-console
  platform control-plane apply --project cloud-console
  platform control-plane serve --addr :8091
  platform notifications list
  platform docs
  platform mcp tools list --server http://localhost:8765 --key <api-key>
  platform create service billing-api
  platform create integration add vercel --project cloud-console
  platform project keys mint --project cloud-console --owner andrew --env dev`)
}

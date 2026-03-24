package runner

import (
	"fmt"
	"strings"

	"github.com/anmho/agent-control-api/internal/domain"
)

type ExecutionSpec struct {
	Prompt  string
	Runtime domain.RuntimeConfig
	Env     map[string]string
}

type Invocation struct {
	Command string
	Args    []string
	Env     map[string]string
}

func BuildInvocation(spec ExecutionSpec) (Invocation, error) {
	spec.Runtime = spec.Runtime.Normalized()

	switch spec.Runtime.Provider {
	case domain.RuntimeClaude:
		return buildClaudeInvocation(spec), nil
	case domain.RuntimeCodex:
		return buildCodexInvocation(spec), nil
	default:
		return Invocation{}, fmt.Errorf("unsupported runtime provider %q", spec.Runtime.Provider)
	}
}

func CloudRunEnv(spec ExecutionSpec) map[string]string {
	runtime := spec.Runtime.Normalized()
	env := map[string]string{
		"AGENT_PROMPT":   spec.Prompt,
		"AGENT_RUNTIME":  string(runtime.Provider),
		"AGENT_SANDBOX":  string(runtime.SandboxMode),
		"AGENT_APPROVAL": string(runtime.ApprovalPolicy),
	}
	if len(runtime.AllowedTools) > 0 {
		env["AGENT_ALLOWED_TOOLS"] = strings.Join(runtime.AllowedTools, ",")
	}
	if len(runtime.AddDirs) > 0 {
		env["AGENT_ADD_DIRS"] = strings.Join(runtime.AddDirs, ",")
	}
	for k, v := range spec.Env {
		env[k] = v
	}
	return env
}

func buildClaudeInvocation(spec ExecutionSpec) Invocation {
	runtime := spec.Runtime.Normalized()
	args := []string{"--print"}
	if runtime.ApprovalPolicy == domain.ApprovalNever {
		if runtime.SandboxMode == domain.SandboxDangerFull {
			args = append(args, "--dangerously-skip-permissions")
		} else {
			args = append(args, "--permission-mode", "dontAsk")
		}
	} else {
		args = append(args, "--permission-mode", "default")
	}
	if len(runtime.AllowedTools) > 0 {
		args = append(args, "--allowedTools", strings.Join(runtime.AllowedTools, ","))
	}
	for _, dir := range runtime.AddDirs {
		if dir == "" {
			continue
		}
		args = append(args, "--add-dir", dir)
	}
	args = append(args, spec.Prompt)
	return Invocation{
		Command: "claude",
		Args:    args,
		Env:     spec.Env,
	}
}

func buildCodexInvocation(spec ExecutionSpec) Invocation {
	runtime := spec.Runtime.Normalized()
	args := []string{
		"-a", string(runtime.ApprovalPolicy),
		"exec",
		"--skip-git-repo-check",
		"--sandbox", string(runtime.SandboxMode),
	}
	for _, dir := range runtime.AddDirs {
		if dir == "" {
			continue
		}
		args = append(args, "--add-dir", dir)
	}
	args = append(args, spec.Prompt)
	return Invocation{
		Command: "codex",
		Args:    args,
		Env:     spec.Env,
	}
}

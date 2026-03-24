package domain

import "fmt"

type RuntimeProvider string

const (
	RuntimeClaude RuntimeProvider = "claude"
	RuntimeCodex  RuntimeProvider = "codex"
)

type SandboxMode string

const (
	SandboxReadOnly       SandboxMode = "read-only"
	SandboxWorkspaceWrite SandboxMode = "workspace-write"
	SandboxDangerFull     SandboxMode = "danger-full-access"
)

type ApprovalPolicy string

const (
	ApprovalNever     ApprovalPolicy = "never"
	ApprovalOnRequest ApprovalPolicy = "on-request"
)

// RuntimeConfig declares which agent runtime to use and which guardrails to apply.
type RuntimeConfig struct {
	Provider       RuntimeProvider `json:"provider"`
	SandboxMode    SandboxMode     `json:"sandbox_mode,omitempty"`
	ApprovalPolicy ApprovalPolicy  `json:"approval_policy,omitempty"`
	AllowedTools   []string        `json:"allowed_tools,omitempty"`
	AddDirs        []string        `json:"add_dirs,omitempty"`
}

func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		Provider:       RuntimeClaude,
		SandboxMode:    SandboxWorkspaceWrite,
		ApprovalPolicy: ApprovalNever,
	}
}

func (c RuntimeConfig) Normalized() RuntimeConfig {
	if c.Provider == "" {
		c.Provider = RuntimeClaude
	}
	if c.SandboxMode == "" {
		c.SandboxMode = SandboxWorkspaceWrite
	}
	if c.ApprovalPolicy == "" {
		c.ApprovalPolicy = ApprovalNever
	}
	return c
}

func (c RuntimeConfig) Validate() error {
	c = c.Normalized()
	switch c.Provider {
	case RuntimeClaude, RuntimeCodex:
	default:
		return fmt.Errorf("invalid runtime provider %q", c.Provider)
	}

	switch c.SandboxMode {
	case SandboxReadOnly, SandboxWorkspaceWrite, SandboxDangerFull:
	default:
		return fmt.Errorf("invalid sandbox mode %q", c.SandboxMode)
	}

	switch c.ApprovalPolicy {
	case ApprovalNever, ApprovalOnRequest:
	default:
		return fmt.Errorf("invalid approval policy %q", c.ApprovalPolicy)
	}

	return nil
}

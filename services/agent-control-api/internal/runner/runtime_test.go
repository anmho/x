package runner

import (
	"reflect"
	"testing"

	"github.com/anmho/agent-control-api/internal/domain"
)

func TestBuildCodexInvocationUsesGlobalApprovalFlag(t *testing.T) {
	invocation := buildCodexInvocation(ExecutionSpec{
		Prompt: "smoke test",
		Runtime: domain.RuntimeConfig{
			Provider:       domain.RuntimeCodex,
			SandboxMode:    domain.SandboxWorkspaceWrite,
			ApprovalPolicy: domain.ApprovalNever,
		},
	})

	want := []string{
		"-a", "never",
		"exec",
		"--skip-git-repo-check",
		"--sandbox", "workspace-write",
		"smoke test",
	}
	if !reflect.DeepEqual(invocation.Args, want) {
		t.Fatalf("unexpected codex args:\nwant %#v\ngot  %#v", want, invocation.Args)
	}
}

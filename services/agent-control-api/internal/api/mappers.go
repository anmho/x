package api

import (
	"time"

	"github.com/anmho/agent-control-api/internal/domain"
	agentcontrolv1 "github.com/anmho/agent-control-api/internal/rpc/gen/agentcontrol/v1"
)

func runToProto(run *domain.AgentRun, includeOutput bool) *agentcontrolv1.AgentRun {
	if run == nil {
		return nil
	}
	out := &agentcontrolv1.AgentRun{
		Id:        run.ID.String(),
		Message:   run.Message,
		Runtime:   runtimeToProto(run.Runtime),
		Status:    statusToProto(run.Status),
		CreatedAt: run.CreatedAt.UTC().Format(time.RFC3339),
		Resources: make([]*agentcontrolv1.Resource, 0, len(run.Resources)),
	}
	for _, resource := range run.Resources {
		out.Resources = append(out.Resources, resourceToProto(resource))
	}
	if run.JobName != "" {
		out.JobName = &run.JobName
	}
	if includeOutput && run.Output != "" {
		out.Output = &run.Output
	}
	if run.StartedAt != nil {
		v := run.StartedAt.UTC().Format(time.RFC3339)
		out.StartedAt = &v
	}
	if run.CompletedAt != nil {
		v := run.CompletedAt.UTC().Format(time.RFC3339)
		out.CompletedAt = &v
	}
	return out
}

func resourceToProto(resource domain.Resource) *agentcontrolv1.Resource {
	out := &agentcontrolv1.Resource{Uri: resource.URI}
	if resource.Title != "" {
		out.Title = &resource.Title
	}
	if resource.MIMEType != "" {
		out.MimeType = &resource.MIMEType
	}
	if resource.Text != "" {
		out.Text = &resource.Text
	}
	return out
}

func resourceFromProto(resource *agentcontrolv1.Resource) domain.Resource {
	if resource == nil {
		return domain.Resource{}
	}
	out := domain.Resource{URI: resource.GetUri()}
	if resource.Title != nil {
		out.Title = resource.GetTitle()
	}
	if resource.MimeType != nil {
		out.MIMEType = resource.GetMimeType()
	}
	if resource.Text != nil {
		out.Text = resource.GetText()
	}
	return out
}

func resourcesFromProto(resources []*agentcontrolv1.Resource) []domain.Resource {
	if len(resources) == 0 {
		return nil
	}
	out := make([]domain.Resource, 0, len(resources))
	for _, resource := range resources {
		out = append(out, resourceFromProto(resource))
	}
	return out
}

func runtimeToProto(cfg domain.RuntimeConfig) *agentcontrolv1.RuntimeConfig {
	cfg = cfg.Normalized()
	return &agentcontrolv1.RuntimeConfig{
		Provider:       providerToProto(cfg.Provider),
		SandboxMode:    sandboxToProto(cfg.SandboxMode),
		ApprovalPolicy: approvalToProto(cfg.ApprovalPolicy),
		AllowedTools:   append([]string(nil), cfg.AllowedTools...),
		AddDirs:        append([]string(nil), cfg.AddDirs...),
	}
}

func runtimeFromProto(cfg *agentcontrolv1.RuntimeConfig) domain.RuntimeConfig {
	if cfg == nil {
		return domain.DefaultRuntimeConfig()
	}
	return domain.RuntimeConfig{
		Provider:       providerFromProto(cfg.GetProvider()),
		SandboxMode:    sandboxFromProto(cfg.GetSandboxMode()),
		ApprovalPolicy: approvalFromProto(cfg.GetApprovalPolicy()),
		AllowedTools:   append([]string(nil), cfg.GetAllowedTools()...),
		AddDirs:        append([]string(nil), cfg.GetAddDirs()...),
	}.Normalized()
}

func controlEventToProto(event *domain.AgentRunEvent) *agentcontrolv1.RunControlEvent {
	if event == nil {
		return nil
	}
	out := &agentcontrolv1.RunControlEvent{
		Id:           event.ID.String(),
		RunId:        event.RunID.String(),
		DeliveryMode: deliveryModeToProto(event.DeliveryMode),
		Message:      event.Message,
		Reason:       event.Reason,
		Metadata:     map[string]string{},
		CreatedAt:    event.CreatedAt.UTC().Format(time.RFC3339),
	}
	if event.Sender != "" {
		out.Sender = &event.Sender
	}
	for k, v := range event.Metadata {
		out.Metadata[k] = v
	}
	return out
}

func statusToProto(status domain.RunStatus) agentcontrolv1.RunStatus {
	switch status {
	case domain.StatusPending:
		return agentcontrolv1.RunStatus_RUN_STATUS_PENDING
	case domain.StatusRunning:
		return agentcontrolv1.RunStatus_RUN_STATUS_RUNNING
	case domain.StatusSucceeded:
		return agentcontrolv1.RunStatus_RUN_STATUS_SUCCEEDED
	case domain.StatusFailed:
		return agentcontrolv1.RunStatus_RUN_STATUS_FAILED
	case domain.StatusCanceled:
		return agentcontrolv1.RunStatus_RUN_STATUS_CANCELED
	default:
		return agentcontrolv1.RunStatus_RUN_STATUS_UNSPECIFIED
	}
}

func providerToProto(provider domain.RuntimeProvider) agentcontrolv1.RuntimeProvider {
	switch provider {
	case domain.RuntimeClaude:
		return agentcontrolv1.RuntimeProvider_RUNTIME_PROVIDER_CLAUDE
	case domain.RuntimeCodex:
		return agentcontrolv1.RuntimeProvider_RUNTIME_PROVIDER_CODEX
	default:
		return agentcontrolv1.RuntimeProvider_RUNTIME_PROVIDER_UNSPECIFIED
	}
}

func providerFromProto(provider agentcontrolv1.RuntimeProvider) domain.RuntimeProvider {
	switch provider {
	case agentcontrolv1.RuntimeProvider_RUNTIME_PROVIDER_CLAUDE:
		return domain.RuntimeClaude
	case agentcontrolv1.RuntimeProvider_RUNTIME_PROVIDER_CODEX:
		return domain.RuntimeCodex
	default:
		return domain.RuntimeClaude
	}
}

func sandboxToProto(mode domain.SandboxMode) agentcontrolv1.SandboxMode {
	switch mode {
	case domain.SandboxReadOnly:
		return agentcontrolv1.SandboxMode_SANDBOX_MODE_READ_ONLY
	case domain.SandboxWorkspaceWrite:
		return agentcontrolv1.SandboxMode_SANDBOX_MODE_WORKSPACE_WRITE
	case domain.SandboxDangerFull:
		return agentcontrolv1.SandboxMode_SANDBOX_MODE_DANGER_FULL_ACCESS
	default:
		return agentcontrolv1.SandboxMode_SANDBOX_MODE_UNSPECIFIED
	}
}

func sandboxFromProto(mode agentcontrolv1.SandboxMode) domain.SandboxMode {
	switch mode {
	case agentcontrolv1.SandboxMode_SANDBOX_MODE_READ_ONLY:
		return domain.SandboxReadOnly
	case agentcontrolv1.SandboxMode_SANDBOX_MODE_DANGER_FULL_ACCESS:
		return domain.SandboxDangerFull
	default:
		return domain.SandboxWorkspaceWrite
	}
}

func approvalToProto(policy domain.ApprovalPolicy) agentcontrolv1.ApprovalPolicy {
	switch policy {
	case domain.ApprovalOnRequest:
		return agentcontrolv1.ApprovalPolicy_APPROVAL_POLICY_ON_REQUEST
	case domain.ApprovalNever:
		return agentcontrolv1.ApprovalPolicy_APPROVAL_POLICY_NEVER
	default:
		return agentcontrolv1.ApprovalPolicy_APPROVAL_POLICY_UNSPECIFIED
	}
}

func approvalFromProto(policy agentcontrolv1.ApprovalPolicy) domain.ApprovalPolicy {
	switch policy {
	case agentcontrolv1.ApprovalPolicy_APPROVAL_POLICY_ON_REQUEST:
		return domain.ApprovalOnRequest
	default:
		return domain.ApprovalNever
	}
}

func deliveryModeToProto(mode domain.PushDeliveryMode) agentcontrolv1.PushDeliveryMode {
	switch mode {
	case domain.PushDeliveryInterruptReplan:
		return agentcontrolv1.PushDeliveryMode_PUSH_DELIVERY_MODE_INTERRUPT_AND_REPLAN
	default:
		return agentcontrolv1.PushDeliveryMode_PUSH_DELIVERY_MODE_APPEND_CONTEXT
	}
}

func deliveryModeFromProto(mode agentcontrolv1.PushDeliveryMode) domain.PushDeliveryMode {
	switch mode {
	case agentcontrolv1.PushDeliveryMode_PUSH_DELIVERY_MODE_INTERRUPT_AND_REPLAN:
		return domain.PushDeliveryInterruptReplan
	default:
		return domain.PushDeliveryAppendContext
	}
}

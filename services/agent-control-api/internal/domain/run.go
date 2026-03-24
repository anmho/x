package domain

import (
	"time"

	"github.com/google/uuid"
)

type RunStatus string

const (
	StatusPending   RunStatus = "PENDING"
	StatusRunning   RunStatus = "RUNNING"
	StatusSucceeded RunStatus = "SUCCEEDED"
	StatusFailed    RunStatus = "FAILED"
	StatusCanceled  RunStatus = "CANCELED"
)

type PushDeliveryMode string

const (
	PushDeliveryAppendContext   PushDeliveryMode = "APPEND_CONTEXT"
	PushDeliveryInterruptReplan PushDeliveryMode = "INTERRUPT_AND_REPLAN"
)

// AgentRun represents a single agent execution.
type AgentRun struct {
	ID          uuid.UUID     `json:"id"`
	Message     string        `json:"message"`
	Runtime     RuntimeConfig `json:"runtime"`
	Status      RunStatus     `json:"status"`
	Resources   []Resource    `json:"resources,omitempty"`
	JobName     string        `json:"job_name,omitempty"`
	Output      string        `json:"output,omitempty"`
	StartedAt   *time.Time    `json:"started_at,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}

type Resource struct {
	URI      string `json:"uri"`
	Title    string `json:"title,omitempty"`
	MIMEType string `json:"mime_type,omitempty"`
	Text     string `json:"text,omitempty"`
}

type AgentRunEvent struct {
	ID           uuid.UUID         `json:"id"`
	RunID        uuid.UUID         `json:"run_id"`
	Sequence     int64             `json:"sequence"`
	DeliveryMode PushDeliveryMode  `json:"delivery_mode"`
	Message      string            `json:"message"`
	Reason       string            `json:"reason,omitempty"`
	Sender       string            `json:"sender,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}

// CreateRunRequest is the body for POST /v1/agent-runs.
type CreateRunRequest struct {
	Message   string        `json:"message"`
	Runtime   RuntimeConfig `json:"runtime,omitempty"`
	Resources []Resource    `json:"resources,omitempty"`
}

func (r *CreateRunRequest) Validate() error {
	if r.Message == "" {
		return ErrMissingMessage
	}
	r.Runtime = r.Runtime.Normalized()
	if err := r.Runtime.Validate(); err != nil {
		return err
	}
	return nil
}

func (m PushDeliveryMode) Normalized() PushDeliveryMode {
	switch m {
	case PushDeliveryInterruptReplan:
		return PushDeliveryInterruptReplan
	default:
		return PushDeliveryAppendContext
	}
}

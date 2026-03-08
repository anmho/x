package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string
type NotificationChannel string

const (
	StatusPending    NotificationStatus = "pending"
	StatusProcessing NotificationStatus = "processing"
	StatusSent       NotificationStatus = "sent"
	StatusFailed     NotificationStatus = "failed"
	StatusCancelled  NotificationStatus = "cancelled"

	ChannelEmail    NotificationChannel = "email"
	ChannelSMS      NotificationChannel = "sms"
	ChannelPush     NotificationChannel = "push"
	ChannelWebhook  NotificationChannel = "webhook"
	ChannelApp      NotificationChannel = "app"
	ChannelIMessage NotificationChannel = "imessage"
)

// Notification represents an email notification
type Notification struct {
	ID                 uuid.UUID              `json:"id"`
	UserID             *uuid.UUID             `json:"user_id"`
	TemplateID         *uuid.UUID             `json:"template_id"`
	RecipientEmail     string                 `json:"recipient_email"`
	Subject            string                 `json:"subject"`
	Body               string                 `json:"body"`
	Status             NotificationStatus     `json:"status"`
	TemporalWorkflowID string                 `json:"temporal_workflow_id,omitempty"`
	TemporalRunID      string                 `json:"temporal_run_id,omitempty"`
	ScheduledAt        *time.Time             `json:"scheduled_at,omitempty"`
	SentAt             *time.Time             `json:"sent_at,omitempty"`
	FailedAt           *time.Time             `json:"failed_at,omitempty"`
	ErrorMessage       string                 `json:"error_message,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// NotificationAttemptStatus represents the status of a delivery attempt
type NotificationAttemptStatus string

const (
	AttemptStatusSuccess  NotificationAttemptStatus = "success"
	AttemptStatusFailed   NotificationAttemptStatus = "failed"
	AttemptStatusRetrying NotificationAttemptStatus = "retrying"
)

// NotificationAttempt represents a delivery attempt
type NotificationAttempt struct {
	ID               uuid.UUID                 `json:"id"`
	NotificationID   uuid.UUID                 `json:"notification_id"`
	AttemptNumber    int                       `json:"attempt_number"`
	Status           NotificationAttemptStatus `json:"status"`
	ProviderResponse map[string]interface{}    `json:"provider_response,omitempty"`
	ErrorMessage     string                    `json:"error_message,omitempty"`
	AttemptedAt      time.Time                 `json:"attempted_at"`
}

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	UserID         *uuid.UUID             `json:"user_id"`
	TemplateID     *uuid.UUID             `json:"template_id"`
	Channel        NotificationChannel    `json:"channel,omitempty"`
	Recipient      string                 `json:"recipient,omitempty"`
	RecipientEmail string                 `json:"recipient_email"`
	Subject        string                 `json:"subject,omitempty"`
	Body           string                 `json:"body,omitempty"`
	Variables      map[string]string      `json:"variables,omitempty"`
	ScheduledAt    *time.Time             `json:"scheduled_at,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Validate validates the create notification request
func (r *CreateNotificationRequest) Validate() error {
	if r.Channel == "" {
		r.Channel = ChannelEmail
	}

	if r.RecipientEmail == "" && r.Recipient != "" {
		r.RecipientEmail = r.Recipient
	}

	if r.RecipientEmail == "" {
		return ErrInvalidEmail
	}

	switch r.Channel {
	case ChannelEmail, ChannelSMS, ChannelPush, ChannelWebhook, ChannelApp, ChannelIMessage:
	default:
		return ErrInvalidChannel
	}

	// If template ID is provided, we don't need subject/body
	if r.TemplateID == nil {
		if r.Subject == "" {
			return ErrMissingSubject
		}
		if r.Body == "" {
			return ErrMissingBody
		}
	}

	return nil
}

package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestCreateNotificationRequest_Validate(t *testing.T) {
	templateID := uuid.New()
	tests := []struct {
		name    string
		req     CreateNotificationRequest
		wantErr bool
	}{
		{
			name: "valid with template",
			req: CreateNotificationRequest{
				TemplateID:     &templateID,
				RecipientEmail: "user@example.com",
				Channel:        ChannelEmail,
			},
			wantErr: false,
		},
		{
			name: "valid with subject and body",
			req: CreateNotificationRequest{
				RecipientEmail: "user@example.com",
				Subject:        "Hello",
				Body:           "World",
				Channel:        ChannelEmail,
			},
			wantErr: false,
		},
		{
			name: "recipient from Recipient field",
			req: CreateNotificationRequest{
				Recipient: "user@example.com",
				Subject:   "Hello",
				Body:      "World",
			},
			wantErr: false,
		},
		{
			name: "missing email",
			req: CreateNotificationRequest{
				RecipientEmail: "",
				Subject:        "Hello",
				Body:           "World",
			},
			wantErr: true,
		},
		{
			name: "invalid channel",
			req: CreateNotificationRequest{
				RecipientEmail: "user@example.com",
				Subject:        "Hello",
				Body:           "World",
				Channel:        "invalid",
			},
			wantErr: true,
		},
		{
			name: "template without subject/body valid",
			req: CreateNotificationRequest{
				TemplateID:     &templateID,
				RecipientEmail: "user@example.com",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package domain

import (
	"testing"
)

func TestCreateTemplateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateTemplateRequest
		wantErr bool
	}{
		{
			name: "valid",
			req: CreateTemplateRequest{
				Name:    "welcome",
				Subject: "Welcome",
				Body:    "Hello {{name}}",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			req: CreateTemplateRequest{
				Name:    "",
				Subject: "Welcome",
				Body:    "Hello",
			},
			wantErr: true,
		},
		{
			name: "missing subject",
			req: CreateTemplateRequest{
				Name:    "welcome",
				Subject: "",
				Body:    "Hello",
			},
			wantErr: true,
		},
		{
			name: "missing body",
			req: CreateTemplateRequest{
				Name:    "welcome",
				Subject: "Welcome",
				Body:    "",
			},
			wantErr: true,
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

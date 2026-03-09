package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andrewho/omnichannel/internal/repository"
	"go.uber.org/zap"
)

func TestTemplateHandler_Create_Validation(t *testing.T) {
	// Use nil repo - we only test validation paths that return before Create is called
	logger := zap.NewNop()
	handler := NewTemplateHandler(logger, &repository.TemplateRepository{})

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "invalid JSON",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing name",
			body:       `{"subject":"Hi","body":"Hello"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing subject",
			body:       `{"name":"welcome","body":"Hello"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing body",
			body:       `{"name":"welcome","subject":"Hi"}`,
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			handler.Create(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("Create() status = %d, want %d. body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

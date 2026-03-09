package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPrincipal_HasScope(t *testing.T) {
	tests := []struct {
		name     string
		principal Principal
		required string
		want     bool
	}{
		{
			name:     "exact match",
			principal: Principal{Scopes: []string{"notifications:read"}},
			required: "notifications:read",
			want:     true,
		},
		{
			name:     "wildcard",
			principal: Principal{Scopes: []string{"*"}},
			required: "notifications:write",
			want:     true,
		},
		{
			name:     "prefix wildcard",
			principal: Principal{Scopes: []string{"notifications:*"}},
			required: "notifications:read",
			want:     true,
		},
		{
			name:     "no match",
			principal: Principal{Scopes: []string{"auth:read"}},
			required: "notifications:read",
			want:     false,
		},
		{
			name:     "empty scopes",
			principal: Principal{Scopes: []string{}},
			required: "notifications:read",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.principal.HasScope(tt.required)
			if got != tt.want {
				t.Errorf("HasScope(%q) = %v, want %v", tt.required, got, tt.want)
			}
		})
	}
}

func TestPrincipal_HasAnyScope(t *testing.T) {
	p := Principal{Scopes: []string{"notifications:read"}}
	if !p.HasAnyScope([]string{"notifications:read", "auth:admin"}) {
		t.Error("HasAnyScope should return true when principal has one of required")
	}
	if p.HasAnyScope([]string{"auth:admin", "templates:write"}) {
		t.Error("HasAnyScope should return false when principal has none of required")
	}
	if p.HasAnyScope([]string{}) {
		t.Error("HasAnyScope with empty required should return false (no scope to match)")
	}
}

func TestRequireAnyScope(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// No principal in context -> 401
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	RequireAnyScope("notifications:read")(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("no principal: got %d, want 401", rec.Code)
	}

	// Principal with scope -> 200
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(WithPrincipal(req.Context(), &Principal{Scopes: []string{"notifications:read"}}))
	rec = httptest.NewRecorder()
	RequireAnyScope("notifications:read")(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("with scope: got %d, want 200", rec.Code)
	}

	// Principal without scope -> 403
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(WithPrincipal(req.Context(), &Principal{Scopes: []string{"auth:read"}}))
	rec = httptest.NewRecorder()
	RequireAnyScope("notifications:write")(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("missing scope: got %d, want 403", rec.Code)
	}
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Pure function tests (no DB) ---

func TestParseKeyActionPath(t *testing.T) {
	tests := []struct {
		path   string
		keyID  string
		action string
		ok     bool
	}{
		{"/v1/keys/key_123/rotate", "key_123", "rotate", true},
		{"/v1/keys/key_123/revoke", "key_123", "revoke", true},
		{"/v1/keys/abc/rotate", "abc", "rotate", true},
		{"/v1/keys/", "", "", false},
		{"/v1/keys/key_123/invalid", "", "", false},
		{"/v1/keys/key_123", "", "", false},
		{"/v1/keys/key_123/rotate/extra", "", "", false},
		{"v1/keys/key_123/rotate", "key_123", "rotate", true},
	}
	for _, tt := range tests {
		keyID, action, ok := parseKeyActionPath(tt.path)
		if keyID != tt.keyID || action != tt.action || ok != tt.ok {
			t.Errorf("parseKeyActionPath(%q) = (%q, %q, %v), want (%q, %q, %v)",
				tt.path, keyID, action, ok, tt.keyID, tt.action, tt.ok)
		}
	}
}

func TestIsValidEnvironment(t *testing.T) {
	valid := []string{"dev", "staging", "prod"}
	for _, v := range valid {
		if !isValidEnvironment(v) {
			t.Errorf("isValidEnvironment(%q) = false, want true", v)
		}
	}
	invalid := []string{"", "development", "production", "test", "DEV"}
	for _, v := range invalid {
		if isValidEnvironment(v) {
			t.Errorf("isValidEnvironment(%q) = true, want false", v)
		}
	}
}

func TestActorFromPrincipal(t *testing.T) {
	tests := []struct {
		p      Principal
		actor  string
	}{
		{Principal{Owner: "alice"}, "alice"},
		{Principal{Kind: "admin", Owner: ""}, "admin"},
		{Principal{Kind: "issued", Owner: ""}, "system"},
		{Principal{Kind: "admin", Owner: "admin"}, "admin"},
	}
	for _, tt := range tests {
		got := actorFromPrincipal(tt.p)
		if got != tt.actor {
			t.Errorf("actorFromPrincipal(%+v) = %q, want %q", tt.p, got, tt.actor)
		}
	}
}

func TestScopeSet(t *testing.T) {
	scopes := []ServiceScope{
		{Service: "notifications", Scope: "read"},
		{Service: "notifications", Scope: "write"},
		{Service: "auth", Scope: "admin"},
	}
	out := scopeSet(scopes)
	if _, ok := out["notifications"]["read"]; !ok {
		t.Error("scopeSet: missing notifications/read")
	}
	if _, ok := out["notifications"]["write"]; !ok {
		t.Error("scopeSet: missing notifications/write")
	}
	if _, ok := out["auth"]["admin"]; !ok {
		t.Error("scopeSet: missing auth/admin")
	}
}

func TestPrincipalAllows(t *testing.T) {
	admin := Principal{Kind: "admin"}
	if !admin.Allows("notifications", "write", "prod") {
		t.Error("admin should allow any scope")
	}

	issued := Principal{
		Kind:        "issued",
		KeyID:       "key_1",
		Environment: "prod",
		ScopeSet: map[string]map[string]struct{}{
			"notifications": {"read": {}, "write": {}},
		},
	}
	if !issued.Allows("notifications", "read", "prod") {
		t.Error("issued should allow notifications/read in prod")
	}
	if !issued.Allows("notifications", "write", "") {
		t.Error("issued should allow when environment empty")
	}
	if issued.Allows("notifications", "admin", "prod") {
		t.Error("issued should not allow notifications/admin")
	}
	if issued.Allows("notifications", "read", "staging") {
		t.Error("issued should not allow when environment mismatch")
	}
}

func TestHashKey(t *testing.T) {
	h1 := hashKey("secret")
	h2 := hashKey("secret")
	if h1 != h2 {
		t.Errorf("hashKey should be deterministic: %q != %q", h1, h2)
	}
	if len(h1) != 64 {
		t.Errorf("SHA256 hex should be 64 chars, got %d", len(h1))
	}
}

func TestPrefix(t *testing.T) {
	if got := prefix("short", 10); got != "short" {
		t.Errorf("prefix(short, 10) = %q, want short", got)
	}
	if got := prefix("longerstring", 5); got != "longe" {
		t.Errorf("prefix(longerstring, 5) = %q, want longe", got)
	}
}

func TestValidateMintRequest(t *testing.T) {
	app := &App{catalog: defaultCatalog()}

	valid := MintKeyRequest{
		Application:   "myapp",
		Environment:   "prod",
		Owner:         "alice",
		ServiceScopes: []ServiceScope{{Service: "notifications", Scope: "read"}},
	}
	if err := app.validateMintRequest(valid); err != nil {
		t.Errorf("valid request should pass: %v", err)
	}

	missingApp := MintKeyRequest{Environment: "prod", Owner: "alice", ServiceScopes: []ServiceScope{{Service: "notifications", Scope: "read"}}}
	if err := app.validateMintRequest(missingApp); err == nil {
		t.Error("missing application should fail")
	}

	invalidEnv := MintKeyRequest{Application: "a", Environment: "invalid", Owner: "a", ServiceScopes: []ServiceScope{{Service: "notifications", Scope: "read"}}}
	if err := app.validateMintRequest(invalidEnv); err == nil {
		t.Error("invalid environment should fail")
	}

	emptyScopes := MintKeyRequest{Application: "a", Environment: "prod", Owner: "a", ServiceScopes: []ServiceScope{}}
	if err := app.validateMintRequest(emptyScopes); err == nil {
		t.Error("empty service_scopes should fail")
	}

	invalidServiceScope := MintKeyRequest{Application: "a", Environment: "prod", Owner: "a", ServiceScopes: []ServiceScope{{Service: "unknown", Scope: "read"}}}
	if err := app.validateMintRequest(invalidServiceScope); err == nil {
		t.Error("invalid service/scope should fail")
	}
}

func TestCatalogAllows(t *testing.T) {
	app := &App{catalog: defaultCatalog()}
	if !app.catalogAllows("notifications", "read") {
		t.Error("notifications/read should be allowed")
	}
	if !app.catalogAllows("auth", "admin") {
		t.Error("auth/admin should be allowed")
	}
	if app.catalogAllows("unknown", "read") {
		t.Error("unknown service should not be allowed")
	}
	if app.catalogAllows("notifications", "invalid") {
		t.Error("invalid scope should not be allowed")
	}
}

// --- HTTP handler tests (admin key auth, requireAdmin, handleKeys, handleKeyActions) ---
// These use a real DB; skip if unavailable so CI passes without Postgres.

func TestAdminKeyAuthAndRequireAdmin(t *testing.T) {
	pool, skip := testPool(t)
	if skip {
		return
	}
	defer pool.Close()

	app := &App{
		db:          pool,
		adminAPIKey: "test-admin-key-123",
		catalog:     defaultCatalog(),
	}
	ensureTestSchema(t, pool)

	// No X-API-Key -> 401
	req := httptest.NewRequest(http.MethodGet, "/v1/keys", nil)
	rec := httptest.NewRecorder()
	app.requireAdmin(app.handleKeys).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("no key: got status %d, want 401", rec.Code)
	}

	// Wrong key -> 401
	req = httptest.NewRequest(http.MethodGet, "/v1/keys", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	rec = httptest.NewRecorder()
	app.requireAdmin(app.handleKeys).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("wrong key: got status %d, want 401", rec.Code)
	}

	// Admin key -> 200 (list keys, may be empty)
	req = httptest.NewRequest(http.MethodGet, "/v1/keys", nil)
	req.Header.Set("X-API-Key", "test-admin-key-123")
	rec = httptest.NewRecorder()
	app.requireAdmin(app.handleKeys).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("admin key: got status %d, want 200", rec.Code)
	}
	var listResp struct {
		Keys []APIKey `json:"keys"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
}

func TestHandleKeysPost(t *testing.T) {
	pool, skip := testPool(t)
	if skip {
		return
	}
	defer pool.Close()

	app := &App{
		db:          pool,
		adminAPIKey: "test-admin-key-456",
		catalog:     defaultCatalog(),
	}
	ensureTestSchema(t, pool)

	body := `{"application":"testapp","environment":"dev","owner":"bob","service_scopes":[{"service":"notifications","scope":"read"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/keys", bytes.NewReader([]byte(body)))
	req.Header.Set("X-API-Key", "test-admin-key-456")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.requireAdmin(app.handleKeys).ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Errorf("mint key: got status %d, want 201. body: %s", rec.Code, rec.Body.String())
	}
	var mintResp struct {
		Key struct {
			ID       string `json:"id"`
			Key      string `json:"key"`
			KeyPrefix string `json:"key_prefix"`
		} `json:"key"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&mintResp); err != nil {
		t.Fatalf("decode mint response: %v", err)
	}
	if mintResp.Key.ID == "" || mintResp.Key.Key == "" || mintResp.Key.KeyPrefix == "" {
		t.Errorf("mint response missing key fields: %+v", mintResp)
	}
}

func TestHandleKeyActionsRotateRevoke(t *testing.T) {
	pool, skip := testPool(t)
	if skip {
		return
	}
	defer pool.Close()

	app := &App{
		db:          pool,
		adminAPIKey: "test-admin-789",
		catalog:     defaultCatalog(),
	}
	ensureTestSchema(t, pool)

	// Mint a key first
	mintBody := `{"application":"testapp","environment":"dev","owner":"carol","service_scopes":[{"service":"notifications","scope":"read"}]}`
	mintReq := httptest.NewRequest(http.MethodPost, "/v1/keys", bytes.NewReader([]byte(mintBody)))
	mintReq.Header.Set("X-API-Key", "test-admin-789")
	mintReq.Header.Set("Content-Type", "application/json")
	mintRec := httptest.NewRecorder()
	app.requireAdmin(app.handleKeys).ServeHTTP(mintRec, mintReq)
	if mintRec.Code != http.StatusCreated {
		t.Fatalf("mint failed: %d %s", mintRec.Code, mintRec.Body.String())
	}
	var mintResp struct {
		Key struct {
			ID string `json:"id"`
		} `json:"key"`
	}
	if err := json.NewDecoder(mintRec.Body).Decode(&mintResp); err != nil {
		t.Fatalf("decode mint: %v", err)
	}
	keyID := mintResp.Key.ID
	if keyID == "" {
		t.Fatal("mint returned empty key id")
	}

	// Rotate
	rotateReq := httptest.NewRequest(http.MethodPost, "/v1/keys/"+keyID+"/rotate", nil)
	rotateReq.Header.Set("X-API-Key", "test-admin-789")
	rotateRec := httptest.NewRecorder()
	app.requireAdmin(app.handleKeyActions).ServeHTTP(rotateRec, rotateReq)
	if rotateRec.Code != http.StatusOK {
		t.Errorf("rotate: got %d, want 200. body: %s", rotateRec.Code, rotateRec.Body.String())
	}

	// Revoke (revoke the new rotated key)
	var rotateResp struct {
		Key struct {
			ID string `json:"id"`
		} `json:"key"`
	}
	if err := json.NewDecoder(rotateRec.Body).Decode(&rotateResp); err != nil {
		t.Fatalf("decode rotate: %v", err)
	}
	rotatedID := rotateResp.Key.ID

	revokeReq := httptest.NewRequest(http.MethodPost, "/v1/keys/"+rotatedID+"/revoke", nil)
	revokeReq.Header.Set("X-API-Key", "test-admin-789")
	revokeRec := httptest.NewRecorder()
	app.requireAdmin(app.handleKeyActions).ServeHTTP(revokeRec, revokeReq)
	if revokeRec.Code != http.StatusOK {
		t.Errorf("revoke: got %d, want 200. body: %s", revokeRec.Code, revokeRec.Body.String())
	}
}

// testPool creates a DB pool for integration tests. Skips if DB unavailable (e.g. CI without Postgres).
func testPool(t *testing.T) (*pgxpool.Pool, bool) {
	t.Helper()
	dbURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if dbURL == "" {
		dbURL = defaultDatabaseURL
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("skip: no test database: %v", err)
		return nil, true
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("skip: cannot connect to test database: %v", err)
		return nil, true
	}
	return pool, false
}

func ensureTestSchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	if err := ensureSchema(ctx, pool); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}
}

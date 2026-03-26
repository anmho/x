package keys

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// KeyRecord represents a stored API key.
type KeyRecord struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key"`
	CreatedAt string `json:"created_at"`
}

type keyStore struct {
	Keys []KeyRecord `json:"keys"`
}

// DefaultStorePath returns the default path for the key store.
func DefaultStorePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".x-mcp", "keys.json")
}

// Load loads key records from the given file. Returns an empty slice if the file does not exist.
func Load(path string) ([]KeyRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []KeyRecord{}, nil
		}
		return nil, fmt.Errorf("keys.Load: %w", err)
	}
	var store keyStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("keys.Load: %w", err)
	}
	return store.Keys, nil
}

// Save writes key records atomically to the given path.
func Save(path string, records []KeyRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("keys.Save: %w", err)
	}
	store := keyStore{Keys: records}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("keys.Save: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("keys.Save: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("keys.Save: %w", err)
	}
	return nil
}

// Generate creates a new key with the given name, saves it, and returns the record.
func Generate(path, name string) (KeyRecord, error) {
	records, err := Load(path)
	if err != nil {
		return KeyRecord{}, fmt.Errorf("keys.Generate: %w", err)
	}

	// Generate 24 random bytes → 48 hex chars
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return KeyRecord{}, fmt.Errorf("keys.Generate: %w", err)
	}
	idRaw := make([]byte, 8)
	if _, err := rand.Read(idRaw); err != nil {
		return KeyRecord{}, fmt.Errorf("keys.Generate: %w", err)
	}

	rec := KeyRecord{
		ID:        hex.EncodeToString(idRaw),
		Name:      name,
		Key:       "mcp_" + hex.EncodeToString(raw),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	records = append(records, rec)
	if err := Save(path, records); err != nil {
		return KeyRecord{}, fmt.Errorf("keys.Generate: %w", err)
	}
	return rec, nil
}

// Configured returns all API keys available via env or the local store.
func Configured(path string) ([]string, error) {
	keys := configuredFromEnv()

	records, err := Load(path)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		if value := strings.TrimSpace(record.Key); value != "" {
			keys = append(keys, value)
		}
	}

	return keys, nil
}

// Validate returns true if the given key exists in the store.
func Validate(path, key string) bool {
	key = strings.TrimSpace(key)
	if key == "" {
		return false
	}

	configured, err := Configured(path)
	if err != nil {
		return false
	}
	for _, candidate := range configured {
		if secureEqual(candidate, key) {
			return true
		}
	}
	return false
}

// Revoke removes the key with the given ID. Returns (found, error).
func Revoke(path, id string) (bool, error) {
	records, err := Load(path)
	if err != nil {
		return false, fmt.Errorf("keys.Revoke: %w", err)
	}
	found := false
	updated := make([]KeyRecord, 0, len(records))
	for _, r := range records {
		if r.ID == id {
			found = true
			continue
		}
		updated = append(updated, r)
	}
	if found {
		if err := Save(path, updated); err != nil {
			return false, fmt.Errorf("keys.Revoke: %w", err)
		}
	}
	return found, nil
}

func configuredFromEnv() []string {
	raw := strings.TrimSpace(os.Getenv("MCP_API_KEYS"))
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	keys := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			keys = append(keys, value)
		}
	}
	return keys
}

func secureEqual(left, right string) bool {
	if len(left) != len(right) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

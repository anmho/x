package repository

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/andrewho/omnichannel/internal/database"
	"github.com/andrewho/omnichannel/internal/domain"
	"github.com/google/uuid"
)

func TestTemplateRepository_CreateAndGetFull(t *testing.T) {
	dbURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if dbURL == "" {
		dbURL = "postgresql://postgres:postgres@127.0.0.1:5432/omnichannel?sslmode=disable"
	}
	ctx := context.Background()
	db, err := database.New(ctx, dbURL)
	if err != nil {
		t.Skipf("skip: no test database: %v", err)
		return
	}
	defer db.Close()

	repo := NewTemplateRepository(db)
	template := &domain.Template{
		ID:          uuid.New(),
		Name:        "test-template-" + uuid.New().String(),
		Description: "Test template",
		Subject:     "Test Subject",
		Body:        "Hello {{name}}",
		Variables:   []string{"name"},
		IsActive:    true,
	}

	if err := repo.Create(ctx, template); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetFull(ctx, template.ID)
	if err != nil {
		t.Fatalf("GetFull: %v", err)
	}
	if got.Name != template.Name || got.Subject != template.Subject || got.Body != template.Body {
		t.Errorf("GetFull: got %+v, want name=%q subject=%q body=%q", got, template.Name, template.Subject, template.Body)
	}
}

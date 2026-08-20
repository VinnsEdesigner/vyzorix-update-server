package serviceaccount

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
)

func setupServiceAccountTestDB(t *testing.T) *storage.SQLite {
	t.Helper()
	cfg := storage.DefaultConfig(filepath.Join(t.TempDir(), "sa-test.db"))
	s, err := storage.Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestService_CRUD(t *testing.T) {
	s := setupServiceAccountTestDB(t)
	service := NewService(
		storage.NewServiceAccountRepository(s.DB()),
		storage.NewServiceAccountTokenRepository(s.DB()),
	)
	ctx := context.Background()

	sa, err := service.Create(ctx, "org-1", "ci-deployer")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if sa.ID == "" {
		t.Fatal("expected generated ID")
	}

	accounts, err := service.List(ctx, "org-1")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(accounts) != 1 {
		t.Errorf("List returned %d, want 1", len(accounts))
	}

	// Cross-org hidden.
	if _, err := service.Get(ctx, "org-2", sa.ID); err != ErrServiceAccountNotFound {
		t.Errorf("cross-org Get: %v", err)
	}

	if err := service.Delete(ctx, "org-1", sa.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := service.Delete(ctx, "org-1", sa.ID); err != ErrServiceAccountNotFound {
		t.Errorf("second Delete: %v", err)
	}
}

func TestService_TokenLifecycle(t *testing.T) {
	s := setupServiceAccountTestDB(t)
	service := NewService(
		storage.NewServiceAccountRepository(s.DB()),
		storage.NewServiceAccountTokenRepository(s.DB()),
	)
	ctx := context.Background()

	sa, err := service.Create(ctx, "org-1", "ci-deployer")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	token, fullKey, err := service.CreateToken(ctx, &TokenInput{
		ServiceID: sa.ID,
		Name:      "default",
		Scopes:    []string{"read"},
	})
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	if fullKey == "" || token.KeyPrefix == "" {
		t.Error("expected full key and prefix")
	}
	if !token.IsUsable() {
		t.Error("expected usable token")
	}

	// Auth validates the key.
	validated, err := service.Authenticate(ctx, fullKey)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if validated.ID != token.ID {
		t.Errorf("expected token %s, got %s", token.ID, validated.ID)
	}

	// Invalid key rejected.
	if _, err := service.Authenticate(ctx, "bogus"); err == nil {
		t.Error("expected auth failure for bogus key")
	}

	// Rotate: old revoked, new usable.
	newToken, newKey, err := service.RotateToken(ctx, token.ID, &TokenInput{
		ServiceID: sa.ID,
		Scopes:    []string{"write"},
	})
	if err != nil {
		t.Fatalf("RotateToken: %v", err)
	}
	if newToken.ID == token.ID {
		t.Error("rotate should create a new token")
	}
	if !newToken.IsUsable() {
		t.Error("expected usable rotated token")
	}
	if _, err := service.Authenticate(ctx, newKey); err != nil {
		t.Errorf("new key auth: %v", err)
	}

	// Old key no longer authenticates.
	if _, err := service.Authenticate(ctx, fullKey); err == nil {
		t.Error("old key should be revoked")
	}

	// Full key never returned again.
	tokens, _ := service.ListTokens(ctx, sa.ID)
	if tokens[0].KeyHash == fullKey {
		t.Error("hash must not be the full key")
	}
}

func TestSecretscan(t *testing.T) {
	s := setupServiceAccountTestDB(t)
	service := NewService(
		storage.NewServiceAccountRepository(s.DB()),
		storage.NewServiceAccountTokenRepository(s.DB()),
	)
	ctx := context.Background()

	sa, err := service.Create(ctx, "org-1", "ci-deployer")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	token, _, err := service.CreateToken(ctx, &TokenInput{
		ServiceID: sa.ID,
		Name:      "default",
		Scopes:    []string{"read"},
	})
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}

	leaks, err := service.Secretscan(ctx, "org-1", []byte("key is "+token.KeyPrefix))
	if err != nil {
		t.Fatalf("Secretscan: %v", err)
	}
	if leaks != 1 {
		t.Errorf("expected 1 leak, got %d", leaks)
	}

	// Clean payload has no leak.
	leaks, _ = service.Secretscan(ctx, "org-1", []byte("nothing to see"))
	if leaks != 0 {
		t.Errorf("expected 0 leaks, got %d", leaks)
	}
}

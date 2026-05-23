package identity

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStoreSeedMembershipAndRoleAssignment(t *testing.T) {
	store, err := NewStore(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	members := store.Directory(DefaultWorkspaceID)
	if len(members) != 10 {
		t.Fatalf("expected 10 seeded members, got %d", len(members))
	}
	member, err := store.DirectoryMember(DefaultWorkspaceID, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if member.Roles[0] != "project_owner" {
		t.Fatalf("expected alice project_owner, got %v", member.Roles)
	}
	if err := store.AssignRole(DefaultWorkspaceID, "alice", "admin", "tester"); err != nil {
		t.Fatal(err)
	}
	member, _ = store.DirectoryMember(DefaultWorkspaceID, "alice")
	if member.Roles[0] != "admin" {
		t.Fatalf("expected updated role, got %v", member.Roles)
	}
	if len(store.AuditEvents()) == 0 {
		t.Fatal("expected role change audit event")
	}
}

func TestDisabledAccountCannotAuthenticate(t *testing.T) {
	store, _ := NewStore(context.Background(), "")
	if _, err := store.Authenticate("alice", "password"); err != nil {
		t.Fatal(err)
	}
	if err := store.DisableUser("alice"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Authenticate("alice", "password"); err != ErrDisabledAccount {
		t.Fatalf("expected disabled account, got %v", err)
	}
}

func TestBuiltInProviderAndExternalIdentityMapping(t *testing.T) {
	store, _ := NewStore(context.Background(), "")
	provider := NewBuiltInProvider(store)
	user, err := provider.Authenticate(context.Background(), "tester", "password")
	if err != nil {
		t.Fatal(err)
	}
	if user.Provider != ProviderBuiltIn || user.ExternalSubject != "tester" {
		t.Fatalf("unexpected provider user: %+v", user)
	}
	if err := store.MapExternalIdentity("oidc", "subject-1", user.User.ID); err != nil {
		t.Fatal(err)
	}
	if len(store.AuditEvents()) == 0 {
		t.Fatal("expected external identity audit event")
	}
}

func TestTokenIssueValidateAndExpiry(t *testing.T) {
	now := time.Now().UTC()
	claims := Claims{Sub: "alice", WorkspaceID: DefaultWorkspaceID, Roles: []string{"project_owner"}, DisplayName: "alice", Email: "alice@example.local", Provider: ProviderBuiltIn, Version: 1, Iat: now.Unix(), Exp: now.Add(time.Minute).Unix()}
	token, err := IssueToken("secret", claims)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ValidateToken("secret", token, now)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Sub != "alice" || parsed.Roles[0] != "project_owner" {
		t.Fatalf("unexpected claims: %+v", parsed)
	}
	if _, err := ValidateToken("secret", token, now.Add(2*time.Minute)); err == nil {
		t.Fatal("expected expired token error")
	}
}

func TestLoginIntrospectAndJWKSHandlers(t *testing.T) {
	server, err := NewServer(context.Background(), Config{TokenSecret: "secret", AppEnv: "development", AuthMode: "dev-token", TokenTTL: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(LoginRequest{Username: "alice", Password: "password", WorkspaceID: DefaultWorkspaceID})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	var login LoginResponse
	json.Unmarshal(resp.Body.Bytes(), &login)
	if login.AccessToken == "" || login.User.ID != "alice" {
		t.Fatalf("unexpected login response: %+v", login)
	}

	introspectBody, _ := json.Marshal(map[string]string{"token": login.AccessToken})
	introspectReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/introspect", bytes.NewReader(introspectBody))
	introspectResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(introspectResp, introspectReq)
	if introspectResp.Code != http.StatusOK || !bytes.Contains(introspectResp.Body.Bytes(), []byte(`"active":true`)) {
		t.Fatalf("expected active introspection, got %d body=%s", introspectResp.Code, introspectResp.Body.String())
	}

	jwksReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/jwks", nil)
	jwksResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(jwksResp, jwksReq)
	if jwksResp.Code != http.StatusOK || !bytes.Contains(jwksResp.Body.Bytes(), []byte("dev-hmac-1")) {
		t.Fatalf("expected signing metadata, got %d body=%s", jwksResp.Code, jwksResp.Body.String())
	}
}

func TestLoginHandlerRejectsInvalidAndDisabledUsers(t *testing.T) {
	server, err := NewServer(context.Background(), Config{TokenSecret: "secret", AppEnv: "development", AuthMode: "dev-token", TokenTTL: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	badBody, _ := json.Marshal(LoginRequest{Username: "alice", Password: "wrong", WorkspaceID: DefaultWorkspaceID})
	badReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(badBody))
	badResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(badResp, badReq)
	if badResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected invalid login rejected, got %d", badResp.Code)
	}

	if err := server.store.DisableUser("alice"); err != nil {
		t.Fatal(err)
	}
	disabledBody, _ := json.Marshal(LoginRequest{Username: "alice", Password: "password", WorkspaceID: DefaultWorkspaceID})
	disabledReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(disabledBody))
	disabledResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(disabledResp, disabledReq)
	if disabledResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected disabled login rejected, got %d", disabledResp.Code)
	}
	if len(server.store.AuditEvents()) == 0 {
		t.Fatal("expected failed login audit events")
	}
}

func TestProductionRejectsDevelopmentIdentityMode(t *testing.T) {
	_, err := NewServer(context.Background(), Config{TokenSecret: "secret", AppEnv: "production", AuthMode: "dev-token"})
	if err != ErrForbidden {
		t.Fatalf("expected forbidden production dev-token mode, got %v", err)
	}
}

func TestExternalProviderConfigurationPlaceholders(t *testing.T) {
	cfg := Config{Provider: ProviderConfig{BuiltInEnabled: true, OIDCIssuer: "https://issuer.example", LDAPURL: "ldap://directory.example"}}
	if !cfg.Provider.BuiltInEnabled || cfg.Provider.OIDCIssuer == "" || cfg.Provider.LDAPURL == "" {
		t.Fatalf("expected provider placeholders to be configurable: %+v", cfg.Provider)
	}
}

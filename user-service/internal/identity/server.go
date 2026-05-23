package identity

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	ListenAddr  string
	MySQLDSN    string
	TokenSecret string
	AppEnv      string
	AuthMode    string
	TokenTTL    time.Duration
	Provider    ProviderConfig
}

func LoadConfig() Config {
	return Config{ListenAddr: getenv("USER_SERVICE_LISTEN_ADDR", ":8090"), MySQLDSN: getenv("USER_MYSQL_DSN", ""), TokenSecret: getenv("AUTH_TOKEN_SECRET", "dev-secret-change-me"), AppEnv: getenv("APP_ENV", "development"), AuthMode: getenv("AUTH_MODE", "dev-token"), TokenTTL: durationFromEnv("AUTH_TOKEN_TTL", time.Hour), Provider: ProviderConfig{BuiltInEnabled: true, OIDCIssuer: os.Getenv("OIDC_ISSUER"), LDAPURL: os.Getenv("LDAP_URL")}}
}

func (cfg Config) Validate() error {
	if cfg.AppEnv == "production" && (cfg.AuthMode == "dev-header" || cfg.AuthMode == "dev-token") {
		return ErrForbidden
	}
	return nil
}

type Server struct {
	cfg      Config
	store    *Store
	provider IdentityProvider
	mux      *http.ServeMux
	now      func() time.Time
}

func NewServer(ctx context.Context, cfg Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	store, err := NewStore(ctx, cfg.MySQLDSN)
	if err != nil {
		return nil, err
	}
	s := &Server{cfg: cfg, store: store, provider: NewBuiltInProvider(store), mux: http.NewServeMux(), now: func() time.Time { return time.Now().UTC() }}
	s.routes()
	return s, nil
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) routes() {
	s.mux.HandleFunc("/api/v1/health", s.handleHealth)
	s.mux.HandleFunc("/api/v1/users", s.handleUsers)
	s.mux.HandleFunc("/api/v1/users/manage", s.handleUserManage)
	s.mux.HandleFunc("/api/v1/users/update", s.handleUserUpdate)
	s.mux.HandleFunc("/api/v1/users/disable", s.handleUserDisable)
	s.mux.HandleFunc("/api/v1/users/enable", s.handleUserEnable)
	s.mux.HandleFunc("/api/v1/users/role", s.handleRoleAssign)
	s.mux.HandleFunc("/api/v1/auth/login", s.handleLogin)
	s.mux.HandleFunc("/api/v1/auth/introspect", s.handleIntrospect)
	s.mux.HandleFunc("/api/v1/auth/jwks", s.handleJWKS)
	s.mux.HandleFunc("/api/v1/dev-token", s.handleDevToken)
	s.mux.HandleFunc("/api/v1/audit-events", s.handleAuditEvents)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "storageBackend": s.store.StorageBackend(), "authMode": s.cfg.AuthMode})
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	workspaceID := r.URL.Query().Get("workspaceId")
	query := r.URL.Query().Get("q")
	if query != "" {
		writeJSON(w, http.StatusOK, s.store.SearchUsers(query))
		return
	}
	writeJSON(w, http.StatusOK, s.store.Directory(workspaceID))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	var req LoginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	workspaceID := req.WorkspaceID
	if workspaceID == "" {
		workspaceID = DefaultWorkspaceID
	}
	providerUser, err := s.provider.Authenticate(r.Context(), strings.TrimSpace(req.Username), req.Password)
	if err != nil {
		s.store.auditLoginFailure(req.Username, err)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}
	member, err := s.store.DirectoryMember(workspaceID, providerUser.User.ID)
	if err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	claims := ClaimsFromMember(providerUser.User, member, workspaceID, s.cfg.TokenTTL, s.now())
	token, err := IssueToken(s.cfg.TokenSecret, claims)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, LoginResponse{AccessToken: token, TokenType: "Bearer", ExpiresIn: int64(s.cfg.TokenTTL.Seconds()), User: member})
}

func (s *Store) auditLoginFailure(username string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.auditLocked("auth.login_failed", username, username, map[string]string{"reason": err.Error()})
}

func (s *Server) handleIntrospect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	var req struct {
		Token string `json:"token"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	claims, err := ValidateToken(s.cfg.TokenSecret, req.Token, s.now())
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"active": false, "error": err.Error()})
		return
	}
	member, err := s.store.Membership(claims.WorkspaceID, claims.Sub)
	if err != nil || member.Status != MembershipActive || member.Version > claims.Version {
		writeJSON(w, http.StatusOK, map[string]interface{}{"active": false, "error": "membership inactive or token version stale"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"active": true, "claims": claims})
}

func (s *Server) handleJWKS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, SigningMetadata())
}

func (s *Server) handleDevToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if s.cfg.AppEnv == "production" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "development tokens are disabled in production"})
		return
	}
	var req struct {
		UserID      string `json:"userId"`
		WorkspaceID string `json:"workspaceId"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	workspaceID := req.WorkspaceID
	if workspaceID == "" {
		workspaceID = DefaultWorkspaceID
	}
	member, err := s.store.DirectoryMember(workspaceID, req.UserID)
	if err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	user, err := s.store.UserByID(req.UserID)
	if err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	claims := ClaimsFromMember(user, member, workspaceID, s.cfg.TokenTTL, s.now())
	token, err := IssueToken(s.cfg.TokenSecret, claims)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, LoginResponse{AccessToken: token, TokenType: "Bearer", ExpiresIn: int64(s.cfg.TokenTTL.Seconds()), User: member})
}

func (s *Server) handleAuditEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, s.store.AuditEvents())
}

func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if s.cfg.AuthMode == "dev-header" || s.cfg.AuthMode == "dev-token" {
		role := r.Header.Get("X-Role")
		if role != "admin" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin required"})
			return false
		}
		return true
	}
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin required"})
		return false
	}
	claims, err := ValidateToken(s.cfg.TokenSecret, strings.TrimPrefix(auth, "Bearer "), s.now())
	if err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return false
	}
	hasAdmin := false
	for _, role := range claims.Roles {
		if role == "admin" {
			hasAdmin = true
			break
		}
	}
	if !hasAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin required"})
		return false
	}
	return true
}

func (s *Server) handleUserManage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if !s.requireAdmin(w, r) {
		return
	}
	var req CreateUserRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	user, err := s.store.CreateUser(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

func (s *Server) handleUserUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.NotFound(w, r)
		return
	}
	if !s.requireAdmin(w, r) {
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}
	var req UpdateUserRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	user, err := s.store.UpdateUser(id, req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) handleUserDisable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if !s.requireAdmin(w, r) {
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}
	if err := s.store.DisableUser(id); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "disabled"})
}

func (s *Server) handleUserEnable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if !s.requireAdmin(w, r) {
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}
	if err := s.store.EnableUser(id); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "enabled"})
}

func (s *Server) handleRoleAssign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.NotFound(w, r)
		return
	}
	if !s.requireAdmin(w, r) {
		return
	}
	var req RoleAssignRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.UserID == "" || req.Role == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "userId and role are required"})
		return
	}
	if err := s.store.AssignRole(req.WorkspaceID, req.UserID, req.Role, "admin"); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "role assigned"})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func durationFromEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("invalid %s=%q: %v", key, value, err)
		return fallback
	}
	return parsed
}

package identity

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Store struct {
	mu          sync.RWMutex
	db          *sql.DB
	users       map[string]User
	byUsername  map[string]string
	workspaces  map[string]Workspace
	memberships map[string]Membership
	externals   map[string]ExternalIdentity
	auditEvents []AuditEvent
	now         func() time.Time
}

func NewStore(ctx context.Context, dsn string) (*Store, error) {
	store := &Store{
		users:       map[string]User{},
		byUsername:  map[string]string{},
		workspaces:  map[string]Workspace{},
		memberships: map[string]Membership{},
		externals:   map[string]ExternalIdentity{},
		now:         func() time.Time { return time.Now().UTC() },
	}
	if dsn != "" {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, err
		}
		if err := waitForMySQL(ctx, db, 60*time.Second); err != nil {
			db.Close()
			return nil, err
		}
		store.db = db
		if err := store.EnsureSchema(ctx); err != nil {
			db.Close()
			return nil, err
		}
	}
	store.SeedDevelopmentData()
	return store, nil
}

func waitForMySQL(ctx context.Context, db *sql.DB, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		if err := db.PingContext(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("mysql not ready after %s: %w", timeout, lastErr)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

func (s *Store) EnsureSchema(ctx context.Context) error {
	if s.db == nil {
		return nil
	}
	for _, stmt := range schemaStatements {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("ensure user-service schema: %w", err)
		}
	}
	return nil
}

var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS users (id VARCHAR(64) PRIMARY KEY, username VARCHAR(128) NOT NULL UNIQUE, display_name VARCHAR(255) NOT NULL, email VARCHAR(255) NOT NULL, status VARCHAR(32) NOT NULL, provider VARCHAR(64) NOT NULL, password_hash VARCHAR(255) NOT NULL DEFAULT '', token_version BIGINT NOT NULL DEFAULT 1, created_at DATETIME NOT NULL, updated_at DATETIME NOT NULL)`,
	`CREATE TABLE IF NOT EXISTS workspaces (id VARCHAR(64) PRIMARY KEY, name VARCHAR(255) NOT NULL, created_at DATETIME NOT NULL, updated_at DATETIME NOT NULL)`,
	`CREATE TABLE IF NOT EXISTS memberships (workspace_id VARCHAR(64) NOT NULL, user_id VARCHAR(64) NOT NULL, role VARCHAR(64) NOT NULL, status VARCHAR(32) NOT NULL, version BIGINT NOT NULL DEFAULT 1, created_at DATETIME NOT NULL, updated_at DATETIME NOT NULL, PRIMARY KEY (workspace_id, user_id))`,
	`CREATE TABLE IF NOT EXISTS external_identities (provider VARCHAR(64) NOT NULL, external_subject VARCHAR(255) NOT NULL, user_id VARCHAR(64) NOT NULL, created_at DATETIME NOT NULL, updated_at DATETIME NOT NULL, PRIMARY KEY (provider, external_subject))`,
	`CREATE TABLE IF NOT EXISTS sessions (id VARCHAR(64) PRIMARY KEY, user_id VARCHAR(64) NOT NULL, workspace_id VARCHAR(64) NOT NULL, provider VARCHAR(64) NOT NULL, revoked BOOLEAN NOT NULL DEFAULT FALSE, expires_at DATETIME NOT NULL, created_at DATETIME NOT NULL)`,
	`CREATE TABLE IF NOT EXISTS signing_keys (id VARCHAR(64) PRIMARY KEY, alg VARCHAR(32) NOT NULL, status VARCHAR(32) NOT NULL, created_at DATETIME NOT NULL, rotated_at DATETIME NULL)`,
	`CREATE TABLE IF NOT EXISTS audit_events (id VARCHAR(64) PRIMARY KEY, event_type VARCHAR(128) NOT NULL, actor_id VARCHAR(64) NOT NULL, target_id VARCHAR(64) NOT NULL, metadata JSON NOT NULL, created_at DATETIME NOT NULL)`,
}

func (s *Store) StorageBackend() string {
	if s.db != nil {
		return "mysql"
	}
	return "memory"
}

func (s *Store) SeedDevelopmentData() {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.workspaces[DefaultWorkspaceID]; !ok {
		s.workspaces[DefaultWorkspaceID] = Workspace{ID: DefaultWorkspaceID, Name: "Default Workspace", CreatedAt: now, UpdatedAt: now}
	}
	seed := []struct{ username, role string }{
		{"admin", "admin"},
		{"tester", "admin"},
		{"alice", "project_owner"},
		{"bob", "contributor"},
		{"carol", "viewer"},
		{"frontend-user", "portfolio_manager"},
		{"leader1", "project_owner"},
		{"leader2", "project_owner"},
		{"eng1", "contributor"},
		{"eng2", "contributor"},
	}
	for _, item := range seed {
		id := item.username
		if _, ok := s.users[id]; !ok {
			s.users[id] = User{ID: id, Username: item.username, DisplayName: item.username, Email: item.username + "@example.local", Status: StatusActive, Provider: ProviderBuiltIn, PasswordHash: hashPassword("password"), TokenVersion: 1, CreatedAt: now, UpdatedAt: now}
			s.byUsername[item.username] = id
		}
		key := membershipKey(DefaultWorkspaceID, id)
		if _, ok := s.memberships[key]; !ok {
			s.memberships[key] = Membership{WorkspaceID: DefaultWorkspaceID, UserID: id, Role: item.role, Status: MembershipActive, Version: 1, CreatedAt: now, UpdatedAt: now}
		}
	}
}

func hashPassword(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (s *Store) Authenticate(username, password string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byUsername[username]
	if !ok {
		return User{}, ErrInvalidCredentials
	}
	user := s.users[id]
	if user.Status != StatusActive {
		return User{}, ErrDisabledAccount
	}
	if user.PasswordHash != hashPassword(password) {
		return User{}, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Store) Directory(workspaceID string) []DirectoryMember {
	if workspaceID == "" {
		workspaceID = DefaultWorkspaceID
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	members := []DirectoryMember{}
	for _, membership := range s.memberships {
		if membership.WorkspaceID != workspaceID || membership.Status != MembershipActive {
			continue
		}
		user, ok := s.users[membership.UserID]
		if !ok || user.Status != StatusActive {
			continue
		}
		members = append(members, DirectoryMember{ID: user.ID, Username: user.Username, DisplayName: user.DisplayName, Email: user.Email, Status: user.Status, Roles: []string{membership.Role}})
	}
	sort.Slice(members, func(i, j int) bool { return members[i].ID < members[j].ID })
	return members
}

func (s *Store) DirectoryMember(workspaceID, userID string) (DirectoryMember, error) {
	if workspaceID == "" {
		workspaceID = DefaultWorkspaceID
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[userID]
	if !ok {
		return DirectoryMember{}, ErrNotFound
	}
	membership, ok := s.memberships[membershipKey(workspaceID, userID)]
	if !ok || membership.Status != MembershipActive || user.Status != StatusActive {
		return DirectoryMember{}, ErrForbidden
	}
	return DirectoryMember{ID: user.ID, Username: user.Username, DisplayName: user.DisplayName, Email: user.Email, Status: user.Status, Roles: []string{membership.Role}}, nil
}

func (s *Store) UserByID(userID string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[userID]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (s *Store) Membership(workspaceID, userID string) (Membership, error) {
	if workspaceID == "" {
		workspaceID = DefaultWorkspaceID
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	membership, ok := s.memberships[membershipKey(workspaceID, userID)]
	if !ok {
		return Membership{}, ErrNotFound
	}
	return membership, nil
}

func (s *Store) DisableUser(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[userID]
	if !ok {
		return ErrNotFound
	}
	user.Status = StatusDisabled
	user.TokenVersion++
	user.UpdatedAt = s.now()
	s.users[userID] = user
	s.auditLocked("user.disabled", "system", userID, map[string]string{"status": StatusDisabled})
	return nil
}

func (s *Store) EnableUser(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[userID]
	if !ok {
		return ErrNotFound
	}
	user.Status = StatusActive
	user.TokenVersion++
	user.UpdatedAt = s.now()
	s.users[userID] = user
	s.auditLocked("user.enabled", "system", userID, map[string]string{"status": StatusActive})
	return nil
}

func (s *Store) CreateUser(req CreateUserRequest) (User, error) {
	if req.Username == "" {
		return User{}, ErrInvalidCredentials
	}
	role := req.Role
	if role == "" {
		role = "contributor"
	}
	workspaceID := req.WorkspaceID
	if workspaceID == "" {
		workspaceID = DefaultWorkspaceID
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byUsername[req.Username]; ok {
		return User{}, ErrForbidden
	}
	now := s.now()
	id := req.Username
	user := User{
		ID:           id,
		Username:     req.Username,
		DisplayName:  req.DisplayName,
		Email:        req.Email,
		Status:       StatusActive,
		Provider:     ProviderBuiltIn,
		PasswordHash: hashPassword(req.Password),
		TokenVersion: 1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.users[id] = user
	s.byUsername[req.Username] = id
	key := membershipKey(workspaceID, id)
	s.memberships[key] = Membership{
		WorkspaceID: workspaceID,
		UserID:      id,
		Role:        role,
		Status:      MembershipActive,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.auditLocked("user.created", "admin", id, map[string]string{"username": req.Username, "role": role, "workspaceId": workspaceID})
	return user, nil
}

func (s *Store) UpdateUser(userID string, req UpdateUserRequest) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[userID]
	if !ok {
		return User{}, ErrNotFound
	}
	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	user.UpdatedAt = s.now()
	s.users[userID] = user
	s.auditLocked("user.updated", "admin", userID, map[string]string{})
	return user, nil
}

func (s *Store) SearchUsers(query string) []DirectoryMember {
	all := s.Directory(DefaultWorkspaceID)
	if query == "" {
		return all
	}
	lower := strings.ToLower(query)
	filtered := make([]DirectoryMember, 0, len(all))
	for _, m := range all {
		if strings.Contains(strings.ToLower(m.DisplayName), lower) ||
			strings.Contains(strings.ToLower(m.Username), lower) ||
			strings.Contains(strings.ToLower(m.Email), lower) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func (s *Store) AssignRole(workspaceID, userID, role, actorID string) error {
	if workspaceID == "" {
		workspaceID = DefaultWorkspaceID
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	membership, ok := s.memberships[membershipKey(workspaceID, userID)]
	if !ok {
		return ErrNotFound
	}
	previous := membership.Role
	membership.Role = role
	membership.Version++
	membership.UpdatedAt = s.now()
	s.memberships[membershipKey(workspaceID, userID)] = membership
	s.auditLocked("membership.role_changed", actorID, userID, map[string]string{"workspaceId": workspaceID, "previousRole": previous, "newRole": role})
	return nil
}

func (s *Store) RemoveMember(workspaceID, userID, actorID string) error {
	if workspaceID == "" {
		workspaceID = DefaultWorkspaceID
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	membership, ok := s.memberships[membershipKey(workspaceID, userID)]
	if !ok {
		return ErrNotFound
	}
	membership.Status = MembershipRemoved
	membership.Version++
	membership.UpdatedAt = s.now()
	s.memberships[membershipKey(workspaceID, userID)] = membership
	s.auditLocked("membership.removed", actorID, userID, map[string]string{"workspaceId": workspaceID})
	return nil
}

func (s *Store) MapExternalIdentity(provider, subject, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[userID]; !ok {
		return ErrNotFound
	}
	key := provider + ":" + subject
	now := s.now()
	if existing, ok := s.externals[key]; ok && existing.UserID != userID {
		s.auditLocked("external_identity.conflict", "system", userID, map[string]string{"provider": provider, "externalSubject": subject, "existingUserId": existing.UserID})
		return ErrForbidden
	}
	s.externals[key] = ExternalIdentity{Provider: provider, ExternalSubject: subject, UserID: userID, CreatedAt: now, UpdatedAt: now}
	s.auditLocked("external_identity.mapped", "system", userID, map[string]string{"provider": provider, "externalSubject": subject})
	return nil
}

func (s *Store) AuditEvents() []AuditEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := append([]AuditEvent{}, s.auditEvents...)
	return items
}

func (s *Store) auditLocked(eventType, actorID, targetID string, metadata map[string]string) {
	s.auditEvents = append(s.auditEvents, AuditEvent{ID: fmt.Sprintf("audit-%d", len(s.auditEvents)+1), EventType: eventType, ActorID: actorID, TargetID: targetID, Metadata: metadata, CreatedAt: s.now()})
}

func membershipKey(workspaceID, userID string) string { return workspaceID + ":" + userID }

func MarshalMetadata(value map[string]string) string {
	payload, _ := json.Marshal(value)
	return string(payload)
}

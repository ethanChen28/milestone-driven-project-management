package identity

import (
	"errors"
	"time"
)

const (
	DefaultWorkspaceID = "default"
	ProviderBuiltIn    = "builtin"
	StatusActive       = "active"
	StatusDisabled     = "disabled"
	MembershipActive   = "active"
	MembershipRemoved  = "removed"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrDisabledAccount    = errors.New("account disabled")
	ErrNotFound           = errors.New("not found")
	ErrForbidden          = errors.New("forbidden")
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	DisplayName  string    `json:"displayName"`
	Email        string    `json:"email"`
	Status       string    `json:"status"`
	Provider     string    `json:"provider"`
	PasswordHash string    `json:"-"`
	TokenVersion int       `json:"version"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Membership struct {
	WorkspaceID string    `json:"workspaceId"`
	UserID      string    `json:"userId"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type DirectoryMember struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"displayName"`
	Email       string   `json:"email"`
	Status      string   `json:"status"`
	Roles       []string `json:"roles"`
}

type ExternalIdentity struct {
	Provider        string    `json:"provider"`
	ExternalSubject string    `json:"externalSubject"`
	UserID          string    `json:"userId"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type AuditEvent struct {
	ID        string            `json:"id"`
	EventType string            `json:"eventType"`
	ActorID   string            `json:"actorId"`
	TargetID  string            `json:"targetId"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"createdAt"`
}

type LoginRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	WorkspaceID string `json:"workspaceId"`
}

type LoginResponse struct {
	AccessToken string          `json:"accessToken"`
	TokenType   string          `json:"tokenType"`
	ExpiresIn   int64           `json:"expiresIn"`
	User        DirectoryMember `json:"user"`
}

type CreateUserRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Role        string `json:"role"`
	WorkspaceID string `json:"workspaceId"`
}

type UpdateUserRequest struct {
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"email,omitempty"`
}

type RoleAssignRequest struct {
	UserID      string `json:"userId"`
	Role        string `json:"role"`
	WorkspaceID string `json:"workspaceId"`
}

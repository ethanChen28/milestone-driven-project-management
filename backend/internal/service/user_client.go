package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type UserProfile struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"displayName"`
	Email       string   `json:"email"`
	Status      string   `json:"status"`
	Roles       []string `json:"roles"`
}

type UserDirectoryClient struct {
	baseURL string
	client  *http.Client
}

func NewUserDirectoryClient(baseURL string) *UserDirectoryClient {
	return &UserDirectoryClient{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{Timeout: 3 * time.Second}}
}

func (c *UserDirectoryClient) ListMembers(ctx context.Context, workspaceID string) ([]UserProfile, error) {
	if c == nil || c.baseURL == "" {
		return DefaultUserProfiles(), nil
	}
	endpoint := c.baseURL + "/api/v1/users"
	if workspaceID != "" {
		endpoint += "?workspaceId=" + url.QueryEscape(workspaceID)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, ErrForbidden
	}
	var users []UserProfile
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}
	return users, nil
}

func DefaultUserProfiles() []UserProfile {
	users := []UserProfile{
		{ID: "admin", Username: "admin", DisplayName: "admin", Email: "admin@example.local", Status: "active", Roles: []string{"admin"}},
		{ID: "tester", Username: "tester", DisplayName: "tester", Email: "tester@example.local", Status: "active", Roles: []string{"admin"}},
		{ID: "alice", Username: "alice", DisplayName: "alice", Email: "alice@example.local", Status: "active", Roles: []string{"project_owner"}},
		{ID: "bob", Username: "bob", DisplayName: "bob", Email: "bob@example.local", Status: "active", Roles: []string{"contributor"}},
		{ID: "carol", Username: "carol", DisplayName: "carol", Email: "carol@example.local", Status: "active", Roles: []string{"viewer"}},
		{ID: "frontend-user", Username: "frontend-user", DisplayName: "frontend-user", Email: "frontend-user@example.local", Status: "active", Roles: []string{"portfolio_manager"}},
		{ID: "leader1", Username: "leader1", DisplayName: "leader1", Email: "leader1@example.local", Status: "active", Roles: []string{"project_owner"}},
		{ID: "leader2", Username: "leader2", DisplayName: "leader2", Email: "leader2@example.local", Status: "active", Roles: []string{"project_owner"}},
		{ID: "eng1", Username: "eng1", DisplayName: "eng1", Email: "eng1@example.local", Status: "active", Roles: []string{"contributor"}},
		{ID: "eng2", Username: "eng2", DisplayName: "eng2", Email: "eng2@example.local", Status: "active", Roles: []string{"contributor"}},
	}
	sort.Slice(users, func(i, j int) bool { return users[i].ID < users[j].ID })
	return users
}

func (c *UserDirectoryClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	if c == nil || c.baseURL == "" {
		return nil, ErrForbidden
	}
	var reqBody io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.client.Do(req)
}

func (c *UserDirectoryClient) CreateUser(ctx context.Context, reqBody interface{}) (UserProfile, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/api/v1/users/manage", reqBody)
	if err != nil {
		return UserProfile{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return UserProfile{}, ErrForbidden
	}
	var user UserProfile
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return UserProfile{}, err
	}
	return user, nil
}

func (c *UserDirectoryClient) UpdateUser(ctx context.Context, id string, reqBody interface{}) (UserProfile, error) {
	path := "/api/v1/users/update?id=" + url.QueryEscape(id)
	resp, err := c.doRequest(ctx, http.MethodPut, path, reqBody)
	if err != nil {
		return UserProfile{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return UserProfile{}, ErrForbidden
	}
	var user UserProfile
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return UserProfile{}, err
	}
	return user, nil
}

func (c *UserDirectoryClient) DisableUser(ctx context.Context, id string) error {
	path := "/api/v1/users/disable?id=" + url.QueryEscape(id)
	resp, err := c.doRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ErrForbidden
	}
	return nil
}

func (c *UserDirectoryClient) EnableUser(ctx context.Context, id string) error {
	path := "/api/v1/users/enable?id=" + url.QueryEscape(id)
	resp, err := c.doRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ErrForbidden
	}
	return nil
}

func (c *UserDirectoryClient) AssignRole(ctx context.Context, reqBody interface{}) error {
	resp, err := c.doRequest(ctx, http.MethodPut, "/api/v1/users/role", reqBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ErrForbidden
	}
	return nil
}

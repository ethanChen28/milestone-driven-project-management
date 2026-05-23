package identity

import "context"

type ProviderUser struct {
	Provider        string
	ExternalSubject string
	User            User
}

type IdentityProvider interface {
	Name() string
	Authenticate(ctx context.Context, username, password string) (ProviderUser, error)
	SyncProfile(ctx context.Context, user User) (User, error)
	ResolveExternalIdentity(ctx context.Context, externalSubject string) (string, error)
	RefreshSession(ctx context.Context, userID string) error
}

type BuiltInProvider struct {
	store *Store
}

func NewBuiltInProvider(store *Store) *BuiltInProvider {
	return &BuiltInProvider{store: store}
}

func (p *BuiltInProvider) Name() string { return ProviderBuiltIn }

func (p *BuiltInProvider) Authenticate(ctx context.Context, username, password string) (ProviderUser, error) {
	user, err := p.store.Authenticate(username, password)
	if err != nil {
		return ProviderUser{}, err
	}
	return ProviderUser{Provider: ProviderBuiltIn, ExternalSubject: user.Username, User: user}, nil
}

func (p *BuiltInProvider) SyncProfile(ctx context.Context, user User) (User, error) {
	return user, nil
}

func (p *BuiltInProvider) ResolveExternalIdentity(ctx context.Context, externalSubject string) (string, error) {
	user, err := p.store.Authenticate(externalSubject, "password")
	if err != nil {
		return "", err
	}
	return user.ID, nil
}

func (p *BuiltInProvider) RefreshSession(ctx context.Context, userID string) error {
	_, err := p.store.DirectoryMember(DefaultWorkspaceID, userID)
	return err
}

type ProviderConfig struct {
	BuiltInEnabled bool   `json:"builtInEnabled"`
	OIDCIssuer     string `json:"oidcIssuer,omitempty"`
	LDAPURL        string `json:"ldapUrl,omitempty"`
}

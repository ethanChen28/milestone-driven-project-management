package identity

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const TokenPrefix = "gm1"

type Claims struct {
	Sub         string   `json:"sub"`
	WorkspaceID string   `json:"workspace_id"`
	Roles       []string `json:"roles"`
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
	Provider    string   `json:"provider"`
	Version     int      `json:"version"`
	Iat         int64    `json:"iat"`
	Exp         int64    `json:"exp"`
}

func IssueToken(secret string, claims Claims) (string, error) {
	if secret == "" {
		return "", errors.New("token secret is required")
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	sig := sign(secret, encodedPayload)
	return TokenPrefix + "." + encodedPayload + "." + sig, nil
}

func ValidateToken(secret, token string, now time.Time) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != TokenPrefix {
		return Claims{}, errors.New("invalid token format")
	}
	if !hmac.Equal([]byte(sign(secret, parts[1])), []byte(parts[2])) {
		return Claims{}, errors.New("invalid token signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, err
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Claims{}, err
	}
	if claims.Exp <= now.Unix() {
		return Claims{}, errors.New("token expired")
	}
	return claims, nil
}

func sign(secret, encodedPayload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encodedPayload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func ClaimsFromMember(user User, member DirectoryMember, workspaceID string, ttl time.Duration, now time.Time) Claims {
	return Claims{Sub: user.ID, WorkspaceID: workspaceID, Roles: member.Roles, DisplayName: user.DisplayName, Email: user.Email, Provider: user.Provider, Version: user.TokenVersion, Iat: now.Unix(), Exp: now.Add(ttl).Unix()}
}

func BearerToken(header string) string {
	if strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return strings.TrimSpace(header[7:])
	}
	return ""
}

func SigningMetadata() map[string]interface{} {
	return map[string]interface{}{
		"keys": []map[string]string{{"kid": "dev-hmac-1", "kty": "oct", "alg": "HS256", "use": "sig"}},
		"note": fmt.Sprintf("%s tokens are HMAC signed; production deployments may replace this with JWT/JWKS or introspection", TokenPrefix),
	}
}

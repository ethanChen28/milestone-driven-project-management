package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const TokenPrefix = "gm1"

type IdentityClaims struct {
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

func IssueIdentityToken(secret string, claims IdentityClaims) (string, error) {
	if secret == "" {
		return "", errors.New("token secret is required")
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	return TokenPrefix + "." + encodedPayload + "." + signToken(secret, encodedPayload), nil
}

func ValidateIdentityToken(secret, token string, now time.Time) (IdentityClaims, error) {
	if secret == "" {
		return IdentityClaims{}, errors.New("token secret is required")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != TokenPrefix {
		return IdentityClaims{}, errors.New("invalid token format")
	}
	if !hmac.Equal([]byte(signToken(secret, parts[1])), []byte(parts[2])) {
		return IdentityClaims{}, errors.New("invalid token signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return IdentityClaims{}, err
	}
	var claims IdentityClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return IdentityClaims{}, err
	}
	if claims.Exp <= now.Unix() {
		return IdentityClaims{}, errors.New("token expired")
	}
	return claims, nil
}

func BearerToken(header string) string {
	if strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return strings.TrimSpace(header[7:])
	}
	return ""
}

func signToken(secret, encodedPayload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encodedPayload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

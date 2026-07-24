package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

const (
	oauthStateCookieName = "si_oauth"
	oauthStateTTL        = 10 * time.Minute
)

type oauthStatePayload struct {
	State     string `json:"state"`
	Verifier  string `json:"verifier"`
	ExpiresAt int64  `json:"exp"`
}

// BeginOAuthState generates a random CSRF state value and a PKCE verifier,
// stashes both in a short-lived signed cookie (there is no server-side store
// for a value that only needs to survive the few seconds between /auth/login
// and /auth/callback), and returns them for building the authorize URL.
func BeginOAuthState(w http.ResponseWriter, secret []byte) (state, verifier string, err error) {
	state, err = randomString(24)
	if err != nil {
		return "", "", err
	}
	verifier = oauth2.GenerateVerifier()

	value, err := signedValue(secret, oauthStatePayload{
		State:     state,
		Verifier:  verifier,
		ExpiresAt: time.Now().Add(oauthStateTTL).Unix(),
	})
	if err != nil {
		return "", "", err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    value,
		Path:     "/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(oauthStateTTL.Seconds()),
	})
	return state, verifier, nil
}

// VerifyOAuthState checks the callback's ?state= against the signed cookie
// (CSRF protection) and returns the matching PKCE verifier. It always clears
// the cookie, whether verification succeeds or fails.
func VerifyOAuthState(w http.ResponseWriter, r *http.Request, secret []byte, callbackState string) (verifier string, err error) {
	defer clearOAuthState(w)

	cookie, err := r.Cookie(oauthStateCookieName)
	if err != nil {
		return "", errors.New("missing oauth state cookie")
	}
	var payload oauthStatePayload
	if err := verifySignedValue(secret, cookie.Value, &payload); err != nil {
		return "", fmt.Errorf("invalid oauth state cookie: %w", err)
	}
	if time.Now().Unix() > payload.ExpiresAt {
		return "", errors.New("oauth state expired")
	}
	if payload.State != callbackState || callbackState == "" {
		return "", errors.New("oauth state mismatch")
	}
	return payload.Verifier, nil
}

func clearOAuthState(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     "/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func randomString(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate random state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

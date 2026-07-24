package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const SessionCookieName = "si_session"

var ErrInvalidSession = errors.New("invalid or expired session")

type sessionPayload struct {
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt int64     `json:"exp"`
}

// IssueSession sets a stateless, HMAC-signed session cookie. There is no
// server-side session table: the cookie itself carries the user id and an
// expiry, authenticated by the signature.
func IssueSession(w http.ResponseWriter, secret []byte, userID uuid.UUID, ttl time.Duration) error {
	value, err := signedValue(secret, sessionPayload{UserID: userID, ExpiresAt: time.Now().Add(ttl).Unix()})
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl.Seconds()),
	})
	return nil
}

// VerifySession validates the session cookie's signature and expiry,
// returning the authenticated user id.
func VerifySession(r *http.Request, secret []byte) (uuid.UUID, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return uuid.Nil, ErrInvalidSession
	}
	var payload sessionPayload
	if err := verifySignedValue(secret, cookie.Value, &payload); err != nil {
		return uuid.Nil, ErrInvalidSession
	}
	if time.Now().Unix() > payload.ExpiresAt {
		return uuid.Nil, ErrInvalidSession
	}
	return payload.UserID, nil
}

// ClearSession expires the session cookie immediately (logout).
func ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// signedValue encodes payload as base64url(json) + "." + hex(HMAC-SHA256).
func signedValue(secret []byte, payload any) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal session payload: %w", err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(encoded))
	sig := hex.EncodeToString(mac.Sum(nil))
	return encoded + "." + sig, nil
}

func verifySignedValue(secret []byte, value string, out any) error {
	dot := strings.IndexByte(value, '.')
	if dot < 0 {
		return errors.New("malformed signed value")
	}
	encoded, sig := value[:dot], value[dot+1:]

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(encoded))
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return errors.New("signature mismatch")
	}

	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("decode signed value: %w", err)
	}
	return json.Unmarshal(raw, out)
}

package widget

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// NewToken creates a self-verifying token: base64url("userID:issuedAtUnix")
// + "." + hex(HMAC-SHA256). It carries everything Verify needs without a DB
// round-trip; the DB is only consulted afterward to check revocation.
func NewToken(secret []byte, userID uuid.UUID) string {
	payload := fmt.Sprintf("%s:%d", userID.String(), time.Now().Unix())
	encoded := base64.RawURLEncoding.EncodeToString([]byte(payload))
	return encoded + "." + sign(secret, encoded)
}

// Verify checks a token's signature locally (no DB lookup) and returns the
// embedded user id if it's valid.
func Verify(secret []byte, token string) (uuid.UUID, bool) {
	dot := strings.IndexByte(token, '.')
	if dot < 0 {
		return uuid.Nil, false
	}
	encoded, sig := token[:dot], token[dot+1:]
	if !hmac.Equal([]byte(sig), []byte(sign(secret, encoded))) {
		return uuid.Nil, false
	}

	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return uuid.Nil, false
	}
	parts := strings.SplitN(string(raw), ":", 2)
	if len(parts) != 2 {
		return uuid.Nil, false
	}
	userID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, false
	}
	if _, err := strconv.ParseInt(parts[1], 10, 64); err != nil {
		return uuid.Nil, false
	}
	return userID, true
}

func sign(secret []byte, encoded string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(encoded))
	return hex.EncodeToString(mac.Sum(nil))
}

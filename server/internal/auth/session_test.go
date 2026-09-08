package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"
	"uuid"
)

func TestTokenEntropyEncodingAndHash(t *testing.T) {
	seen := map[string]bool{}
	for range 100 {
		token := newToken()
		raw, err := base64.RawURLEncoding.Strict().DecodeString(token)
		if err != nil || len(raw) != 32 || len(token) != 43 || seen[token] {
			t.Fatal("token entropy/encoding/uniqueness contract failed")
		}
		seen[token] = true
		got, err := tokenHash(token)
		if err != nil || got != sha256.Sum256([]byte(token)) {
			t.Fatal("session lookup does not hash browser token")
		}
	}
	for _, bad := range []string{"", strings.Repeat("a", 42), strings.Repeat("a", 44), strings.Repeat("+", 43), strings.Repeat("a", 42) + "b"} {
		if _, err := tokenHash(bad); err == nil {
			t.Error("malformed token accepted")
		}
	}
	id := uuid.NewV7()
	if id[6]>>4 != 7 || dbID(id).Bytes != [16]byte(id) {
		t.Fatal("standard-library UUIDv7 persistence conversion failed")
	}
}

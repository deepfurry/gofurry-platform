package auth

import (
	"github.com/gofurry/easyhash"
	"strings"
	"testing"
)

func TestEmailNormalization(t *testing.T) {
	for input, want := range map[string]string{"  Some.One+Tag@EXAMPLE.COM\n": "some.one+tag@example.com", "test@example.com": "test@example.com"} {
		got, err := NormalizeEmail(input)
		if err != nil || got != want {
			t.Error("canonical email normalization failed")
		}
	}
	for _, input := range []string{"Name <a@example.com>", "<a@example.com>", "a@example.com (comment)", "a@example.com,b@example.com", "missing", "a@", "", "a\n@example.com"} {
		if _, err := NormalizeEmail(input); err == nil {
			t.Error("non-plain email accepted")
		}
	}
}

func TestPasswordPolicy(t *testing.T) {
	for _, input := range []string{strings.Repeat(" ", 15), strings.Repeat("界", 15), strings.Repeat("🦊", 128), "  a long password  "} {
		if ValidatePassword(input) != nil {
			t.Error("allowed Unicode/space password rejected")
		}
	}
	for _, input := range []string{"", strings.Repeat("a", 14), strings.Repeat("界", 129), string([]byte{0xff})} {
		if ValidatePassword(input) == nil {
			t.Error("invalid password accepted")
		}
		if _, err := HashPassword(input); err == nil {
			t.Error("invalid password reached hashing")
		}
		if _, _, _, err := verifyPassword(input, "not-a-hash"); err == nil || err == errPasswordOperation {
			t.Error("password validation did not precede KDF")
		}
	}
}

func TestArgon2AndUpgrade(t *testing.T) {
	password := "  unicode password 🦊  "
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatal("hash creation failed")
	}
	algorithm, err := easyhash.Identify(hash)
	if err != nil || algorithm != easyhash.AlgorithmArgon2id {
		t.Fatal("new password is not Argon2id")
	}
	for _, input := range []string{password, strings.TrimSpace(password), "wrong password here"} {
		ok, _, upgraded, err := verifyPassword(input, hash)
		if err != nil || ok != (input == password) || upgraded {
			t.Error("verification/normalization/no-downgrade contract failed")
		}
	}
	composed := "a long password é"
	composedHash, _ := HashPassword(composed)
	if ok, _, _, _ := verifyPassword("a long password e\u0301", composedHash); ok {
		t.Error("password was Unicode-normalized")
	}
	bcrypt, err := easyhash.Hash(password, easyhash.WithBcryptCost(4))
	if err != nil {
		t.Fatal("bcrypt fixture failed")
	}
	ok, replacement, upgraded, err := verifyPassword(password, bcrypt)
	if err != nil || !ok || !upgraded {
		t.Fatal("bcrypt upgrade failed")
	}
	if algorithm, _ := easyhash.Identify(replacement); algorithm != easyhash.AlgorithmArgon2id {
		t.Error("upgrade target is not Argon2id")
	}
	stronger := easyhash.DefaultArgon2()
	stronger.TimeCost++
	strongHash, err := easyhash.Hash(password, easyhash.WithArgon2idConfig(stronger))
	if err != nil {
		t.Fatal("strong hash fixture failed")
	}
	if ok, _, upgraded, err := verifyPassword(password, strongHash); err != nil || !ok || upgraded {
		t.Error("stronger hash was downgraded")
	}
}

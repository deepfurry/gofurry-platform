package mail

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gofurry/easyhash"
)

func TestLocalCapturePrivacyAndConfinement(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "mail")
	l, err := NewLocal(root, dir, "http://localhost:4321")
	if err != nil {
		t.Fatal("capture initialization failed")
	}
	t.Cleanup(func() { _ = l.Close() })
	email := "capture@example.invalid"
	token, err := easyhash.GenerateToken()
	if err != nil {
		t.Fatal("token fixture failed")
	}
	for _, send := range []func() error{
		func() error { return l.SendEmailVerification(t.Context(), email, token) },
		func() error { return l.SendPasswordReset(t.Context(), email, token) },
	} {
		if err := send(); err != nil {
			t.Fatal("capture failed")
		}
	}
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 2 {
		t.Fatal("wrong capture count")
	}
	paths := map[string]bool{}
	for _, file := range files {
		if strings.Contains(file.Name(), email) || strings.Contains(file.Name(), token) {
			t.Fatal("private filename")
		}
		info, _ := file.Info()
		if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
			t.Fatal("capture permissions too broad")
		}
		data, err := os.ReadFile(filepath.Join(dir, file.Name()))
		var msg struct{ To, Subject, Link string }
		if err != nil || json.Unmarshal(data, &msg) != nil || msg.To != email {
			t.Fatal("invalid capture (contents withheld)")
		}
		link, err := url.Parse(msg.Link)
		if err != nil || link.RawQuery != "" || link.Fragment != "token="+token || link.Host != "localhost:4321" {
			t.Fatal("unsafe challenge link")
		}
		paths[link.Path] = true
	}
	if !paths["/verify-email"] || !paths["/reset-password"] {
		t.Fatal("wrong challenge destinations")
	}
	if escaped, err := NewLocal(root, filepath.Join(root, "..", "escaped-mail"), "http://localhost:4321"); err == nil {
		escaped.Close()
		t.Fatal("escaped capture directory accepted")
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err == nil {
		if escaped, err := NewLocal(root, filepath.Join(root, "escape", "mail"), "http://localhost:4321"); err == nil {
			escaped.Close()
			t.Fatal("capture followed escaping symlink")
		}
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if broad, err := NewLocal(root, dir, "http://localhost:4321"); err == nil {
			broad.Close()
			t.Fatal("broad existing capture accepted")
		}
	}
	if (Disabled{}).SendPasswordReset(t.Context(), email, token) == nil {
		t.Fatal("disabled delivery reported success")
	}
}

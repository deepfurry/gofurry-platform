package redisstore

import "testing"

func TestKeysCannotLeaveProjectNamespace(t *testing.T) {
	s := &Store{prefix: "gfp:"}
	for _, suffix := range []string{"smoke:abc", "foreign:key", "", ":key"} {
		if got := s.Key(suffix); got != "gfp:"+suffix {
			t.Fatal("namespace escaped")
		}
	}
	if _, err := Open("redis://localhost:6379", "foreign:"); err == nil {
		t.Fatal("foreign prefix accepted")
	}
}

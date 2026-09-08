package identity

import (
	"strings"
	"testing"
)

func TestProfileValidation(t *testing.T) {
	for _, handle := range []string{"abc", "a_b-c", strings.Repeat("a", 32)} {
		if ValidateHandle(handle) != nil {
			t.Error("valid handle rejected")
		}
	}
	for _, handle := range []string{"", "ab", "Abc", "_abc", "abc.def", "abc\n", "界界界", strings.Repeat("a", 33)} {
		if ValidateHandle(handle) == nil {
			t.Error("invalid handle accepted")
		}
	}
	name, bio := strings.Repeat("🦊", 80), strings.Repeat("界", 500)
	update := ProfileUpdate{DisplayName: Field[string]{Set: true, Value: &name}, Bio: Field[string]{Set: true, Value: &bio}}
	if update.Validate() != nil {
		t.Error("Unicode code points counted incorrectly")
	}
	name += "a"
	if update.Validate() == nil {
		t.Error("oversize display name accepted")
	}
	name = "abc"
	bio += "a"
	if update.Validate() == nil {
		t.Error("oversize bio accepted")
	}
	if (ProfileUpdate{Handle: Field[string]{Set: true}}).Validate() != nil {
		t.Error("nullable handle rejected")
	}
}

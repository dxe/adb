package activists

import (
	"errors"
	"testing"

	"github.com/dxe/adb/pkg/shared"
)

func TestContactMethodLabelsRoundTrip(t *testing.T) {
	seen := map[ContactMethod]string{}
	for _, label := range ValidContactMethods {
		m, err := ParseContactMethod(label)
		if err != nil {
			t.Fatalf("ParseContactMethod(%q) returned error: %v", label, err)
		}
		if m == ContactMethodUnset {
			t.Errorf("ParseContactMethod(%q) = unset, want a stored value", label)
		}
		if dup, ok := seen[m]; ok {
			t.Errorf("stored value %d used by both %q and %q", m, dup, label)
		}
		seen[m] = label
		if got := m.String(); got != label {
			t.Errorf("ContactMethod(%d).String() = %q, want %q", m, got, label)
		}
		if !m.IsValid() {
			t.Errorf("ContactMethod(%d) (%q) is not valid", m, label)
		}
	}
}

func TestParseContactMethod(t *testing.T) {
	for _, in := range []string{"", "   "} {
		m, err := ParseContactMethod(in)
		if err != nil {
			t.Errorf("ParseContactMethod(%q) returned error: %v", in, err)
		}
		if m != ContactMethodUnset {
			t.Errorf("ParseContactMethod(%q) = %d, want unset", in, m)
		}
	}

	if _, err := ParseContactMethod(" Signal "); err != nil {
		t.Errorf("ParseContactMethod with surrounding space returned error: %v", err)
	}

	_, err := ParseContactMethod("Carrier Pigeon")
	if !errors.Is(err, shared.ErrValidation) {
		t.Errorf("ParseContactMethod of an unknown label returned %v, want a validation error", err)
	}
}

func TestContactMethodUnsetAndUnknown(t *testing.T) {
	if got := ContactMethodUnset.String(); got != "" {
		t.Errorf("unset String() = %q, want empty", got)
	}
	if !ContactMethodUnset.IsValid() {
		t.Error("unset should be valid")
	}

	unknown := ContactMethod(200)
	if got := unknown.String(); got != "" {
		t.Errorf("unknown String() = %q, want empty", got)
	}
	if unknown.IsValid() {
		t.Error("unknown stored value should not be valid")
	}
}

func TestContactMethodValueAndScan(t *testing.T) {
	v, err := ContactMethodSignal.Value()
	if err != nil {
		t.Fatalf("Value() returned error: %v", err)
	}
	if v != int64(ContactMethodSignal) {
		t.Errorf("Value() = %v, want %d", v, ContactMethodSignal)
	}

	tests := []struct {
		src  any
		want ContactMethod
	}{
		{int64(2), ContactMethodSMS},
		{[]byte("2"), ContactMethodSMS},
		{"2", ContactMethodSMS},
		{nil, ContactMethodUnset},
	}
	for _, tt := range tests {
		var m ContactMethod
		if err := m.Scan(tt.src); err != nil {
			t.Errorf("Scan(%#v) returned error: %v", tt.src, err)
			continue
		}
		if m != tt.want {
			t.Errorf("Scan(%#v) = %d, want %d", tt.src, m, tt.want)
		}
	}

	var m ContactMethod
	if err := m.Scan("Signal"); err == nil {
		t.Error("Scan of a non-numeric string should fail")
	}
}

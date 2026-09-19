package activists

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"

	"github.com/dxe/adb/pkg/shared"
)

// ContactMethod is a way an activist may ask to be contacted.
type ContactMethod int

const (
	ContactMethodUnset     ContactMethod = 0
	ContactMethodSignal    ContactMethod = 1
	ContactMethodSMS       ContactMethod = 2
	ContactMethodPhoneCall ContactMethod = 3
	ContactMethodEmail     ContactMethod = 4
	ContactMethodWhatsApp  ContactMethod = 5
	ContactMethodTelegram  ContactMethod = 6
	ContactMethodInstagram ContactMethod = 7
	ContactMethodFacebook  ContactMethod = 8
)

// contactMethods pairs each stored value with the label the API uses for it,
// in the order clients should present them. Keep the labels in sync with
// CONTACT_METHODS in the frontend's filter-types.ts.
var contactMethods = []struct {
	Value ContactMethod
	Label string
}{
	{ContactMethodSignal, "Signal"},
	{ContactMethodSMS, "Text / SMS"},
	{ContactMethodPhoneCall, "Phone (Call)"},
	{ContactMethodEmail, "Email"},
	{ContactMethodWhatsApp, "WhatsApp"},
	{ContactMethodTelegram, "Telegram"},
	{ContactMethodInstagram, "Instagram"},
	{ContactMethodFacebook, "Facebook"},
}

// ValidContactMethods are the labels the API accepts, in presentation order.
// The empty string is accepted too and means the activist has no contact
// method recorded.
var ValidContactMethods = func() []string {
	labels := make([]string, 0, len(contactMethods))
	for _, cm := range contactMethods {
		labels = append(labels, cm.Label)
	}
	return labels
}()

// String returns the API label for m, or "" when it is unset. A stored value
// this build does not know (written by a newer one) also reads as "".
func (m ContactMethod) String() string {
	for _, cm := range contactMethods {
		if cm.Value == m {
			return cm.Label
		}
	}
	return ""
}

// IsValid reports whether m is unset or one of the known contact methods.
func (m ContactMethod) IsValid() bool {
	return m == ContactMethodUnset || m.String() != ""
}

// ParseContactMethod converts an API label into its stored value. The empty
// string (after trimming) parses to ContactMethodUnset.
func ParseContactMethod(label string) (ContactMethod, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return ContactMethodUnset, nil
	}
	for _, cm := range contactMethods {
		if cm.Label == label {
			return cm.Value, nil
		}
	}
	return ContactMethodUnset, shared.ValidationErrorf("invalid contact method: %q", label)
}

// Value implements the driver.Valuer interface.
func (m ContactMethod) Value() (driver.Value, error) {
	return int64(m), nil
}

// Scan implements the sql.Scanner interface.
func (m *ContactMethod) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*m = ContactMethodUnset
	case int64:
		*m = ContactMethod(v)
	case []byte:
		return m.scanString(string(v))
	case string:
		return m.scanString(v)
	default:
		return fmt.Errorf("cannot scan %T into ContactMethod", src)
	}
	return nil
}

func (m *ContactMethod) scanString(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("scanning contact method %q: %w", s, err)
	}
	*m = ContactMethod(n)
	return nil
}

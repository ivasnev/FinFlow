package password

import (
	"errors"
	"strings"
	"testing"
)

// TestPolicy проверяет стандартный минимум, настройку, Unicode и ограничение bcrypt.
func TestPolicy(t *testing.T) {
	for _, tc := range []struct {
		name, value string
		minimum     int
		valid       bool
	}{
		{"default short", "1234567", 0, false},
		{"default boundary", "12345678", 0, true},
		{"configured short", "12345678901", 12, false},
		{"configured boundary", "123456789012", 12, true},
		{"unicode short", strings.Repeat("я", 4), 8, false},
		{"unicode boundary", strings.Repeat("я", 8), 8, true},
		{"bcrypt boundary", strings.Repeat("a", 72), 8, true},
		{"bcrypt overflow", strings.Repeat("a", 73), 8, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.value, tc.minimum)
			if tc.valid && err != nil || !tc.valid && !errors.Is(err, ErrInvalid) {
				t.Fatalf("valid=%v err=%v", tc.valid, err)
			}
		})
	}
}

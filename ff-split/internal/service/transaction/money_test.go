package transaction

import (
	"math"
	"testing"
)

// TestDebtCents проверяет точный перевод копеек и отказ при недопустимых суммах.
func TestDebtCents(t *testing.T) {
	for _, tc := range []struct {
		name    string
		amount  float64
		want    int
		invalid bool
	}{
		{"oneCent", 0.01, 1, false}, {"belowHalf", 0.4, 40, false}, {"aboveHalf", 0.6, 60, false}, {"fraction", 1.23, 123, false},
		{"maxDatabaseAmount", 99999999.99, 9999999999, false},
		{"zero", 0, 0, true}, {"negative", -1, 0, true}, {"subCent", 0.001, 0, true},
		{"nan", math.NaN(), 0, true}, {"infinite", math.Inf(1), 0, true}, {"overflow", float64(math.MaxInt), 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := debtCents(tc.amount)
			if (err != nil) != tc.invalid || (!tc.invalid && got != tc.want) {
				t.Fatalf("cents=%d error=%v want=%d invalid=%t", got, err, tc.want, tc.invalid)
			}
		})
	}
}

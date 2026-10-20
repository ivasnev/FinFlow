package debt_calculator

import (
	"testing"

	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// TestPortionParticipants проверяет согласованность users и ключей portion для всех стратегий.
func TestPortionParticipants(t *testing.T) {
	for _, kind := range []string{PercentType, AmountType, UnitsType} {
		for _, tc := range []struct {
			name    string
			users   []int64
			portion map[string]float64
		}{
			{"foreign portion", []int64{1, 2}, map[string]float64{"1": 50, "3": 50}},
			{"duplicate consumer", []int64{1, 1}, map[string]float64{"1": 100}},
			{"empty consumers", nil, map[string]float64{"1": 100}},
			{"nonpositive consumer", []int64{0, 1}, map[string]float64{"0": 50, "1": 50}},
			{"suffix", []int64{1, 2}, map[string]float64{"1": 50, "2evil": 50}},
			{"leading zero", []int64{1, 2}, map[string]float64{"1": 50, "02": 50}},
			{"plus", []int64{1, 2}, map[string]float64{"1": 50, "+2": 50}},
			{"negative id", []int64{1, 2}, map[string]float64{"1": 50, "-2": 50}},
		} {
			t.Run(kind+"/"+tc.name, func(t *testing.T) {
				calculator, _ := GetCalculator(kind)
				if _, _, err := calculator.Calculate(&service.TransactionRequest{Amount: 100, FromUser: 1, Users: tc.users, Portion: tc.portion}, 1); err == nil {
					t.Fatal("invalid participants accepted")
				}
			})
		}
	}
	for _, kind := range []string{PercentType, AmountType} {
		calculator, _ := GetCalculator(kind)
		for _, portion := range []map[string]float64{{"1": 100}, {"1": 50, "-1": 50}} {
			if _, _, err := calculator.Calculate(&service.TransactionRequest{Amount: 100, FromUser: 1, Users: []int64{1, 2}, Portion: portion}, 1); err == nil {
				t.Fatalf("%s accepted missing/sentinel portion", kind)
			}
		}
	}
}

// TestPayerMayPayForOthers разрешает плательщику оплатить расход только других участников.
func TestPayerMayPayForOthers(t *testing.T) {
	for _, kind := range []string{PercentType, AmountType, UnitsType} {
		calculator, _ := GetCalculator(kind)
		shares, debts, err := calculator.Calculate(&service.TransactionRequest{Amount: 100, FromUser: 1, Users: []int64{2}, Portion: map[string]float64{"2": 100}}, 1)
		if err != nil || len(shares) != 1 || len(debts) != 1 || debts[0].Amount != 100 {
			t.Fatalf("%s: shares=%+v debts=%+v err=%v", kind, shares, debts, err)
		}
	}
}

package debt_calculator

import (
	"errors"
	"math"
	"reflect"
	"testing"

	commonerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// TestInvalidDistribution проверяет некорректные суммы и веса всех стратегий.
func TestInvalidDistribution(t *testing.T) {
	for _, strategy := range []string{PercentType, AmountType, UnitsType} {
		for _, tc := range []struct {
			name    string
			amount  float64
			portion map[string]float64
		}{
			{"negative share", 100, map[string]float64{"1": -50, "2": 150}},
			{"nan share", 100, map[string]float64{"1": math.NaN(), "2": 100}},
			{"infinite share", 100, map[string]float64{"1": math.Inf(1), "2": 100}},
			{"zero shares", 100, map[string]float64{"1": 0, "2": 0}},
			{"negative amount", -100, map[string]float64{"1": 50, "2": 50}},
			{"zero amount", 0, map[string]float64{"1": 50, "2": 50}},
			{"nan amount", math.NaN(), map[string]float64{"1": 50, "2": 50}},
			{"infinite amount", math.Inf(1), map[string]float64{"1": 50, "2": 50}},
			{"fractional cent", 100.001, map[string]float64{"1": 50, "2": 50}},
			{"overflow amount", 100000000, map[string]float64{"1": 50, "2": 50}},
		} {
			t.Run(strategy+"/"+tc.name, func(t *testing.T) {
				calculator, _ := GetCalculator(strategy)
				_, _, err := calculator.Calculate(&service.TransactionRequest{Amount: tc.amount, FromUser: 1, Users: []int64{1, 2}, Portion: tc.portion}, 1)
				var validation *commonerrors.ValidationError
				if !errors.As(err, &validation) {
					t.Fatalf("want validation error, got %v", err)
				}
			})
		}
	}
}

// TestRoundingConservesMoney проверяет копейки, порядок и долги при дробном распределении.
func TestRoundingConservesMoney(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeName string
		amount   float64
		portion  map[string]float64
		want     []Share
	}{
		{"thirds percent", PercentType, 100, map[string]float64{"1": 33.33, "2": 33.33, "3": 33.33}, []Share{{1, 33.34}, {2, 33.33}, {3, 33.33}}},
		{"thirds units", UnitsType, 100, nil, []Share{{1, 33.34}, {2, 33.33}, {3, 33.33}}},
		{"one cent", UnitsType, 0.01, nil, []Share{{1, 0.01}, {2, 0}, {3, 0}}},
		{"small weights", UnitsType, 1, map[string]float64{"1": math.SmallestNonzeroFloat64, "2": math.SmallestNonzeroFloat64, "3": math.SmallestNonzeroFloat64}, []Share{{1, 0.34}, {2, 0.33}, {3, 0.33}}},
		{"explicit zero remainder", UnitsType, 1, map[string]float64{"1": 1, "-1": 0}, []Share{{1, 1}, {2, 0}, {3, 0}}},
		{"fixed cents", AmountType, 1, map[string]float64{"1": 0.34, "2": 0.33, "3": 0.33}, []Share{{1, 0.34}, {2, 0.33}, {3, 0.33}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calculator, _ := GetCalculator(tc.typeName)
			for i := 0; i < 20; i++ {
				shares, debts, err := calculator.Calculate(&service.TransactionRequest{Amount: tc.amount, FromUser: 1, Users: []int64{3, 1, 2}, Portion: tc.portion}, 1)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(shares, tc.want) {
					t.Fatalf("got %+v want %+v", shares, tc.want)
				}
				var sum int64
				for _, share := range shares {
					sum += int64(math.Round(share.Value * 100))
				}
				if sum != int64(math.Round(tc.amount*100)) {
					t.Fatalf("lost cents: %d", sum)
				}
				var debtSum int64
				for _, debt := range debts {
					debtSum += int64(math.Round(debt.Amount * 100))
				}
				if debtSum != sum-int64(math.Round(shares[0].Value*100)) {
					t.Fatalf("incorrect debts: %+v", debts)
				}
			}
		})
	}
}

// TestFixedAmountsRequireExactCents исключает скрытое округление и расхождение на копейку.
func TestFixedAmountsRequireExactCents(t *testing.T) {
	for _, portion := range []map[string]float64{{"1": .333, "2": .667}, {"1": .50, "2": .49}, {"1": .50, "2": .51}} {
		if _, _, err := (&AmountStrategy{}).Calculate(&service.TransactionRequest{Amount: 1, Portion: portion}, 1); err == nil {
			t.Fatalf("accepted %+v", portion)
		}
	}
}

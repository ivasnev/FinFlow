package tests

import (
	"context"
	"testing"

	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// TestPersonalDebtsWithoutLoadedProfiles проверяет персональные ответы на настоящем ID-only маппинге PostgreSQL.
func TestPersonalDebtsWithoutLoadedProfiles(t *testing.T) {
	_, c := setupOptimizationTest(t)
	ctx := context.Background()
	if _, err := c.TransactionService.CreateTransaction(ctx, 1, optimizationRequest(100)); err != nil {
		t.Fatal(err)
	}
	// Сценарий: репозиторий возвращает существующий долг 2→1, но не заполняет объекты профилей.
	raw, err := c.TransactionRepository.GetDebtsByEventIDFromUser(1, 2)
	if err != nil || len(raw) != 1 {
		t.Fatalf("stored debts: %+v, %v", raw, err)
	}
	if raw[0].FromUser != nil || raw[0].ToUser != nil {
		t.Fatal("regression requires actual ID-only mapping")
	}
	assertRequestor := func(t *testing.T, user *service.DebtsUserResponse, amount, wantAmount float64, wantID int64) {
		t.Helper()
		if user == nil || user.ID != wantID || user.Name != "Incognito" || user.Photo != "" || amount != wantAmount {
			t.Fatalf("opponent=%+v, amount=%.2f; want ID=%d, amount=%.2f", user, amount, wantID, wantAmount)
		}
	}
	for _, tc := range []struct {
		name     string
		get      func() ([]service.DebtDTO, error)
		amount   float64
		opponent int64
	}{
		{"outgoing", func() ([]service.DebtDTO, error) { return c.TransactionService.GetDebtsByEventIDFromUser(1, 2) }, -50, 1},
		{"incoming", func() ([]service.DebtDTO, error) { return c.TransactionService.GetDebtsByEventIDToUser(1, 1) }, 50, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			debts, err := tc.get()
			if err != nil || len(debts) != 1 {
				t.Fatalf("personal debts: %+v, %v", debts, err)
			}
			assertRequestor(t, debts[0].Requestor, debts[0].Amount, tc.amount, tc.opponent)
		})
	}
	// Сценарии: первый оптимизированный ответ вычисляется, а последующие читают ID-only кэш.
	for _, tc := range []struct {
		name     string
		get      func() ([]service.OptimizedDebtDTO, error)
		amount   float64
		opponent int64
	}{
		{"optimized outgoing", func() ([]service.OptimizedDebtDTO, error) {
			return c.TransactionService.GetOptimizedDebtsByEventIDFromUser(ctx, 1, 2)
		}, -50, 1},
		{"optimized incoming", func() ([]service.OptimizedDebtDTO, error) {
			return c.TransactionService.GetOptimizedDebtsByEventIDToUser(ctx, 1, 1)
		}, 50, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < 2; i++ {
				debts, err := tc.get()
				if err != nil || len(debts) != 1 {
					t.Fatalf("personal optimized debts: %+v, %v", debts, err)
				}
				assertRequestor(t, debts[0].Requestor, debts[0].Amount, tc.amount, tc.opponent)
			}
		})
	}
}

package validator

import (
	"testing"

	"github.com/ivasnev/FinFlow/ff-common/optimizers"
)

func TestValidateSuccess(t *testing.T) {
	original := []optimizers.Transfer{
		{From: "A", To: "B", Amount: 10},
	}
	optimized := []optimizers.Transfer{
		{From: "A", To: "B", Amount: 10},
	}

	v := NewValidator(WithAllChecks())
	report := v.Validate(original, optimized)
	if !report.Valid {
		t.Fatalf("expected valid report, got violations: %+v", report.Violations)
	}
}

func TestValidateBalanceViolation(t *testing.T) {
	original := []optimizers.Transfer{
		{From: "A", To: "B", Amount: 10},
	}
	optimized := []optimizers.Transfer{
		{From: "A", To: "B", Amount: 5},
	}

	v := NewValidator(WithBalancesCheck())
	report := v.Validate(original, optimized)
	if !hasCriterion(report, CriterionBalances) {
		t.Fatalf("expected balance violation")
	}
}

func TestValidateNoNewDebtsViolation(t *testing.T) {
	original := []optimizers.Transfer{
		{From: "A", To: "B", Amount: 10},
	}
	optimized := []optimizers.Transfer{
		{From: "A", To: "C", Amount: 10},
	}

	v := NewValidator(WithNoNewDebtsCheck())
	report := v.Validate(original, optimized)
	if !hasCriterion(report, CriterionNoNewDebts) {
		t.Fatalf("expected no-new-debts violation")
	}
}

func TestValidateTotalNotMoreViolation(t *testing.T) {
	original := []optimizers.Transfer{
		{From: "A", To: "B", Amount: 10},
		{From: "A", To: "C", Amount: 10},
	}
	optimized := []optimizers.Transfer{
		{From: "A", To: "B", Amount: 15},
		{From: "A", To: "C", Amount: 10},
	}

	v := NewValidator(WithTotalNotMoreCheck())
	report := v.Validate(original, optimized)
	if !hasCriterion(report, CriterionTotalNotMore) {
		t.Fatalf("expected total-not-more violation")
	}
}

func TestValidatorWithSelectiveChecks(t *testing.T) {
	original := []optimizers.Transfer{
		{From: "A", To: "B", Amount: 10},
	}
	optimized := []optimizers.Transfer{
		{From: "A", To: "B", Amount: 5},
	}

	// Только проверка на новые долги - должна пройти
	v := NewValidator(WithNoNewDebtsCheck())
	report := v.Validate(original, optimized)
	if !report.Valid {
		t.Fatalf("expected valid with only no-new-debts check")
	}

	// Проверка балансов - должна упасть
	v = NewValidator(WithBalancesCheck())
	report = v.Validate(original, optimized)
	if report.Valid {
		t.Fatalf("expected invalid with balances check")
	}
}

func TestValidatorWithAllChecks(t *testing.T) {
	original := []optimizers.Transfer{
		{From: "A", To: "B", Amount: 10},
	}
	optimized := []optimizers.Transfer{
		{From: "A", To: "B", Amount: 10},
	}

	v := NewValidator(WithAllChecks())
	report := v.Validate(original, optimized)
	if !report.Valid {
		t.Fatalf("expected valid with all checks: %+v", report.Violations)
	}
}

func TestValidatorEmptyConfig(t *testing.T) {
	original := []optimizers.Transfer{
		{From: "A", To: "B", Amount: 10},
	}
	optimized := []optimizers.Transfer{
		{From: "A", To: "C", Amount: 100},
	}

	v := NewValidator()
	report := v.Validate(original, optimized)
	if !report.Valid {
		t.Fatalf("expected valid with no checks enabled")
	}
}

func hasCriterion(report Report, criterion string) bool {
	for _, v := range report.Violations {
		if v.Criterion == criterion {
			return true
		}
	}
	return false
}

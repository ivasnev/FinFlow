package pushrelabel

import (
	"testing"

	"github.com/ivasnev/FinFlow/ff-common/optimizers/tests/testutil"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/utils/validator"
)

func TestPushRelabelDirectSuccess(t *testing.T) {
	opt := New()
	v := validator.NewValidator(validator.WithAllChecks())
	debts := testutil.DebtsDirectSimple()
	result, err := opt.Optimize(debts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	report := v.Validate(debts, result)
	if !report.Valid {
		t.Fatalf("expected valid result: %+v", report.Violations)
	}
}

func TestPushRelabelIntermediateSuccess(t *testing.T) {
	opt := New()
	v := validator.NewValidator(validator.WithAllChecks())
	debts := testutil.DebtsNeedsIntermediate()
	result, err := opt.Optimize(debts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	report := v.Validate(debts, result)
	if !report.Valid {
		t.Fatalf("expected valid result: %+v", report.Violations)
	}
}

func TestPushRelabelComplexGraph(t *testing.T) {
	opt := New()
	v := validator.NewValidator(validator.WithAllChecks())
	debts := testutil.DebtsComplexGraph()
	result, err := opt.Optimize(debts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	report := v.Validate(debts, result)
	if !report.Valid {
		t.Fatalf("expected valid result: %+v", report.Violations)
	}
}

func TestPushRelabelEmpty(t *testing.T) {
	opt := New()
	result, err := opt.Optimize(testutil.DebtsEmpty())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestPushRelabelTriangle(t *testing.T) {
	opt := New()
	result, err := opt.Optimize(testutil.DebtsTriangle())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result for balanced triangle, got %v", result)
	}
}

package optimizers_test

import (
	"fmt"
	"testing"

	"github.com/ivasnev/FinFlow/ff-common/optimizers"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/dinic"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/edmonds_karp"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/greedy"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/testutil"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/validator"
)

func assertOptimizerValid(t *testing.T, debts []optimizers.Transfer, opt optimizers.Optimizer, v *validator.Validator) {
	t.Helper()
	result, err := opt.Optimize(debts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	report := v.Validate(debts, result)
	fmt.Printf("result: %+v\n", result)
	if !report.Valid {
		t.Fatalf("invalid result: %+v", report.Violations)
	}
}

func TestGreedyOptimizer(t *testing.T) {
	opt := greedy.New()
	v := validator.NewValidator(validator.WithBalancesCheck())
	t.Run("direct_simple", func(t *testing.T) {
		assertOptimizerValid(t, testutil.DebtsDirectSimple(), opt, v)
	})
	t.Run("needs_intermediate", func(t *testing.T) {
		assertOptimizerValid(t, testutil.DebtsNeedsIntermediate(), opt, v)
	})
	t.Run("complex_graph", func(t *testing.T) {
		assertOptimizerValid(t, testutil.DebtsComplexGraph(), opt, v)
	})
}

func TestDinicOptimizer(t *testing.T) {
	opt := dinic.New()
	v := validator.NewValidator(validator.WithAllChecks())
	t.Run("direct_simple", func(t *testing.T) {
		assertOptimizerValid(t, testutil.DebtsDirectSimple(), opt, v)
	})
	t.Run("needs_intermediate", func(t *testing.T) {
		assertOptimizerValid(t, testutil.DebtsNeedsIntermediate(), opt, v)
	})
	t.Run("complex_graph", func(t *testing.T) {
		assertOptimizerValid(t, testutil.DebtsComplexGraph(), opt, v)
	})
}

func TestEdmondsKarpOptimizer(t *testing.T) {
	opt := edmonds_karp.New()
	v := validator.NewValidator(validator.WithAllChecks())
	t.Run("direct_simple", func(t *testing.T) {
		assertOptimizerValid(t, testutil.DebtsDirectSimple(), opt, v)
	})
	t.Run("needs_intermediate", func(t *testing.T) {
		assertOptimizerValid(t, testutil.DebtsNeedsIntermediate(), opt, v)
	})
	t.Run("complex_graph", func(t *testing.T) {
		assertOptimizerValid(t, testutil.DebtsComplexGraph(), opt, v)
	})
}

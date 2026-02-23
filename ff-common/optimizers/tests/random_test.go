package optimizers_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/ivasnev/FinFlow/ff-common/optimizers"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/dinic"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/edmonds_karp"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/greedy"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/validator"
)

// RandomTransfers генерирует случайный набор переводов: n узлов (имена "0".."n-1"),
// m транзакций — каждый раз выбираются два случайных узла и случайная сумма.
// seed задаёт воспроизводимость. from и to всегда различны, amount от 1 до maxAmount (если 0 — 1000).
func RandomTransfers(nodes, transactions int, seed int64, maxAmount int) []optimizers.Transfer {
	if nodes < 2 || transactions <= 0 {
		return nil
	}
	if maxAmount <= 0 {
		maxAmount = 1000
	}

	names := make([]string, nodes)
	for i := 0; i < nodes; i++ {
		names[i] = fmt.Sprintf("%d", i)
	}

	rng := rand.New(rand.NewSource(seed))
	out := make([]optimizers.Transfer, 0, transactions)

	for i := 0; i < transactions; i++ {
		from := names[rng.Intn(nodes)]
		to := names[rng.Intn(nodes)]
		for to == from {
			to = names[rng.Intn(nodes)]
		}
		amount := rng.Intn(maxAmount) + 1
		out = append(out, optimizers.Transfer{From: from, To: to, Amount: amount})
	}

	return out
}

type benchmarkMetrics struct {
	reductions  int
	valid       int
	sumLenRaw   int64
	sumLenInput int64
	iterations  int
}

func runOptimizerBenchmark(b *testing.B, opt optimizers.Optimizer, v *validator.Validator, nodes, trans int, seed int64) benchmarkMetrics {
	var m benchmarkMetrics
	for i := 0; i < b.N; i++ {
		raw := RandomTransfers(nodes, trans, seed+int64(i), 500)
		input, err := optimizers.CollapseTransfers(raw)
		if err != nil || len(input) == 0 {
			continue
		}
		m.sumLenRaw += int64(len(raw))
		m.sumLenInput += int64(len(input))
		m.iterations++

		result, err := opt.Optimize(input)
		if err != nil {
			continue
		}
		report := v.Validate(input, result)
		if report.Valid {
			m.valid++
			if len(result) < len(input) {
				m.reductions++
			}
		}
	}
	return m
}

func reportBenchmarkMetrics(b *testing.B, m benchmarkMetrics) {
	if m.iterations > 0 {
		b.ReportMetric(float64(m.sumLenInput)/float64(m.iterations), "input_after_collapse")
		if m.sumLenRaw > 0 {
			b.ReportMetric((1-float64(m.sumLenInput)/float64(m.sumLenRaw))*100, "collapse_pct")
		}
	}
	b.ReportMetric(float64(m.reductions)/float64(b.N)*100, "reduction_pct")
	b.ReportMetric(float64(m.valid)/float64(b.N)*100, "valid_pct")
}

func BenchmarkRandom_Greedy(b *testing.B) {
	opt := greedy.New()
	v := validator.NewValidator(validator.WithBalancesCheck())
	const nodes, trans = 20, 4000
	seed := int64(12345)
	m := runOptimizerBenchmark(b, opt, v, nodes, trans, seed)
	reportBenchmarkMetrics(b, m)
}

func BenchmarkRandom_Dinic(b *testing.B) {
	opt := dinic.New()
	v := validator.NewValidator(validator.WithAllChecks())
	const nodes, trans = 20, 4000
	seed := int64(12345)
	m := runOptimizerBenchmark(b, opt, v, nodes, trans, seed)
	reportBenchmarkMetrics(b, m)
}

func BenchmarkRandom_EdmondsKarp(b *testing.B) {
	opt := edmonds_karp.New()
	v := validator.NewValidator(validator.WithAllChecks())
	const nodes, trans = 20, 4000
	seed := int64(12345)
	m := runOptimizerBenchmark(b, opt, v, nodes, trans, seed)
	reportBenchmarkMetrics(b, m)
}

// transValuesForSweep возвращает слайс размеров транзакций для sweep-бенчмарка:
// 10..100 с шагом 10, затем 1000..10000 с шагом 1000.
func transValuesForSweep() []int {
	var out []int
	for t := 10; t <= 100; t += 10 {
		out = append(out, t)
	}
	for t := 1000; t <= 10000; t += 1000 {
		out = append(out, t)
	}
	return out
}

func runSweepBenchmark(b *testing.B, opt optimizers.Optimizer, v *validator.Validator, seed int64) {
	transValues := transValuesForSweep()
	for nodes := 5; nodes <= 200; nodes += 15 {
		for _, trans := range transValues {
			if trans > nodes*(nodes-1) {
				continue
			}
			nodes, trans := nodes, trans
			name := fmt.Sprintf("nodes=%d_trans=%d", nodes, trans)
			b.Run(name, func(b *testing.B) {
				m := runOptimizerBenchmark(b, opt, v, nodes, trans, seed)
				reportBenchmarkMetrics(b, m)
			})
		}
	}
}

func BenchmarkRandom_Sweep_Greedy(b *testing.B) {
	opt := greedy.New()
	v := validator.NewValidator(validator.WithBalancesCheck())
	runSweepBenchmark(b, opt, v, 12345)
}

func BenchmarkRandom_Sweep_Dinic(b *testing.B) {
	opt := dinic.New()
	v := validator.NewValidator(validator.WithAllChecks())
	runSweepBenchmark(b, opt, v, 12345)
}

func BenchmarkRandom_Sweep_EdmondsKarp(b *testing.B) {
	opt := edmonds_karp.New()
	v := validator.NewValidator(validator.WithAllChecks())
	runSweepBenchmark(b, opt, v, 12345)
}

func TestRandomTransfers(t *testing.T) {
	out := RandomTransfers(5, 20, 1, 100)
	if len(out) != 20 {
		t.Fatalf("expected 20 transfers, got %d", len(out))
	}
	for i, tr := range out {
		if tr.From == tr.To || tr.Amount <= 0 {
			t.Fatalf("invalid transfer %d: %+v", i, tr)
		}
	}
}

func TestRandomTransfers_CollapseThenOptimize(t *testing.T) {
	raw := RandomTransfers(8, 30, 42, 200)
	input, err := optimizers.CollapseTransfers(raw)
	if err != nil {
		t.Fatalf("collapse: %v", err)
	}

	opt := dinic.New()
	v := validator.NewValidator(validator.WithAllChecks())
	result, err := opt.Optimize(input)
	if err != nil {
		t.Fatalf("optimize: %v", err)
	}
	report := v.Validate(input, result)
	if !report.Valid {
		t.Fatalf("validation failed: %+v", report.Violations)
	}
	if len(result) > len(input) {
		t.Logf("transfers: input %d -> output %d", len(input), len(result))
	}
}

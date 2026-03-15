package optimizers_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/ivasnev/FinFlow/ff-common/optimizers"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/dinic"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/edmonds_karp"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/greedy"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/interactive_maxflow"
	mfd "github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow/dinic"
	mfek "github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow/edmondskarp"
	mfpr "github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow/pushrelabel"
	optpr "github.com/ivasnev/FinFlow/ff-common/optimizers/pushrelabel"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/utils"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/utils/validator"
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
	reductions       int   // среди валидных: число случаев, где len(result) < len(input)
	valid            int   // count of valid results (Validate(...).Valid == true)
	collapseOK       int   // раз удалось получить input после collapse (len>0, err==nil)
	optimizeOK       int   // раз Optimize вернул без ошибки
	sumLenRaw        int64
	sumLenInput      int64
	sumInputLenValid int64 // сумма len(input) по валидным итерациям (для reduction_len_pct)
	sumSaved         int64 // сумма (len(input)-len(result)) по валидным; может быть < 0
	sumCollapseNs    int64 // суммарное время collapse по всем b.N итерациям
	sumOptimizeNs    int64 // суммарное время optimize по всем вызовам (collapseOK раз)
}

func runOptimizerBenchmark(b *testing.B, opt optimizers.Optimizer, v *validator.Validator, nodes, trans int, seed int64) benchmarkMetrics {
	var m benchmarkMetrics
	for i := 0; i < b.N; i++ {
		raw := RandomTransfers(nodes, trans, seed+int64(i), 500)
		startCollapse := time.Now()
		input, err := utils.CollapseTransfers(raw)
		m.sumCollapseNs += time.Since(startCollapse).Nanoseconds()
		if err != nil || len(input) == 0 {
			continue
		}
		m.collapseOK++
		m.sumLenRaw += int64(len(raw))
		m.sumLenInput += int64(len(input))

		startOptimize := time.Now()
		result, err := opt.Optimize(input)
		m.sumOptimizeNs += time.Since(startOptimize).Nanoseconds()
		if err != nil {
			continue
		}
		m.optimizeOK++
		report := v.Validate(input, result)
		if report.Valid {
			m.valid++
			m.sumInputLenValid += int64(len(input))
			saved := int64(len(input) - len(result)) // честно: может быть < 0 (ухудшение)
			m.sumSaved += saved
			if len(result) < len(input) {
				m.reductions++
			}
		}
	}
	return m
}

func reportBenchmarkMetrics(b *testing.B, m benchmarkMetrics) {
	var inputAfter float64
	var collapsePct float64
	if m.collapseOK > 0 {
		inputAfter = float64(m.sumLenInput) / float64(m.collapseOK)
		if m.sumLenRaw > 0 {
			collapsePct = (1 - float64(m.sumLenInput)/float64(m.sumLenRaw)) * 100
		}
	}
	b.ReportMetric(inputAfter, "input_after_collapse")
	b.ReportMetric(collapsePct, "collapse_pct")
	// collapse меряем на всех b.N итерациях; optimize — только при успешном collapse (collapseOK раз)
	collapseNs := float64(m.sumCollapseNs) / float64(b.N)
	optimizeNs := float64(0)
	if m.collapseOK > 0 {
		optimizeNs = float64(m.sumOptimizeNs) / float64(m.collapseOK)
	}
	b.ReportMetric(collapseNs, "collapse_ns_per_op")
	b.ReportMetric(optimizeNs, "optimize_ns_per_op")
	reduceSuccess := float64(0)
	if m.valid > 0 {
		reduceSuccess = float64(m.reductions) / float64(m.valid) * 100
	}
	b.ReportMetric(reduceSuccess, "reduce_success_pct")
	reductionLen := float64(0)
	if m.sumInputLenValid > 0 {
		reductionLen = float64(m.sumSaved) / float64(m.sumInputLenValid) * 100
	}
	b.ReportMetric(reductionLen, "reduction_len_pct")
	validPct := float64(0)
	if m.optimizeOK > 0 {
		validPct = float64(m.valid) / float64(m.optimizeOK) * 100
	}
	b.ReportMetric(validPct, "valid_pct")
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

func BenchmarkRandom_PushRelabel(b *testing.B) {
	opt := optpr.New()
	v := validator.NewValidator(validator.WithAllChecks())
	const nodes, trans = 20, 4000
	seed := int64(12345)
	m := runOptimizerBenchmark(b, opt, v, nodes, trans, seed)
	reportBenchmarkMetrics(b, m)
}

func BenchmarkRandom_InteractiveMaxflow_Dinic(b *testing.B) {
	opt := interactive_maxflow.New(mfd.Solver{})
	v := validator.NewValidator(validator.WithAllChecks())
	const nodes, trans = 20, 4000
	seed := int64(12345)
	m := runOptimizerBenchmark(b, opt, v, nodes, trans, seed)
	reportBenchmarkMetrics(b, m)
}

func BenchmarkRandom_InteractiveMaxflow_EdmondsKarp(b *testing.B) {
	opt := interactive_maxflow.New(mfek.Solver{})
	v := validator.NewValidator(validator.WithAllChecks())
	const nodes, trans = 20, 4000
	seed := int64(12345)
	m := runOptimizerBenchmark(b, opt, v, nodes, trans, seed)
	reportBenchmarkMetrics(b, m)
}

func BenchmarkRandom_InteractiveMaxflow_PushRelabel(b *testing.B) {
	opt := interactive_maxflow.New(mfpr.Solver{})
	v := validator.NewValidator(validator.WithAllChecks())
	const nodes, trans = 20, 4000
	seed := int64(12345)
	m := runOptimizerBenchmark(b, opt, v, nodes, trans, seed)
	reportBenchmarkMetrics(b, m)
}

// nodesForSweep — набор узлов для sweep-бенчмарка.
var nodesForSweep = []int{5, 10, 25, 50, 100, 200}

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
	for _, nodes := range nodesForSweep {
		for _, trans := range transValues {
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

func BenchmarkRandom_Sweep_PushRelabel(b *testing.B) {
	opt := optpr.New()
	v := validator.NewValidator(validator.WithAllChecks())
	runSweepBenchmark(b, opt, v, 12345)
}

func BenchmarkRandom_Sweep_InteractiveMaxflow_Dinic(b *testing.B) {
	opt := interactive_maxflow.New(mfd.Solver{})
	v := validator.NewValidator(validator.WithAllChecks())
	runSweepBenchmark(b, opt, v, 12345)
}

func BenchmarkRandom_Sweep_InteractiveMaxflow_EdmondsKarp(b *testing.B) {
	opt := interactive_maxflow.New(mfek.Solver{})
	v := validator.NewValidator(validator.WithAllChecks())
	runSweepBenchmark(b, opt, v, 12345)
}

func BenchmarkRandom_Sweep_InteractiveMaxflow_PushRelabel(b *testing.B) {
	opt := interactive_maxflow.New(mfpr.Solver{})
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
	input, err := utils.CollapseTransfers(raw)
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

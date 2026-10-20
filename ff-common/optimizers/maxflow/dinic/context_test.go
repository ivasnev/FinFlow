package dinic

import (
	"context"
	"errors"
	"testing"

	"github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow"
)

// traversalContext детерминированно отменяет контекст после начала обхода графа.
type traversalContext struct {
	context.Context
	cancel context.CancelFunc
	checks int
}

// Err моделирует отмену во время вычисления без таймеров и ожиданий планировщика.
func (c *traversalContext) Err() error {
	c.checks++
	if c.checks > 4 {
		c.cancel()
	}
	return c.Context.Err()
}

// TestSolverContextCancellation проверяет отмену до вычисления и во время обхода.
func TestSolverContextCancellation(t *testing.T) {
	for _, during := range []bool{false, true} {
		name := "before"
		if during {
			name = "during"
		}
		t.Run(name, func(t *testing.T) {
			base, cancel := context.WithCancel(context.Background())
			defer cancel()
			var ctx context.Context = base
			if during {
				ctx = &traversalContext{Context: base, cancel: cancel}
			} else {
				cancel()
			}
			graph := maxflow.NewGraph(100)
			for i := 0; i < 99; i++ {
				graph.AddEdge(i, i+1, 10)
			}
			if _, err := (Solver{}).GetMaxFlowContext(ctx, graph, 0, 99); !errors.Is(err, context.Canceled) {
				t.Fatalf("cancellation ignored: %v", err)
			}
		})
	}
}

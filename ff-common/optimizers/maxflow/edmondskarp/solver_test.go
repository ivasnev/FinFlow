package edmondskarp

import (
	"testing"

	"github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow"
)

func TestSolver_GetMaxFlow(t *testing.T) {
	g := maxflow.NewGraph(3)
	g.AddEdge(0, 1, 10)
	g.AddEdge(1, 2, 10)
	flow := Solver{}.GetMaxFlow(g, 0, 2)
	if flow != 10 {
		t.Errorf("GetMaxFlow = %d, want 10", flow)
	}
}

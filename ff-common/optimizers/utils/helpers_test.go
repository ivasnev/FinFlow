package utils

import (
	"testing"

	"github.com/ivasnev/FinFlow/ff-common/optimizers"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow/dinic"
	"github.com/stretchr/testify/require"
)

func TestCollapseTransfers(t *testing.T) {
	testcases := []struct {
		name               string
		transfers          []optimizers.Transfer
		optimizedTransfers []optimizers.Transfer
	}{
		{
			name: "reverse transfer",
			transfers: []optimizers.Transfer{
				{From: "a", To: "b", Amount: 10},
				{From: "a", To: "b", Amount: 20},
				{From: "b", To: "a", Amount: 40},
			},
			optimizedTransfers: []optimizers.Transfer{
				{From: "b", To: "a", Amount: 10},
			},
		},
		{
			name: "zero net",
			transfers: []optimizers.Transfer{
				{From: "a", To: "b", Amount: 10},
				{From: "a", To: "b", Amount: 20},
				{From: "b", To: "a", Amount: 30},
			},
			optimizedTransfers: nil,
		},
		{
			name: "simple",
			transfers: []optimizers.Transfer{
				{From: "a", To: "b", Amount: 10},
				{From: "a", To: "b", Amount: 20},
				{From: "b", To: "a", Amount: 20},
			},
			optimizedTransfers: []optimizers.Transfer{
				{From: "a", To: "b", Amount: 10},
			},
		},
		{
			name:               "empty",
			transfers:          nil,
			optimizedTransfers: nil,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CollapseTransfers(tc.transfers)
			require.NoError(t, err)
			require.Equal(t, tc.optimizedTransfers, got)
		})
	}
}

func TestResidualGraph(t *testing.T) {
	g := maxflow.NewGraph(3)
	g.AddEdge(0, 1, 10)
	g.AddEdge(1, 2, 10)
	_ = dinic.Solver{}.GetMaxFlow(g, 0, 2)
	res := ResidualGraph(g)
	if res.NumNodes() != 3 {
		t.Errorf("ResidualGraph NumNodes = %d, want 3", res.NumNodes())
	}
	out1 := res.OutEdges(1)
	var back10 bool
	for _, e := range out1 {
		if e.To == 0 && e.Cap == 10 {
			back10 = true
			break
		}
	}
	if !back10 {
		t.Errorf("residual: expected edge 1->0 with cap 10, got OutEdges(1)=%v", out1)
	}
}

func TestTransfersFromFlowGraph(t *testing.T) {
	// n=2, вершины 0,1 in, 2,3 out. Поток 5 по 2->1.
	g := maxflow.NewGraph(4)
	g.AddEdge(0, 2, 5)
	g.AddEdge(2, 1, 5)
	g.AddEdge(1, 3, 5)
	for _, e := range g.Adj[2] {
		if e.To == 1 {
			e.Flow = 5
			break
		}
	}
	users := []string{"A", "B"}
	got := TransfersFromFlowGraph(g, 2, users)
	if len(got) != 1 {
		t.Fatalf("got %d transfers, want 1", len(got))
	}
	if got[0].From != "A" || got[0].To != "B" || got[0].Amount != 5 {
		t.Errorf("got %+v, want From=A To=B Amount=5", got[0])
	}
}

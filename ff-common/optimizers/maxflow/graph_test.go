package maxflow

import (
	"testing"
)

func TestGraph_AddEdge_OutEdges_NumNodes(t *testing.T) {
	g := NewGraph(3)
	g.AddEdge(0, 1, 10)
	g.AddEdge(1, 2, 10)
	if g.NumNodes() != 3 {
		t.Errorf("NumNodes = %d, want 3", g.NumNodes())
	}
	out := g.OutEdges(1)
	if len(out) != 2 {
		t.Fatalf("OutEdges(1) len = %d", len(out))
	}
	for _, e := range out {
		if e.To == 2 {
			if e.Cap != 10 || e.Flow != 0 {
				t.Errorf("edge 1->2: Cap=%d Flow=%d, want Cap=10 Flow=0", e.Cap, e.Flow)
			}
		}
	}
}

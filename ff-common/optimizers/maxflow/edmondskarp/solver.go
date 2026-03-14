package edmondskarp

import (
	"math"

	"github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow"
)

// Solver реализует алгоритм Эдмондса-Карпа (BFS-версия Форда-Фалкерсона).
// Сложность O(V·E²).
type Solver struct{}

// GetMaxFlow находит максимальный поток от source до sink, мутирует граф.
func (Solver) GetMaxFlow(g *maxflow.Graph, source, sink int) int {
	flow := 0
	for {
		parent := bfsKarp(g, source, sink)
		if parent == nil {
			break
		}
		pathFlow := math.MaxInt32
		for v := sink; v != source; {
			e := parent[v]
			if e.Cap-e.Flow < pathFlow {
				pathFlow = e.Cap - e.Flow
			}
			v = g.Adj[e.To][e.Rev].To
		}
		for v := sink; v != source; {
			e := parent[v]
			e.Flow += pathFlow
			g.Adj[e.To][e.Rev].Flow -= pathFlow
			v = g.Adj[e.To][e.Rev].To
		}
		flow += pathFlow
	}
	return flow
}

func bfsKarp(g *maxflow.Graph, s, t int) map[int]*maxflow.Edge {
	parent := make(map[int]*maxflow.Edge)
	visited := make([]bool, g.NumNodes())
	visited[s] = true
	queue := []int{s}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		for _, e := range g.Adj[v] {
			if !visited[e.To] && e.Cap-e.Flow > 0 {
				visited[e.To] = true
				parent[e.To] = e
				if e.To == t {
					return parent
				}
				queue = append(queue, e.To)
			}
		}
	}
	return nil
}

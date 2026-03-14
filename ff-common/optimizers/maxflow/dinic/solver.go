package dinic

import (
	"math"

	"github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow"
)

// Solver реализует алгоритм Диница для максимального потока.
// Сложность O(V²·E).
type Solver struct{}

// GetMaxFlow находит максимальный поток от source до sink, мутирует граф.
func (Solver) GetMaxFlow(g *maxflow.Graph, source, sink int) int {
	n := g.NumNodes()
	level := make([]int, n)
	iter := make([]int, n)
	flow := 0
	for bfsDinic(g, source, sink, level) {
		for i := range iter {
			iter[i] = 0
		}
		for {
			f := dfsDinic(g, source, sink, math.MaxInt32, level, iter)
			if f == 0 {
				break
			}
			flow += f
		}
	}
	return flow
}

func bfsDinic(g *maxflow.Graph, s, t int, level []int) bool {
	for i := range level {
		level[i] = -1
	}
	level[s] = 0
	queue := []int{s}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		for _, e := range g.Adj[v] {
			if e.Cap-e.Flow > 0 && level[e.To] < 0 {
				level[e.To] = level[v] + 1
				queue = append(queue, e.To)
			}
		}
	}
	return level[t] >= 0
}

func dfsDinic(g *maxflow.Graph, v, t, f int, level, iter []int) int {
	if v == t {
		return f
	}
	for ; iter[v] < len(g.Adj[v]); iter[v]++ {
		e := g.Adj[v][iter[v]]
		if e.Cap-e.Flow <= 0 || level[v]+1 != level[e.To] {
			continue
		}
		d := dfsDinic(g, e.To, t, minInt(f, e.Cap-e.Flow), level, iter)
		if d > 0 {
			e.Flow += d
			g.Adj[e.To][e.Rev].Flow -= d
			return d
		}
	}
	return 0
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

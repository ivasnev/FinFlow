package dinic

import (
	"context"
	"math"

	"github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow"
)

// Solver реализует алгоритм Диница для максимального потока.
// Сложность O(V²·E).
type Solver struct{}

// GetMaxFlow находит максимальный поток от source до sink, мутирует граф.
func (s Solver) GetMaxFlow(g *maxflow.Graph, source, sink int) int {
	flow, _ := s.GetMaxFlowContext(context.Background(), g, source, sink)
	return flow
}

// GetMaxFlowContext проверяет отмену при обходе рёбер; при ошибке граф может быть частично изменён.
func (Solver) GetMaxFlowContext(ctx context.Context, g *maxflow.Graph, source, sink int) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	n := g.NumNodes()
	level := make([]int, n)
	iter := make([]int, n)
	flow := 0
	for {
		reachable, err := bfsDinic(ctx, g, source, sink, level)
		if err != nil {
			return flow, err
		}
		if !reachable {
			return flow, nil
		}
		for i := range iter {
			iter[i] = 0
		}
		for {
			f, err := dfsDinic(ctx, g, source, sink, math.MaxInt, level, iter)
			if err != nil {
				return flow, err
			}
			if f == 0 {
				break
			}
			flow += f
		}
	}
}

// bfsDinic строит слои остаточной сети с проверкой отмены.
func bfsDinic(ctx context.Context, g *maxflow.Graph, s, t int, level []int) (bool, error) {
	for i := range level {
		level[i] = -1
	}
	level[s] = 0
	queue := []int{s}
	for len(queue) > 0 {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		v := queue[0]
		queue = queue[1:]
		for _, e := range g.Adj[v] {
			if err := ctx.Err(); err != nil {
				return false, err
			}
			if e.Cap-e.Flow > 0 && level[e.To] < 0 {
				level[e.To] = level[v] + 1
				queue = append(queue, e.To)
			}
		}
	}
	return level[t] >= 0, nil
}

// dfsDinic отправляет поток по слоистой сети, прекращая обход при отмене.
func dfsDinic(ctx context.Context, g *maxflow.Graph, v, t, f int, level, iter []int) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if v == t {
		return f, nil
	}
	for ; iter[v] < len(g.Adj[v]); iter[v]++ {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		e := g.Adj[v][iter[v]]
		if e.Cap-e.Flow <= 0 || level[v]+1 != level[e.To] {
			continue
		}
		d, err := dfsDinic(ctx, g, e.To, t, minInt(f, e.Cap-e.Flow), level, iter)
		if err != nil {
			return 0, err
		}
		if d > 0 {
			e.Flow += d
			g.Adj[e.To][e.Rev].Flow -= d
			return d, nil
		}
	}
	return 0, nil
}

// minInt выбирает меньшую допустимую величину потока.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

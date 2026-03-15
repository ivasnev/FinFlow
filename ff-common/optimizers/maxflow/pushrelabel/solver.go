package pushrelabel

import (
	"fmt"
	"math"

	"github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow"
)

// maxIterMultiplier — верхняя граница итераций: O(n³) для защиты от бесконечного цикла.
// При превышении вызывается panic.
const maxIterMultiplier = 4

// Solver реализует алгоритм Черкасского (Cheriyan–Maheshwari) — push-relabel
// с выбором вершины с максимальной высотой (highest label).
// Сложность O(V·E + V²√E).
type Solver struct{}

// GetMaxFlow находит максимальный поток от source до sink, мутирует граф.
func (Solver) GetMaxFlow(g *maxflow.Graph, source, sink int) int {
	n := g.NumNodes()
	if n == 0 {
		return 0
	}
	const inf = math.MaxInt

	height := make([]int, n)
	height[source] = n
	excess := make([]int, n)
	excess[source] = inf

	// Начальное проталкивание из source во все соседи (как в эталоне: excess[s]=inf, затем push(s,i))
	for _, e := range g.Adj[source] {
		if e.Cap <= 0 {
			continue
		}
		d := min(excess[source], e.Cap-e.Flow)
		if d <= 0 {
			continue
		}
		e.Flow += d
		g.Adj[e.To][e.Rev].Flow -= d
		excess[source] -= d
		excess[e.To] += d
	}

	maxIter := maxIterMultiplier * n * n * (n + 1) // O(n³), защита от бесконечного цикла
	iter := 0
	current := make([]int, 0, n) // переиспользуемый буфер, чтобы не аллоцировать на каждой итерации
	for {
		if iter >= maxIter {
			panic(fmt.Sprintf("pushrelabel: превышен лимит итераций %d (n=%d); возможна ошибка в графе или в алгоритме", maxIter, n))
		}
		iter++

		current = findMaxHeightVerticesReuse(current, excess, height, source, sink)
		if len(current) == 0 {
			break
		}
		done := false
		for _, u := range current {
			if excess[u] <= 0 {
				continue
			}
			pushed := false
			for _, e := range g.Adj[u] {
				if excess[u] <= 0 {
					break
				}
				res := e.Cap - e.Flow
				if res > 0 && height[u] == height[e.To]+1 {
					d := min(excess[u], res)
					e.Flow += d
					g.Adj[e.To][e.Rev].Flow -= d
					excess[u] -= d
					excess[e.To] += d
					pushed = true
				}
			}
			if !pushed {
				relabel(g, u, height, inf)
				done = true
				break
			}
		}
		if done {
			continue
		}
	}

	return excess[sink]
}

// findMaxHeightVerticesReuse заполняет buf вершинами с максимальной высотой и положительным excess; buf переиспользуется.
func findMaxHeightVerticesReuse(buf []int, excess, height []int, source, sink int) []int {
	buf = buf[:0]
	maxH := -1
	for i := range excess {
		if i == source || i == sink || excess[i] <= 0 {
			continue
		}
		h := height[i]
		if h > maxH {
			maxH = h
			buf = buf[:0]
		}
		if h == maxH {
			buf = append(buf, i)
		}
	}
	return buf
}

func relabel(g *maxflow.Graph, u int, height []int, inf int) {
	d := inf
	for _, e := range g.Adj[u] {
		if e.Cap-e.Flow > 0 && height[e.To] < d {
			d = height[e.To]
		}
	}
	if d < inf {
		height[u] = d + 1
	} else {
		// Нет исходящих рёбер с остаточной способностью — поднимаем выше source (n+1), чтобы поток мог вернуться к source
		height[u] = len(height) + 1
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

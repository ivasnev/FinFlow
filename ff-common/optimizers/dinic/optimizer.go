package dinic

import (
	"fmt"
	"math"

	"github.com/ivasnev/FinFlow/ff-common/optimizers"
)

// Optimizer реализует алгоритм Диница для оптимизации долгов.
// Сложность: O(V^2 * E), где V - количество пользователей, E - количество долговых связей.
type Optimizer struct{}

// New создаёт новый экземпляр оптимизатора на основе алгоритма Диница.
func New() *Optimizer {
	return &Optimizer{}
}

func (o *Optimizer) Optimize(debts []optimizers.Transfer) ([]optimizers.Transfer, error) {
	if len(debts) == 0 {
		return nil, nil
	}

	balances, err := optimizers.Balances(debts)
	if err != nil {
		return nil, err
	}

	users := optimizers.Users(debts)
	index := make(map[string]int, len(users))
	for i, u := range users {
		index[u] = i
	}

	// Вычисляем ограничения на суммарный исходящий поток для каждого пользователя
	originalOut := make(map[string]int, len(users))
	for _, tr := range debts {
		if tr.Amount > 0 {
			originalOut[tr.From] += tr.Amount
		}
	}

	totalDemand := 0
	for _, user := range users {
		if balances[user] > 0 {
			totalDemand += balances[user]
		}
	}

	if totalDemand == 0 {
		return nil, nil
	}

	// Строим граф с разбиением вершин (node splitting):
	// [0..n-1]     - u_in
	// [n..2n-1]    - u_out
	// source = 2n
	// sink   = 2n+1
	n := len(users)
	source := 2 * n
	sink := 2*n + 1
	graph := newGraph(2*n + 2)

	// Ограничение суммарного исходящего потока для каждого пользователя (node capacity)
	for i, user := range users {
		in := i
		out := i + n
		cap := originalOut[user]
		if cap > 0 {
			graph.addEdge(in, out, cap)
		}
	}

	// Источник/сток по балансам
	for i, user := range users {
		in := i
		balance := balances[user]
		if balance < 0 {
			graph.addEdge(source, in, -balance)
		} else if balance > 0 {
			graph.addEdge(in, sink, balance)
		}
	}

	// Разрешённые связи (from->to): u_out -> v_in
	allowed := optimizers.TransferMatrix(debts)
	infCap := totalDemand
	for from, toMap := range allowed {
		fromIdx := index[from]
		fromOut := fromIdx + n
		for to := range toMap {
			toIdx := index[to]
			if fromIdx == toIdx {
				continue
			}
			toIn := toIdx
			graph.addEdge(fromOut, toIn, infCap)
		}
	}

	maxFlow := graph.maxFlow(source, sink)
	if maxFlow != totalDemand {
		return nil, fmt.Errorf("insufficient max flow: %d of %d", maxFlow, totalDemand)
	}

	// Считываем только переводы по разрешённым ребрам: u_out -> v_in
	var transfers []optimizers.Transfer
	for fromIdx := 0; fromIdx < n; fromIdx++ {
		fromOut := fromIdx + n
		fromUser := users[fromIdx]
		for _, e := range graph.adj[fromOut] {
			if e.to < 0 || e.to >= n {
				continue
			}
			if e.flow > 0 {
				transfers = append(transfers, optimizers.Transfer{
					From:   fromUser,
					To:     users[e.to],
					Amount: e.flow,
				})
			}
		}
	}

	return transfers, nil
}

type edge struct {
	to   int
	rev  int
	cap  int
	flow int
}

type graph struct {
	adj   [][]*edge
	level []int
	iter  []int
}

func newGraph(n int) *graph {
	return &graph{
		adj:   make([][]*edge, n),
		level: make([]int, n),
		iter:  make([]int, n),
	}
}

func (g *graph) addEdge(from, to, cap int) {
	fwd := &edge{to: to, rev: len(g.adj[to]), cap: cap}
	rev := &edge{to: from, rev: len(g.adj[from]), cap: 0}
	g.adj[from] = append(g.adj[from], fwd)
	g.adj[to] = append(g.adj[to], rev)
}

func (g *graph) maxFlow(s, t int) int {
	flow := 0
	for g.bfs(s, t) {
		for i := range g.iter {
			g.iter[i] = 0
		}
		for {
			f := g.dfs(s, t, math.MaxInt32)
			if f == 0 {
				break
			}
			flow += f
		}
	}
	return flow
}

func (g *graph) bfs(s, t int) bool {
	for i := range g.level {
		g.level[i] = -1
	}
	g.level[s] = 0
	queue := []int{s}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		for _, e := range g.adj[v] {
			if e.cap-e.flow > 0 && g.level[e.to] < 0 {
				g.level[e.to] = g.level[v] + 1
				queue = append(queue, e.to)
			}
		}
	}
	return g.level[t] >= 0
}

func (g *graph) dfs(v, t, f int) int {
	if v == t {
		return f
	}
	for ; g.iter[v] < len(g.adj[v]); g.iter[v]++ {
		e := g.adj[v][g.iter[v]]
		if e.cap-e.flow <= 0 || g.level[v]+1 != g.level[e.to] {
			continue
		}
		d := g.dfs(e.to, t, minInt(f, e.cap-e.flow))
		if d > 0 {
			e.flow += d
			g.adj[e.to][e.rev].flow -= d
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

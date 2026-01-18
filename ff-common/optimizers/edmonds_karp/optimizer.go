package edmonds_karp

import (
	"fmt"
	"math"

	"github.com/ivasnev/FinFlow/ff-common/optimizers"
)

// Optimizer реализует алгоритм Эдмондса-Карпа для оптимизации долгов.
// Это BFS-версия алгоритма Форда-Фалкерсона.
// Сложность: O(V * E^2), где V - количество пользователей, E - количество долговых связей.
type Optimizer struct{}

// New создаёт новый экземпляр оптимизатора на основе алгоритма Эдмондса-Карпа.
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
	g := newGraph(2*n + 2)

	// Ограничение суммарного исходящего потока для каждого пользователя (node capacity)
	for i, user := range users {
		in := i
		out := i + n
		cap := originalOut[user]
		if cap > 0 {
			g.addEdge(in, out, cap)
		}
	}

	// Источник/сток по балансам
	for i, user := range users {
		in := i
		balance := balances[user]
		if balance < 0 {
			g.addEdge(source, in, -balance)
		} else if balance > 0 {
			g.addEdge(in, sink, balance)
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
			g.addEdge(fromOut, toIn, infCap)
		}
	}

	maxFlow := g.maxFlow(source, sink)
	if maxFlow != totalDemand {
		return nil, fmt.Errorf("insufficient max flow: %d of %d", maxFlow, totalDemand)
	}

	// Считываем только переводы по разрешённым ребрам: u_out -> v_in
	var transfers []optimizers.Transfer
	for fromIdx := 0; fromIdx < n; fromIdx++ {
		fromOut := fromIdx + n
		fromUser := users[fromIdx]
		for _, e := range g.adj[fromOut] {
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
	adj [][]*edge
}

func newGraph(n int) *graph {
	return &graph{
		adj: make([][]*edge, n),
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
	for {
		parent := g.bfs(s, t)
		if parent == nil {
			break
		}

		// Находим минимальную пропускную способность на пути
		pathFlow := math.MaxInt32
		for v := t; v != s; {
			e := parent[v]
			if e.cap-e.flow < pathFlow {
				pathFlow = e.cap - e.flow
			}
			v = g.adj[e.to][e.rev].to
		}

		// Обновляем поток на пути
		for v := t; v != s; {
			e := parent[v]
			e.flow += pathFlow
			g.adj[e.to][e.rev].flow -= pathFlow
			v = g.adj[e.to][e.rev].to
		}

		flow += pathFlow
	}
	return flow
}

// bfs ищет кратчайший увеличивающий путь от s до t.
// Возвращает карту parent, где parent[v] - ребро, ведущее в v.
func (g *graph) bfs(s, t int) map[int]*edge {
	parent := make(map[int]*edge)
	visited := make([]bool, len(g.adj))
	visited[s] = true
	queue := []int{s}

	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]

		for _, e := range g.adj[v] {
			if !visited[e.to] && e.cap-e.flow > 0 {
				visited[e.to] = true
				parent[e.to] = e
				if e.to == t {
					return parent
				}
				queue = append(queue, e.to)
			}
		}
	}

	return nil
}

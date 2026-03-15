package interactive_maxflow

import (
	"github.com/ivasnev/FinFlow/ff-common/optimizers"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/maxflow"
)

type Optimizer struct {
	Solver maxflow.MaxFlowSolver
}

func New(solver maxflow.MaxFlowSolver) *Optimizer {
	return &Optimizer{Solver: solver}
}

func edgeHash(from, to int) int64 {
	return int64(from)*1_000_000_000 + int64(to)
}

type simpleEdge struct {
	from, to, cap int
}

func (o *Optimizer) Optimize(debts []optimizers.Transfer) ([]optimizers.Transfer, error) {
	if len(debts) == 0 {
		return nil, nil
	}

	// ——— Шаг 1: собрать участников ———
	nameToID := make(map[string]int)
	idToName := make([]string, 0)
	getID := func(name string) int {
		if id, ok := nameToID[name]; ok {
			return id
		}
		id := len(idToName)
		nameToID[name] = id
		idToName = append(idToName, name)
		return id
	}
	for _, d := range debts {
		getID(d.From)
		getID(d.To)
	}
	n := len(idToName)

	// ——— Шаг 2: начальные рёбра ———
	var edges []simpleEdge
	for _, d := range debts {
		edges = append(edges, simpleEdge{
			from: getID(d.From),
			to:   getID(d.To),
			cap:  d.Amount,
		})
	}

	// ——— Шаг 3: итеративная оптимизация ———
	visited := make(map[int64]bool)

	for {
		// Найти непосещённое ребро
		edgeIdx := -1
		for i, e := range edges {
			if !visited[edgeHash(e.from, e.to)] {
				edgeIdx = i
			}
		}
		if edgeIdx < 0 {
			break
		}

		source := edges[edgeIdx].from
		sink := edges[edgeIdx].to

		// Построить flow-граф и запомнить какие рёбра прямые
		g := maxflow.NewGraph(n)

		// Запоминаем индексы ПРЯМЫХ рёбер: (nodeIdx, edgeIdxInAdj)
		type fwdRecord struct {
			node int
			idx  int
			from int
			to   int
		}
		var fwdEdges []fwdRecord

		for _, e := range edges {
			idx := g.AddEdge(e.from, e.to, e.cap)
			fwdEdges = append(fwdEdges, fwdRecord{
				node: e.from,
				idx:  idx,
				from: e.from,
				to:   e.to,
			})
		}

		// Max-flow
		mf := o.Solver.GetMaxFlow(g, source, sink)

		// Пометить как посещённое
		visited[edgeHash(source, sink)] = true

		// Собрать новые рёбра ТОЛЬКО из прямых рёбер (remaining capacity)
		var newEdges []simpleEdge
		for _, rec := range fwdEdges {
			e := g.Adj[rec.node][rec.idx]
			remaining := e.Cap - e.Flow
			if remaining > 0 {
				newEdges = append(newEdges, simpleEdge{
					from: rec.from,
					to:   rec.to,
					cap:  remaining,
				})
			}
		}

		// Добавить ребро source→sink с весом maxFlow
		if mf > 0 {
			newEdges = append(newEdges, simpleEdge{
				from: source,
				to:   sink,
				cap:  mf,
			})
		}

		edges = newEdges
	}

	// ——— Шаг 4: собрать результат ———
	// Схлопнуть дубли одинаковых пар
	type pairKey struct{ from, to int }
	combined := make(map[pairKey]int)
	for _, e := range edges {
		combined[pairKey{e.from, e.to}] += e.cap
	}

	var result []optimizers.Transfer
	for key, amount := range combined {
		if amount > 0 {
			result = append(result, optimizers.Transfer{
				From:   idToName[key.from],
				To:     idToName[key.to],
				Amount: amount,
			})
		}
	}

	return result, nil
}
